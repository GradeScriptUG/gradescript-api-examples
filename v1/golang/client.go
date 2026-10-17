// A small client for the GradeScript Grading API: https://console.gradescriptug.com/docs
// Go 1.21 or later, standard library only.
package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

const defaultAPIURL = "https://api.gradescriptug.com/client/v1"

// GETs are safe to repeat, so they are retried after these delays on network
// errors, 429 and 5xx. POST and PATCH are never retried: creating a grading
// twice would grade, and bill, the script twice.
var getRetryDelays = []time.Duration{1 * time.Second, 2 * time.Second, 4 * time.Second}

type Client struct {
	APIURL     string
	APIKey     string
	HTTPClient *http.Client
}

// NewClientFromEnv reads GRADESCRIPT_API_KEY and, optionally, GRADESCRIPT_API_URL.
func NewClientFromEnv() (*Client, error) {
	apiKey := os.Getenv("GRADESCRIPT_API_KEY")
	if apiKey == "" {
		return nil, errors.New("set GRADESCRIPT_API_KEY in the .env file at the repo root (see .env.example)")
	}
	apiURL := os.Getenv("GRADESCRIPT_API_URL")
	if apiURL == "" {
		apiURL = defaultAPIURL
	}
	return &Client{
		APIURL:     strings.TrimRight(apiURL, "/"),
		APIKey:     apiKey,
		HTTPClient: &http.Client{Timeout: 5 * time.Minute},
	}, nil
}

type Grading struct {
	ID              string       `json:"id"`
	Status          string       `json:"status"`
	Pipeline        string       `json:"pipeline"`
	APIKeyID        *string      `json:"apiKeyID"`
	StudentFullname string       `json:"studentFullname"`
	StudentNumber   string       `json:"studentNumber"`
	Subject         string       `json:"subject"`
	Level           string       `json:"level"`
	AssessmentType  string       `json:"assessmentType"`
	Instructions    string       `json:"instructions"`
	Title           string       `json:"title"`
	Description     string       `json:"description"`
	Feedback        string       `json:"feedback"`
	Error           string       `json:"error"`
	TotalMark       *float64     `json:"totalMark"`
	MaximumMark     *float64     `json:"maximumMark"`
	Questions       []Question   `json:"questions"`
	Annotations     []Annotation `json:"annotations"`
	Attachments     []Attachment `json:"attachments"`
	CreatedAt       time.Time    `json:"createdAt"`
	UpdatedAt       time.Time    `json:"updatedAt"`
}

// Finished reports whether the grading has stopped running.
func (g *Grading) Finished() bool {
	switch g.Status {
	case "COMPLETED", "PARTIAL", "FAILED":
		return true
	}
	return false
}

type Question struct {
	Ref          string  `json:"ref"`
	QuestionText string  `json:"questionText"`
	Topic        string  `json:"topic"`
	Awarded      float64 `json:"awarded"`
	Max          float64 `json:"max"`
	Rationale    string  `json:"rationale"`
	Attempted    bool    `json:"attempted"`
	Graded       bool    `json:"graded"`
	NeedsReview  bool    `json:"needsReview"`
	ReviewReason string  `json:"reviewReason"`
}

// Annotation marks a region of a script page. X, Y, Width and Height are
// fractions of the page (0 to 1), measured from the top-left corner.
type Annotation struct {
	AttachmentID string  `json:"attachmentID"`
	Page         int     `json:"page"`
	Type         string  `json:"type"`
	X            float64 `json:"x"`
	Y            float64 `json:"y"`
	Width        float64 `json:"width"`
	Height       float64 `json:"height"`
	QuestionRef  string  `json:"questionRef"`
	Comment      string  `json:"comment"`
	MarkAwarded  float64 `json:"markAwarded"`
	MaximumMark  float64 `json:"maximumMark"`
	NeedsReview  bool    `json:"needsReview"`
	ReviewReason string  `json:"reviewReason"`
}

type Attachment struct {
	ID               string `json:"id"`
	Type             string `json:"type"`
	Index            int    `json:"index"`
	OriginalFilename string `json:"originalFilename"`
	ContentType      string `json:"contentType"`
	Size             int64  `json:"size"`
}

type Pagination struct {
	Limit        int    `json:"limit"`
	NextCursor   string `json:"nextCursor"`
	HasNextItems bool   `json:"hasNextItems"`
	Count        int    `json:"count"`
}

// APIError is an error response from the API: {status, error, message}.
type APIError struct {
	StatusCode int
	Name       string
	Message    string
}

func (e *APIError) Error() string {
	return fmt.Sprintf("HTTP %d %s: %s", e.StatusCode, e.Name, e.Message)
}

type envelope struct {
	Error      string          `json:"error"`
	Message    string          `json:"message"`
	Data       json.RawMessage `json:"data"`
	Pagination *Pagination     `json:"pagination"`
}

func (c *Client) do(method, path string, body []byte, contentType string) (*envelope, error) {
	var delays []time.Duration
	if method == http.MethodGet {
		delays = getRetryDelays
	}

	for attempt := 0; ; attempt++ {
		var reader io.Reader
		if body != nil {
			reader = bytes.NewReader(body)
		}
		request, err := http.NewRequest(method, c.APIURL+path, reader)
		if err != nil {
			return nil, err
		}
		request.Header.Set("Authorization", "Bearer "+c.APIKey)
		if contentType != "" {
			request.Header.Set("Content-Type", contentType)
		}

		response, err := c.HTTPClient.Do(request)
		retryable := err != nil || response.StatusCode == http.StatusTooManyRequests || response.StatusCode >= 500
		if retryable && attempt < len(delays) {
			if response != nil {
				response.Body.Close()
			}
			time.Sleep(delays[attempt])
			continue
		}
		if err != nil {
			return nil, err
		}
		defer response.Body.Close()

		var result envelope
		if err := json.NewDecoder(response.Body).Decode(&result); err != nil && response.StatusCode < 400 {
			return nil, fmt.Errorf("decode response: %w", err)
		}
		if response.StatusCode >= 400 {
			if result.Error == "" {
				result.Error = http.StatusText(response.StatusCode)
			}
			return nil, &APIError{StatusCode: response.StatusCode, Name: result.Error, Message: result.Message}
		}
		return &result, nil
	}
}

