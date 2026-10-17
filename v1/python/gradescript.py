"""A small client for the GradeScript Grading API: https://console.gradescriptug.com/docs

Python 3.8 or later with requests and python-dotenv (pip install -r requirements.txt).
"""
import os
import sys
import time

import requests
from dotenv import load_dotenv

# Read GRADESCRIPT_API_KEY (and GRADESCRIPT_API_URL) from the .env file at the repo root.
# Variables already set in the environment take precedence.
load_dotenv(os.path.join(os.path.dirname(os.path.abspath(__file__)), "..", "..", ".env"))

API_URL = (
    os.environ.get("GRADESCRIPT_API_URL") or "https://api.gradescriptug.com/client/v1"
).rstrip("/")

# A grading is finished once it reaches one of these statuses.
FINISHED = {"COMPLETED", "PARTIAL", "FAILED"}

# GETs are safe to repeat, so they are retried after these delays (seconds) on
# network errors, 429 and 5xx. POST and PATCH are never retried: creating a
# grading twice would grade, and bill, the script twice.
GET_RETRY_DELAYS = (1, 2, 4)


class GradeScriptError(Exception):
    """An error response from the API: {status, error, message}."""

    def __init__(self, status_code, error, message):
        super().__init__(f"HTTP {status_code} {error}: {message}")
        self.status_code = status_code
        self.error = error


def _request(method, path, **kwargs):
    api_key = os.environ.get("GRADESCRIPT_API_KEY")
    if not api_key:
        raise RuntimeError(
            "Set GRADESCRIPT_API_KEY in the .env file at the repo root (see .env.example)."
        )

    delays = GET_RETRY_DELAYS if method == "GET" else ()
    for attempt in range(len(delays) + 1):
        retries_left = attempt < len(delays)
        try:
            # requests sets the multipart Content-Type (and its boundary) from files=.
            response = requests.request(
                method,
                API_URL + path,
                headers={"Authorization": f"Bearer {api_key}"},
                timeout=(10, 120),
                **kwargs,
            )
        except (requests.ConnectionError, requests.Timeout):
            if not retries_left:
                raise
        else:
            if not retries_left or not (
                response.status_code == 429 or response.status_code >= 500
            ):
                break
        time.sleep(delays[attempt])

    try:
        body = response.json()
    except ValueError:
        body = {}
    if not response.ok:
        raise GradeScriptError(
            response.status_code,
            body.get("error", response.reason),
            body.get("message", ""),
        )
    return body


def create_grading(fields, attachments):
    """Upload a script for grading. Returns the queued grading.

    fields: studentFullname and subject (required), plus any of studentNumber,
        level, assessmentType and instructions. Any other field is rejected.
    attachments: [{"type": ..., "path": ..., "index": ...}] where type is SCRIPT,
        MARKING_GUIDE, QUESTION_PAPER or OTHER, and index (optional, from 1)
        orders files of the same type. At least one SCRIPT is required.
    """
    data = dict(fields)
    files = {}
    try:
        for i, attachment in enumerate(attachments):
            data[f"attachments[{i}][type]"] = attachment["type"]
            if "index" in attachment:
                data[f"attachments[{i}][index]"] = str(attachment["index"])
            files[f"attachments[{i}][file]"] = open(attachment["path"], "rb")

        return _request("POST", "/grading", data=data, files=files)["data"]
    finally:
        for file in files.values():
            file.close()


def list_gradings(status=None, limit=None, cursor=None):
    """Returns {"data": [...], "pagination": {...}}.

    Pass pagination["nextCursor"] as cursor to get the next page.
    """
    params = {"status": status, "limit": limit, "cursor": cursor}
    body = _request(
        "GET",
        "/grading",
        params={name: value for name, value in params.items() if value is not None},
    )
    return {"data": body["data"], "pagination": body["pagination"]}


def get_grading(grading_id):
    return _request("GET", f"/grading/{grading_id}")["data"]


def retry_grading(grading_id):
    return _request("PATCH", f"/grading/{grading_id}/retry")["data"]


def delete_grading(grading_id):
    return _request("DELETE", f"/grading/{grading_id}")["data"]


def wait_for_grading(grading_id, interval=5, timeout=30 * 60):
    """Poll a grading every 5 seconds until it finishes. There are no webhooks.

    The server stops a grading after 25 minutes, so the 30-minute timeout is plenty.
    """
    deadline = time.monotonic() + timeout
    last_status = None

    while True:
        grading = get_grading(grading_id)
        if grading["status"] != last_status:
            print(grading["status"], flush=True)
            last_status = grading["status"]
        if grading["status"] in FINISHED:
            return grading

        if time.monotonic() >= deadline:
            raise TimeoutError(
                f"Grading {grading_id} did not finish within {timeout // 60} minutes."
            )
        time.sleep(interval)


def print_summary(grading):
    print(f"\nGrading {grading['id']}: {grading['status']}")
    if grading["title"]:
        print(grading["title"])
    if grading["totalMark"] is not None:
        print(f"Mark: {grading['totalMark']:g} / {grading['maximumMark']:g}")
    for question in grading["questions"]:
        review = " (needs review)" if question["needsReview"] else ""
        print(
            f"  {question['ref']}\t{question['awarded']:g} / {question['max']:g}"
            f"\t{question['topic']}{review}"
        )
    if grading["feedback"]:
        print(f"\nFeedback:\n{grading['feedback']}")
    if grading["error"]:
        print(f"\nError: {grading['error']}")


def run(main):
    """Run a command and print any error as one line instead of a traceback."""
    try:
        main()
    except (GradeScriptError, requests.RequestException, OSError, RuntimeError) as error:
        sys.exit(str(error))
    except KeyboardInterrupt:
        sys.exit(130)
