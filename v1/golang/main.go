// Command-line examples for the GradeScript Grading API.
//
//	go run . grade [file ...]   upload a script and wait for its grading
//	go run . wait <id>          wait for a grading and print its marks
//	go run . list [--status COMPLETED,PARTIAL] [--limit 12] [--cursor C]
//	go run . get <id>           print a grading in full
//	go run . retry <id>         grade a FAILED or PARTIAL grading again
//	go run . delete <id>        delete a finished grading
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"text/tabwriter"
	"time"
)

const usage = `Usage: go run . <command> [arguments]

Commands:
  grade [file ...]   upload a script and wait for its grading
                     (each file is one SCRIPT page; defaults to the sample script)
  wait <id>          wait for a grading and print its marks
  list [--status COMPLETED,PARTIAL] [--limit 12] [--cursor C]
  get <id>           print a grading in full
  retry <id>         grade a FAILED or PARTIAL grading again
  delete <id>        delete a finished grading`

// The server stops a grading after 25 minutes, so 30 minutes is plenty.
const (
	pollInterval = 5 * time.Second
	pollTimeout  = 30 * time.Minute
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, usage)
		os.Exit(2)
	}

	client, err := NewClientFromEnv()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	command, args := os.Args[1], os.Args[2:]
	switch command {
	case "grade":
		err = grade(client, args)
	case "wait":
		err = withID(args, "wait", func(id string) error { return wait(client, id) })
	case "list":
		err = list(client, args)
	case "get":
		err = withID(args, "get", func(id string) error { return get(client, id) })
	case "retry":
		err = withID(args, "retry", func(id string) error { return retry(client, id) })
	case "delete":
		err = withID(args, "delete", func(id string) error { return remove(client, id) })
	default:
		fmt.Fprintln(os.Stderr, usage)
		os.Exit(2)
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func withID(args []string, command string, run func(id string) error) error {
	if len(args) != 1 {
		return fmt.Errorf("usage: go run . %s <grading-id>", command)
	}
	return run(args[0])
}

// sampleScript is the repo's sample answer script, found relative to this source file.
func sampleScript() string {
	_, file, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(file), "..", "..", "sample-script", "Item-one-answer-script.jpeg")
}

func grade(client *Client, files []string) error {
	if len(files) == 0 {
		files = []string{sampleScript()}
	}

	// One SCRIPT attachment per file. The index sets the page order.
	attachments := make([]AttachmentInput, len(files))
	for i, path := range files {
		attachments[i] = AttachmentInput{Type: "SCRIPT", Path: path, Index: i + 1}
	}

	grading, err := client.CreateGrading(CreateGradingInput{
		StudentFullname: "Ada Lovelace",
		Subject:         "Mathematics",
		Level:           "S.2",
		Attachments:     attachments,
	})
	if err != nil {
		return err
	}
	fmt.Println("Queued grading", grading.ID)

	return wait(client, grading.ID)
}

func wait(client *Client, id string) error {
	grading, err := client.WaitForGrading(id, pollInterval, pollTimeout, func(status string) {
		fmt.Println(status)
	})
	if err != nil {
		return err
	}
	printSummary(grading)
	return nil
}

func list(client *Client, args []string) error {
	flags := flag.NewFlagSet("list", flag.ContinueOnError)
	status := flags.String("status", "", "comma-separated statuses, e.g. COMPLETED,PARTIAL")
	limit := flags.Int("limit", 0, "1 to 25 (default 12)")
	cursor := flags.String("cursor", "", "nextCursor from the previous page")
	if err := flags.Parse(args); err != nil {
		return err
	}

	gradings, pagination, err := client.ListGradings(ListGradingsInput{
		Status: *status,
		Limit:  *limit,
		Cursor: *cursor,
	})
	if err != nil {
		return err
	}

	table := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	for _, grading := range gradings {
		mark := "-"
		if grading.TotalMark != nil && grading.MaximumMark != nil {
			mark = formatMark(*grading.TotalMark) + "/" + formatMark(*grading.MaximumMark)
		}
		fmt.Fprintf(table, "%s\t%s\t%s\t%s\t%s\n",
			grading.ID, grading.Status, grading.StudentFullname, mark, grading.CreatedAt.Format(time.RFC3339))
	}
	table.Flush()

	if pagination != nil && pagination.HasNextItems {
		fmt.Printf("\nNext page: go run . list --cursor %s\n", pagination.NextCursor)
	}
	return nil
}

func get(client *Client, id string) error {
	grading, err := client.GetGrading(id)
	if err != nil {
		return err
	}
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(grading)
}

func retry(client *Client, id string) error {
	if _, err := client.RetryGrading(id); err != nil {
		return err
	}
	fmt.Println("Re-queued grading", id)

	return wait(client, id)
}

func remove(client *Client, id string) error {
	deletedID, err := client.DeleteGrading(id)
	if err != nil {
		return err
	}
	fmt.Println("Deleted grading", deletedID)
	return nil
}

func printSummary(grading *Grading) {
	fmt.Printf("\nGrading %s: %s\n", grading.ID, grading.Status)
	if grading.Title != "" {
		fmt.Println(grading.Title)
	}
	if grading.TotalMark != nil && grading.MaximumMark != nil {
		fmt.Printf("Mark: %s / %s\n", formatMark(*grading.TotalMark), formatMark(*grading.MaximumMark))
	}
	for _, question := range grading.Questions {
		review := ""
		if question.NeedsReview {
			review = " (needs review)"
		}
		fmt.Printf("  %s\t%s / %s\t%s%s\n",
			question.Ref, formatMark(question.Awarded), formatMark(question.Max), question.Topic, review)
	}
	if grading.Feedback != "" {
		fmt.Printf("\nFeedback:\n%s\n", grading.Feedback)
	}
	if grading.Error != "" {
		fmt.Printf("\nError: %s\n", grading.Error)
	}
}

// formatMark shows 5.0 as 5 and 2.5 as 2.5.
func formatMark(mark float64) string {
	return strconv.FormatFloat(mark, 'f', -1, 64)
}
