#!/usr/bin/env bash
# Print one grading in full: marks, per-question results, annotations and attachments.
#
#   ./get.sh <grading-id>
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

response=$(curl -sS -w '\n%{http_code}' --retry 3 "$API_URL/grading/$1" \
  -H "Authorization: Bearer $GRADESCRIPT_API_KEY")
response=$(check "$response")

echo "$response" | jq '.data'
