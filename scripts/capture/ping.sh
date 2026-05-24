#!/usr/bin/env bash
# ping.sh — validate that HTB_API_KEY auth works against the live API.
#
# Reads HTB_API_KEY from .env (or the environment) and makes one
# read-only request to /user/info. Prints the response status and a
# scrubbed summary of the body. Does NOT write the response to disk.
#
# Usage:
#   ./scripts/capture/ping.sh
#   HTB_API_BASE_URL=https://labs.hackthebox.com/api/v4 ./scripts/capture/ping.sh

set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"

# Load .env if present, without leaking values into the shell history.
if [[ -f "$ROOT/.env" ]]; then
  set -a
  # shellcheck disable=SC1091
  source "$ROOT/.env"
  set +a
fi

: "${HTB_API_KEY:?HTB_API_KEY is not set. Add it to .env (see .env.example).}"
: "${HTB_API_BASE_URL:=https://labs.hackthebox.com/api/v4}"

URL="${HTB_API_BASE_URL%/}/user/info"

http_code=$(curl --silent --show-error \
  --output /tmp/htb-ping.$$.json \
  --write-out '%{http_code}' \
  --header "Authorization: Bearer $HTB_API_KEY" \
  --header "Accept: application/json" \
  --header "User-Agent: htb-app-mcp/ping" \
  --max-time 15 \
  "$URL")

printf 'GET %s -> %s\n' "$URL" "$http_code"

case "$http_code" in
  200)
    # Print a small, scrubbed summary. Never echo the raw response — it
    # contains identifying user data.
    if command -v jq >/dev/null 2>&1; then
      jq --raw-output '
        {
          id: (.info.id // .id // null),
          name_len: ((.info.name // .name // "") | length),
          has_team: ((.info.team // .team // null) != null)
        } | "ok: user id=\(.id) name_len=\(.name_len) has_team=\(.has_team)"
      ' /tmp/htb-ping.$$.json
    else
      echo "ok: response received (install jq to see scrubbed summary)"
    fi
    ;;
  401|403)
    echo "auth failed: token is missing, expired, or revoked. Regenerate at"
    echo "https://app.hackthebox.com/profile/settings"
    exit 2
    ;;
  429)
    echo "rate limited. Check Retry-After header on next attempt."
    exit 3
    ;;
  *)
    echo "unexpected status. Body (first 200 chars):"
    head -c 200 /tmp/htb-ping.$$.json
    echo
    exit 1
    ;;
esac

rm -f /tmp/htb-ping.$$.json
