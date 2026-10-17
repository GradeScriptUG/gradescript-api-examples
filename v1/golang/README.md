# Go examples

A command-line program that calls the GradeScript Grading API. It uses only the standard
library.

## Requirements

- Go 1.21 or later
- `GRADESCRIPT_API_KEY` in the `.env` file at the repo root (see the
  [main README](../../README.md#before-you-start))

## Usage

```bash
go run . grade                            # grade the sample script and wait for the result
go run . grade page-1.jpg page-2.jpg      # each file is one SCRIPT page, in order
go run . wait <grading-id>                # wait for a grading and print its marks
go run . list --status COMPLETED,PARTIAL  # list gradings; also --limit N and --cursor C
go run . get <grading-id>                 # print the full grading as JSON
go run . retry <grading-id>               # grade a FAILED or PARTIAL grading again
go run . delete <grading-id>              # delete a finished grading
```

To build a binary instead, run `go build -o gradescript .`, then `./gradescript grade`.

## Using the client in your own code

[`client.go`](client.go) is a small client you can copy into your project. It has typed
`Grading`, `Question`, `Annotation` and `Attachment` structs.

```go
client, err := NewClientFromEnv()
if err != nil {
	log.Fatal(err)
}

queued, err := client.CreateGrading(CreateGradingInput{
	StudentFullname: "Ada Lovelace",
	Subject:         "Mathematics",
	Level:           "S.2",
	Attachments: []AttachmentInput{
		{Type: "SCRIPT", Path: "page-1.jpg", Index: 1},
		{Type: "SCRIPT", Path: "page-2.jpg", Index: 2},
		{Type: "MARKING_GUIDE", Path: "marking-guide.pdf"},
	},
})
if err != nil {
	log.Fatal(err)
}

grading, err := client.WaitForGrading(queued.ID, 5*time.Second, 30*time.Minute, func(status string) {
	log.Println(status)
})
if err != nil {
	log.Fatal(err)
}
log.Println(grading.Status, grading.Feedback)
```

The client also has `ListGradings`, `GetGrading`, `RetryGrading` and `DeleteGrading`. API
errors are returned as an `*APIError`, which has `StatusCode` and `Name` (for example
`PAYMENT_REQUIRED`). GET requests are retried on network errors, 429 and 5xx. Create and retry
requests aren't, so a script is never graded twice.
