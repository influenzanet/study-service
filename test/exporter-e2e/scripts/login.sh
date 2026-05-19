#!/usr/bin/env bash
# Log in as the admin user defined in .env, save the JWT to .token.
# The same .token works against both stable and refactor lab runs because
# the same JWT_TOKEN_KEY is shared via docker-compose.
set -euo pipefail

cd "$(dirname "$0")/.."

if [[ -f .env ]]; then
  # shellcheck disable=SC1091
  set -a; source .env; set +a
fi

: "${ADMIN_EMAIL:?ADMIN_EMAIL not set in .env}"
: "${ADMIN_PASSWORD:?ADMIN_PASSWORD not set in .env}"
: "${INSTANCE_ID:?INSTANCE_ID not set in .env}"

BASE_URL="${BASE_URL:-http://localhost:3232}"

echo "Logging in as ${ADMIN_EMAIL} on instance ${INSTANCE_ID}..."

body=$(mktemp)
http_code=$(curl -sS -o "$body" -w '%{http_code}' -X POST "${BASE_URL}/v1/auth/login-with-email" \
  -H 'Content-Type: application/json' \
  -d "$(jq -nc \
        --arg email    "$ADMIN_EMAIL" \
        --arg password "$ADMIN_PASSWORD" \
        --arg inst     "$INSTANCE_ID" \
        '{email:$email, password:$password, instanceId:$inst}')")

if [[ "$http_code" != "200" ]]; then
  echo "login failed (HTTP $http_code):" >&2
  cat "$body" >&2
  rm -f "$body"
  exit 1
fi

token=$(jq -r '.token.accessToken // .accessToken' "$body")
rm -f "$body"

if [[ -z "$token" || "$token" == "null" || "$token" != eyJ* ]]; then
  echo "login response did not contain a valid access token" >&2
  exit 1
fi

printf '%s' "$token" > .token
echo "Token saved to .token ($(wc -c <.token) bytes)."
