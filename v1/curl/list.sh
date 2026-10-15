#!/usr/bin/env bash
# List the API gradings your key's owner made in the key's organization, newest first.
#
#   ./list.sh [--status COMPLETED,PARTIAL] [--limit 12] [--cursor <nextCursor>]
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

# -G sends each --data-urlencode as a query parameter.
query=()
while [ "$#" -gt 0 ]; do
  case "$1" in
    --status) query+=(--data-urlencode "status=$2"); shift 2 ;;
    --limit) query+=(--data-urlencode "limit=$2"); shift 2 ;;
    --cursor) query+=(--data-urlencode "cursor=$2"); shift 2 ;;
    *) echo "Usage: $0 [--status S1,S2] [--limit N] [--cursor C]" >&2; exit 1 ;;
  esac
done

response=$(curl -sS -w '\n%{http_code}' --retry 3 -G "$API_URL/grading" \
  -H "Authorization: Bearer $GRADESCRIPT_API_KEY" \
  "${query[@]}")
response=$(check "$response")

echo "$response" | jq -r '
  (.data[] | [
    .id,
    .status,
    .studentFullname,
    (if .totalMark == null then "-" else "\(.totalMark)/\(.maximumMark)" end),
    .createdAt
  ] | @tsv),
  (select(.pagination.hasNextItems) | "\nNext page: ./list.sh --cursor \(.pagination.nextCursor)")'
