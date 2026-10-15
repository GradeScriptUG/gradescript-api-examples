# cURL examples

Bash scripts that call the GradeScript Grading API with `curl`. Each script shows its full
`curl` command, so you can copy it into your own scripts.

## Requirements

- Bash, `curl` and [`jq`](https://jqlang.org/)
- `GRADESCRIPT_API_KEY` set to your API key (see the [main README](../../README.md#before-you-start))

## Usage

```bash
./grade.sh                              # grade the sample script and wait for the result
./grade.sh page-1.jpg page-2.jpg        # each file is one SCRIPT page, in order
./wait.sh <grading-id>                  # wait for a grading and print its marks
./list.sh --status COMPLETED,PARTIAL    # list gradings; also --limit N and --cursor C
./get.sh <grading-id>                   # print the full grading as JSON
./retry.sh <grading-id>                 # grade a FAILED or PARTIAL grading again
./delete.sh <grading-id>                # delete a finished grading
```

## Files

| File        | Endpoint                                          |
| ----------- | ------------------------------------------------- |
| `grade.sh`  | `POST /grading`, then runs `wait.sh`              |
| `wait.sh`   | `GET /grading/{id}` every 5 seconds until done    |
| `list.sh`   | `GET /grading?status=&limit=&cursor=`             |
| `get.sh`    | `GET /grading/{id}`                               |
| `retry.sh`  | `PATCH /grading/{id}/retry`, then runs `wait.sh`  |
| `delete.sh` | `DELETE /grading/{id}`                            |

The `curl` calls use `-w '\n%{http_code}'` to add the status code as the last line of the
output. The `check` function in each script removes that line and stops on an error. GET requests
also pass `--retry 3`. The create and retry requests don't, so a script is never graded twice.
