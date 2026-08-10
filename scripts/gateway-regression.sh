#!/bin/sh

set -eu

gateway_url=${GATEWAY_URL:-http://localhost:3000}
workspace=$(mktemp -d)
trap 'rm -rf "$workspace"' EXIT HUP INT TERM

username="gateway-regression-$(date +%s)-$$"
password='GatewayRegression123'

fail() {
  message=$1
  body_file=${2:-}
  printf '%s\n' "$message" >&2
  if [ -n "$body_file" ] && [ -f "$body_file" ]; then
    printf 'Response body:\n' >&2
    cat "$body_file" >&2
    printf '\n' >&2
  fi
  exit 1
}

request() {
  output_file=$1
  shift
  curl --silent --show-error --output "$output_file" --write-out '%{http_code}' "$@" || fail 'Gateway request could not be completed.' "$output_file"
}

expect_status() {
  expected=$1
  actual=$2
  step=$3
  body_file=$4
  if [ "$actual" != "$expected" ]; then
    fail "$step returned HTTP $actual; expected $expected." "$body_file"
  fi
}

register_body="$workspace/register.json"
register_status=$(request "$register_body" \
  --header 'Content-Type: application/json' \
  --data "{\"username\":\"$username\",\"password\":\"$password\",\"training_role\":\"buyer\"}" \
  "$gateway_url/api/v1/auth/register")
expect_status 201 "$register_status" 'Registration' "$register_body"

login_body="$workspace/login.txt"
cookie_jar="$workspace/cookies.txt"
login_status=$(request "$login_body" \
  --cookie-jar "$cookie_jar" \
  --header 'Content-Type: application/json' \
  --data "{\"username\":\"$username\",\"password\":\"$password\"}" \
  "$gateway_url/api/v1/auth/login")
expect_status 204 "$login_status" 'Login' "$login_body"
grep -q 'access_token' "$cookie_jar" || fail 'Login did not issue an access_token cookie.' "$login_body"

dashboard_body="$workspace/dashboard.json"
dashboard_status=$(request "$dashboard_body" \
  --cookie "$cookie_jar" \
  "$gateway_url/api/v1/dashboard?role=buyer")
expect_status 200 "$dashboard_status" 'Dashboard' "$dashboard_body"
grep -q '"daily_task":' "$dashboard_body" || fail 'Dashboard response does not contain a daily_task.' "$dashboard_body"
if grep -q '"code":"INTERNAL_ERROR"' "$dashboard_body"; then
  fail 'Dashboard returned INTERNAL_ERROR.' "$dashboard_body"
fi

printf 'Gateway regression passed: Register 201, Login 204 with access_token, Dashboard 200.\n'
