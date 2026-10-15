# GradeScript API examples

Runnable examples for the [GradeScript Grading API](https://console.gradescriptug.com/docs), in the
same four languages as the API reference: **cURL, Node.js, Python and Go**.

Each example uploads a handwritten answer script, waits for it to be graded, and prints the marks
and feedback. It also lists, fetches, retries and deletes gradings, so every endpoint is covered.

## Before you start

1. **Create an API key** in the developer console at
   [console.gradescriptug.com/u/api-keys](https://console.gradescriptug.com/u/api-keys). Give it
   the permissions for the commands you want to run (see [Commands](#commands)). The full key
   (`gsk_...`) is shown only once, so copy it.
2. **Check your token balance.** Gradings are paid from the key owner's balance, and creating or
   retrying a grading needs at least 100 tokens. An API grading costs more tokens than the same
   grading in the web app. Top up in the console under **Settings → Billing**.
3. **Export the key:**

   ```bash
   export GRADESCRIPT_API_KEY="gsk_..."
   ```

Keep the key on your server. Don't put it in frontend code: the API doesn't accept browser
requests from other sites anyway.

| Variable              | Required | Default                                   |
| --------------------- | -------- | ----------------------------------------- |
| `GRADESCRIPT_API_KEY` | Yes      |                                           |
| `GRADESCRIPT_API_URL` | No       | `https://api.gradescriptug.com/client/v1` |

## Run an example

With no arguments, `grade` uses the sample script in
[`sample-script/`](sample-script/Item-one-answer-script.jpeg).

```bash
# cURL (needs curl and jq)
cd v1/curl && ./grade.sh

# Node.js 20 or later (no dependencies)
cd v1/nodejs && node grade.mjs

# Python 3.8 or later
cd v1/python && pip install -r requirements.txt && python grade.py

# Go 1.21 or later (standard library only)
cd v1/golang && go run . grade
```

The output looks like this:

```text
Queued grading 0b7e4f8e-...
PENDING
PROCESSING
COMPLETED

Grading 0b7e4f8e-...: COMPLETED
Mark: 5 / 7
  1    2 / 2   Linear equations
  ...

Feedback:
...
```

Each language folder has its own README with more detail.

## Commands

Every language has the same commands:

| Command                                          | What it does                                                                        | Permissions                          |
| ------------------------------------------------ | ----------------------------------------------------------------------------------- | ------------------------------------ |
| `grade [file ...]`                               | Uploads each file as one `SCRIPT` page, in order, then waits and prints the marks.  | `grade:create:any`, `grade:view:any` |
| `wait <id>`                                      | Waits for a grading to finish and prints its marks.                                 | `grade:view:any`                     |
| `list [--status S1,S2] [--limit N] [--cursor C]` | Lists your API gradings, newest first, one page at a time.                          | `grade:view:any`                     |
| `get <id>`                                       | Prints the full grading: questions, annotations and attachments.                    | `grade:view:any`                     |
| `retry <id>`                                     | Grades a `FAILED` or `PARTIAL` grading again, then waits.                           | `grade:edit:any`, `grade:view:any`   |
| `delete <id>`                                    | Deletes a finished grading. This can't be undone.                                   | `grade:delete:any`                   |

## How grading works

1. `POST /grading` with a `multipart/form-data` body. You get `202` and a grading with status
   `PENDING`.
2. Poll `GET /grading/{id}` every 5 seconds. The status moves from `PENDING` to `PROCESSING`,
   then to `COMPLETED`, `PARTIAL` or `FAILED`. There are no webhooks. The server stops a grading
   after 25 minutes.
3. Read the result: `totalMark`, `maximumMark`, `feedback`, `questions[]` and `annotations[]`.

Things that trip up integrations:

- **Creating a grading isn't idempotent.** Don't retry `POST /grading` automatically, or the
  script may be graded (and billed) twice. If a create request times out, check `list` before
  sending it again. The examples retry only `GET` requests.
- **Only send the documented form fields:** `studentFullname`, `subject`, `studentNumber`,
  `level`, `assessmentType`, `instructions` and `attachments[N][type|file|index]`. Any other
  field is rejected with `400`.
- **Let your HTTP library set the `Content-Type`.** It must include the multipart boundary.
- **Files** must be PDF, PNG, JPEG or WEBP. Each file can be up to 20 MB, a request up to 60 MB,
  and a grading can have up to 20 files and 40 pages. At least one attachment must be a `SCRIPT`.

See the [API reference](https://console.gradescriptug.com/docs/api) for every field.

## Troubleshooting

Errors come back as `{"status": "fail", "error": "UNAUTHORIZED", "message": "..."}`. The examples
print them as `HTTP 401 UNAUTHORIZED: Invalid API key!`.

| Status | Meaning                                                                                        |
| ------ | ---------------------------------------------------------------------------------------------- |
| `401`  | The key is missing, malformed, revoked or expired.                                             |
| `402`  | The key owner has fewer than 100 tokens. Top up in the console.                                |
| `403`  | The key doesn't have the permission for this endpoint, or its owner left the organization.     |
| `404`  | No grading with that id was made through the API by this key's owner in its organization.      |
| `409`  | `retry` on a grading that isn't `FAILED` or `PARTIAL`, or `delete` on one that's still running. |
| `413`  | A file is larger than 20 MB, or the request is larger than 60 MB.                              |
| `415`  | The request body isn't `multipart/form-data`.                                                  |

## Layout

```text
.
├── sample-script/
│   └── Item-one-answer-script.jpeg   a sample handwritten answer script
└── v1/                               examples for /client/v1
    ├── curl/
    ├── nodejs/
    ├── python/
    └── golang/
```
