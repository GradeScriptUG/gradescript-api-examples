# Python examples

Scripts that call the GradeScript Grading API with [`requests`](https://requests.readthedocs.io/),
and read the `.env` file with [`python-dotenv`](https://pypi.org/project/python-dotenv/).

## Requirements

- Python 3.8 or later
- `GRADESCRIPT_API_KEY` in the `.env` file at the repo root (see the
  [main README](../../README.md#before-you-start))

```bash
python -m venv .venv && source .venv/bin/activate
pip install -r requirements.txt
```

## Usage

```bash
python grade.py                            # grade the sample script and wait for the result
python grade.py page-1.jpg page-2.jpg      # each file is one SCRIPT page, in order
python wait.py <grading-id>                # wait for a grading and print its marks
python list.py --status COMPLETED,PARTIAL  # list gradings; also --limit N and --cursor C
python get.py <grading-id>                 # print the full grading as JSON
python retry.py <grading-id>               # grade a FAILED or PARTIAL grading again
python delete.py <grading-id>              # delete a finished grading
```

## Using the client in your own code

[`gradescript.py`](gradescript.py) is a small client you can copy into your project:

```python
from gradescript import create_grading, wait_for_grading

queued = create_grading(
    {"studentFullname": "Ada Lovelace", "subject": "Mathematics", "level": "S.2"},
    [
        {"type": "SCRIPT", "path": "page-1.jpg", "index": 1},
        {"type": "SCRIPT", "path": "page-2.jpg", "index": 2},
        {"type": "MARKING_GUIDE", "path": "marking-guide.pdf"},
    ],
)

grading = wait_for_grading(queued["id"])
print(grading["totalMark"], "/", grading["maximumMark"])
```

It also has `list_gradings`, `get_grading`, `retry_grading` and `delete_grading`. API errors are
raised as a `GradeScriptError`, which has `status_code` and `error` (for example
`PAYMENT_REQUIRED`). GET requests are retried on network errors, 429 and 5xx. Create and retry
requests aren't, so a script is never graded twice.
