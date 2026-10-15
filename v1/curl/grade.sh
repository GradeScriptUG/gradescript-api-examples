#!/usr/bin/env bash
# Grade a script: upload it, then wait for the grading to finish.
#
#   ./grade.sh                        grades the sample script in ../../sample-script
#   ./grade.sh page-1.jpg page-2.jpg  each file is one SCRIPT page, in order
set -eo pipefail

API_URL="${GRADESCRIPT_API_URL:-https://api.gradescriptug.com/client/v1}"
API_URL="${API_URL%/}"
if [ -z "$GRADESCRIPT_API_KEY" ]; then
  echo "Set GRADESCRIPT_API_KEY to your API key (gsk_...)." >&2
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

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
if [ "$#" -eq 0 ]; then
  set -- "$SCRIPT_DIR/../../sample-script/Item-one-answer-script.jpeg"
fi

# One SCRIPT attachment per file. The index sets the page order.
attachments=()
i=0
for file in "$@"; do
  attachments+=(
    -F "attachments[$i][type]=SCRIPT"
    -F "attachments[$i][index]=$((i + 1))"
    -F "attachments[$i][file]=@$file"
  )
  i=$((i + 1))
done

# Create the grading. curl sets the multipart Content-Type (and its boundary) itself.
# Don't add --retry here: creating twice would grade, and bill, the script twice.
response=$(curl -sS -w '\n%{http_code}' -X POST "$API_URL/grading" \
  -H "Authorization: Bearer $GRADESCRIPT_API_KEY" \
  -F "studentFullname=Ada Lovelace" \
  -F "subject=Mathematics" \
  -F "level=S.2" \
  "${attachments[@]}")
response=$(check "$response")

id=$(echo "$response" | jq -r '.data.id')
echo "Queued grading $id"

exec "$SCRIPT_DIR/wait.sh" "$id"