func (c *Client) doGrading(method, path string, body []byte, contentType string) (*Grading, error) {
	result, err := c.do(method, path, body, contentType)
	if err != nil {
		return nil, err
	}
	var grading Grading
	if err := json.Unmarshal(result.Data, &grading); err != nil {
		return nil, fmt.Errorf("decode grading: %w", err)
	}
	return &grading, nil
}

type CreateGradingInput struct {
	StudentFullname string // required
	Subject         string // required
	StudentNumber   string
	Level           string
	AssessmentType  string
	Instructions    string
	Attachments     []AttachmentInput // at least one SCRIPT
}

type AttachmentInput struct {
	Type  string // SCRIPT, MARKING_GUIDE, QUESTION_PAPER or OTHER
	Path  string // a PDF, PNG, JPEG or WEBP file of at most 20 MB
	Index int    // optional, from 1: orders files of the same type
}

// CreateGrading uploads a script for grading and returns the queued grading.
func (c *Client) CreateGrading(input CreateGradingInput) (*Grading, error) {
	var body bytes.Buffer
	form := multipart.NewWriter(&body)

	// Send only the fields that are set: the API rejects unknown fields.
	fields := [][2]string{
		{"studentFullname", input.StudentFullname},
		{"subject", input.Subject},
		{"studentNumber", input.StudentNumber},
		{"level", input.Level},
		{"assessmentType", input.AssessmentType},
		{"instructions", input.Instructions},
	}
	for _, field := range fields {
		if field[1] != "" {
			form.WriteField(field[0], field[1])
		}
	}

	for i, attachment := range input.Attachments {
		form.WriteField(fmt.Sprintf("attachments[%d][type]", i), attachment.Type)
		if attachment.Index > 0 {
			form.WriteField(fmt.Sprintf("attachments[%d][index]", i), strconv.Itoa(attachment.Index))
		}
		if err := attachFile(form, fmt.Sprintf("attachments[%d][file]", i), attachment.Path); err != nil {
			return nil, err
		}
	}
	if err := form.Close(); err != nil {
		return nil, err
	}

	return c.doGrading(http.MethodPost, "/grading", body.Bytes(), form.FormDataContentType())
}

func attachFile(form *multipart.Writer, field, path string) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	// The API detects the file type from its contents, so the part's
	// application/octet-stream Content-Type is fine.
	part, err := form.CreateFormFile(field, filepath.Base(path))
	if err != nil {
		return err
	}
	_, err = io.Copy(part, file)
	return err
}

type ListGradingsInput struct {
	Status string // optional, comma-separated, e.g. "COMPLETED,PARTIAL"
	Limit  int    // optional, 1 to 25 (default 12)
	Cursor string // optional, NextCursor from the previous page
}

// ListGradings returns one page of gradings, newest first.
func (c *Client) ListGradings(input ListGradingsInput) ([]Grading, *Pagination, error) {
	query := url.Values{}
	if input.Status != "" {
		query.Set("status", input.Status)
	}
	if input.Limit > 0 {
		query.Set("limit", strconv.Itoa(input.Limit))
	}
	if input.Cursor != "" {
		query.Set("cursor", input.Cursor)
	}
	path := "/grading"
	if len(query) > 0 {
		path += "?" + query.Encode()
	}

	result, err := c.do(http.MethodGet, path, nil, "")
	if err != nil {
		return nil, nil, err
	}
	var gradings []Grading
	if err := json.Unmarshal(result.Data, &gradings); err != nil {
		return nil, nil, fmt.Errorf("decode gradings: %w", err)
	}
	return gradings, result.Pagination, nil
}

func (c *Client) GetGrading(id string) (*Grading, error) {
	return c.doGrading(http.MethodGet, "/grading/"+url.PathEscape(id), nil, "")
}

// RetryGrading queues a FAILED or PARTIAL grading again.
func (c *Client) RetryGrading(id string) (*Grading, error) {
	return c.doGrading(http.MethodPatch, "/grading/"+url.PathEscape(id)+"/retry", nil, "")
}

// DeleteGrading deletes a finished grading and returns its id.
func (c *Client) DeleteGrading(id string) (string, error) {
	result, err := c.do(http.MethodDelete, "/grading/"+url.PathEscape(id), nil, "")
	if err != nil {
		return "", err
	}
	var deleted struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(result.Data, &deleted); err != nil {
		return "", fmt.Errorf("decode response: %w", err)
	}
	return deleted.ID, nil
}

// WaitForGrading polls a grading every interval until it finishes, calling
// onStatus whenever the status changes. There are no webhooks.
func (c *Client) WaitForGrading(id string, interval, timeout time.Duration, onStatus func(string)) (*Grading, error) {
	deadline := time.Now().Add(timeout)
	lastStatus := ""

	for {
		grading, err := c.GetGrading(id)
		if err != nil {
			return nil, err
		}
		if grading.Status != lastStatus {
			onStatus(grading.Status)
			lastStatus = grading.Status
		}
		if grading.Finished() {
			return grading, nil
		}

		if time.Now().After(deadline) {
			return nil, fmt.Errorf("grading %s did not finish within %s", id, timeout)
		}
		time.Sleep(interval)
	}
}
