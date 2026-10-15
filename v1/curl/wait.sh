#!/usr/bin/env bash
# Poll a grading until it finishes, then print its marks and feedback.
# There are no webhooks, so polling is the only way to know a grading is done.
#
#   ./wait.sh <grading-id>
set -eo pipefail

API_URL="${GRADESCRIPT_API_URL:-https://api.gradescriptug.com/client/v1}"
API_URL="${API_URL%/}"
if [ -z "$GRADESCRIPT_API_KEY" ]; then
  echo "Set GRADESCRIPT_API_KEY to your API key (gsk_...)." >&2
  exit 1
fi
if [ "$#" -ne 1 ]; then
  echo "Usage: $0 <grading-id>" >&2
  exit 1
fi

# curl's -w '\n%{http_code}' adds the status code as the last line. Print the
# body, or "HTTP <code> <ERROR>: <message>" and stop when the API returns an error.
check() {
  local body=${1%$'\n'*} code=${1##*$'\n'} message
  if [ "$code" -ge 400 ]; then
    message=$(echo "$body" | jq -r '"\(.error): \(.message)"' 2>/dev/null) || message=$body
    echo "HTTP $code $message" >&2
    exit 1
  fi
  echo "$body"
}

# The server stops a grading after 25 minutes, so 30 minutes is plenty.
deadline=$((SECONDS + 30 * 60))
last=""

while true; do
  # GETs are safe to repeat: --retry covers timeouts, 429 and 5xx.
  grading=$(curl -sS -w '\n%{http_code}' --retry 3 "$API_URL/grading/$1" \
    -H "Authorization: Bearer $GRADESCRIPT_API_KEY")
  grading=$(check "$grading")

  status=$(echo "$grading" | jq -r '.data.status')
  if [ "$status" != "$last" ]; then
    echo "$status"
    last=$status
  fi

  case "$status" in
    COMPLETED|PARTIAL|FAILED) break ;;
  esac

  if [ "$SECONDS" -ge "$deadline" ]; then
    echo "Grading $1 did not finish within 30 minutes." >&2
    exit 1
  fi
  sleep 5
done

echo "$grading" | jq -r '.data |
  "\nGrading \(.id): \(.status)",
  (select(.title != "") | .title),
  (select(.totalMark != null) | "Mark: \(.totalMark) / \(.maximumMark)"),
  (.questions[] | "  \(.ref)\t\(.awarded) / \(.max)\t\(.topic)\(if .needsReview then " (needs review)" else "" end)"),
  (select(.feedback != "") | "\nFeedback:\n\(.feedback)"),
  (select(.error != "") | "\nError: \(.error)")'
