#!/usr/bin/env bash

set -Eeuo pipefail

base_url="${BASE_URL:-http://127.0.0.1:8080}"
docker_command="${DOCKER_COMMAND:-docker}"
owner_cookie="$(mktemp)"
other_cookie="$(mktemp)"
foreign_body="$(mktemp)"

cleanup() {
  rm -f "$owner_cookie" "$other_cookie" "$foreign_body"
}
trap cleanup EXIT

fail() {
  echo "smoke failure: $*" >&2
  exit 1
}

assert_json() {
  local body="$1"
  local expression="$2"
  local message="$3"

  if ! jq -e "$expression" >/dev/null <<<"$body"; then
    echo "$body" | jq . >&2 || echo "$body" >&2
    fail "$message (jq: $expression)"
  fi
}

wait_ready() {
  local timeout_seconds="${1:-120}"
  local elapsed=0

  until curl --fail --silent --show-error "$base_url/readyz" >/dev/null 2>&1; do
    if (( elapsed >= timeout_seconds )); then
      "$docker_command" compose ps >&2 || true
      "$docker_command" compose logs --no-color >&2 || true
      fail "backend did not become ready within ${timeout_seconds}s"
    fi
    sleep 2
    elapsed=$((elapsed + 2))
  done
}

owner_get() {
  curl --fail --silent --show-error \
    --cookie "$owner_cookie" --cookie-jar "$owner_cookie" \
    "$base_url$1"
}

owner_post() {
  local path="$1"
  local payload="$2"

  curl --fail --silent --show-error \
    --cookie "$owner_cookie" --cookie-jar "$owner_cookie" \
    --header 'Content-Type: application/json' \
    --request POST --data "$payload" \
    "$base_url$path"
}

submit_choice() {
  local attempt_id="$1"
  local node_id="$2"
  local choice_id="$3"
  local idempotency_key="$4"
  local payload

  payload="$(jq -cn \
    --arg node "$node_id" \
    --arg choice "$choice_id" \
    --arg key "$idempotency_key" \
    '{nodeId: $node, choiceId: $choice, idempotencyKey: $key}')"

  owner_post "/api/v1/attempts/$attempt_id/choices" "$payload"
}

command -v jq >/dev/null 2>&1 || fail "jq is required"
command -v curl >/dev/null 2>&1 || fail "curl is required"

wait_ready

health="$(curl --fail --silent --show-error "$base_url/healthz")"
assert_json "$health" '.status == "ok"' 'liveness response is invalid'

readiness="$(curl --fail --silent --show-error "$base_url/readyz")"
assert_json "$readiness" '.status == "ready"' 'readiness response is invalid'

catalog="$(owner_get '/api/v1/scenarios')"
assert_json "$catalog" \
  '[.scenarios[] | select(.id == "buyer-fake-delivery" and .role == "buyer")] | length == 1' \
  'expected scenario is absent from the active catalog'

initial_progress="$(owner_get '/api/v1/progress')"
assert_json "$initial_progress" \
  '.summary.completedAttempts == 0 and .activeAttempts == [] and .recentAttempts == []' \
  'new anonymous profile must start with empty progress'

started="$(owner_post '/api/v1/attempts' '{"scenarioId":"buyer-fake-delivery"}')"
assert_json "$started" \
  '.status == "in_progress" and .currentNodeId == "channel-decision" and .score == 100 and (.scenario.version >= 1)' \
  'attempt did not start at the expected decision'

attempt_id="$(jq -er '.attemptId | select(length > 0)' <<<"$started")"

step_one="$(submit_choice "$attempt_id" 'channel-decision' 'stay-on-platform' 'ci-safe-step-1')"
assert_json "$step_one" \
  '.status == "in_progress" and .currentNodeId == "link-decision" and .acceptedChoice.choiceId == "stay-on-platform"' \
  'first safe decision produced an unexpected transition'

step_two="$(submit_choice "$attempt_id" 'link-decision' 'check-in-app' 'ci-safe-step-2')"
assert_json "$step_two" \
  '.status == "in_progress" and .currentNodeId == "prepay-decision" and .acceptedChoice.choiceId == "check-in-app"' \
  'second safe decision produced an unexpected transition'

step_three="$(submit_choice "$attempt_id" 'prepay-decision' 'refuse-prepay' 'ci-safe-step-3')"
assert_json "$step_three" \
  '.status == "completed" and .currentNodeId == "safe-ending" and .outcome == "safe" and .score == 100 and (.completedAt | length > 0)' \
  'safe scenario did not complete with the expected result'

replayed="$(submit_choice "$attempt_id" 'prepay-decision' 'refuse-prepay' 'ci-safe-step-3')"
if [[ "$(jq -cS . <<<"$replayed")" != "$(jq -cS . <<<"$step_three")" ]]; then
  fail 'idempotent replay returned a different transition'
fi

restored="$(owner_get "/api/v1/attempts/$attempt_id")"
assert_json "$restored" \
  ".attemptId == \"$attempt_id\" and .status == \"completed\" and .score == 100 and (.decisions | length == 3)" \
  'restored attempt is incomplete or contains duplicated decisions'

progress="$(owner_get '/api/v1/progress')"
assert_json "$progress" \
  ".summary.completedAttempts == 1 and .summary.completedScenarios == 1 and .summary.averageScore == 100 and .summary.bestScore == 100 and .summary.latestScore == 100 and (.activeAttempts | length == 0) and (.recentAttempts | length == 1) and .recentAttempts[0].attemptId == \"$attempt_id\" and .recentAttempts[0].decisions == 3" \
  'completed attempt is not reflected exactly once in progress'

foreign_status="$(curl --silent --show-error \
  --output "$foreign_body" --write-out '%{http_code}' \
  --cookie "$other_cookie" --cookie-jar "$other_cookie" \
  "$base_url/api/v1/attempts/$attempt_id")"
if [[ "$foreign_status" != '403' ]]; then
  cat "$foreign_body" >&2
  fail "foreign profile received HTTP $foreign_status instead of 403"
fi

foreign_json="$(<"$foreign_body")"
assert_json "$foreign_json" \
  '.error.code == "ATTEMPT_FORBIDDEN"' \
  'foreign profile received an unexpected error code'

# Applying all migrations again must be safe against the already initialized
# schema. --no-deps avoids changing the state of the running services.
"$docker_command" compose run --rm --no-deps migrate up

"$docker_command" compose restart backend
wait_ready

after_restart="$(owner_get "/api/v1/attempts/$attempt_id")"
assert_json "$after_restart" \
  ".attemptId == \"$attempt_id\" and .status == \"completed\" and .outcome == \"safe\" and .score == 100 and (.decisions | length == 3)" \
  'attempt did not survive backend restart'

progress_after_restart="$(owner_get '/api/v1/progress')"
assert_json "$progress_after_restart" \
  ".summary.completedAttempts == 1 and .summary.latestScore == 100 and (.recentAttempts | length == 1) and .recentAttempts[0].attemptId == \"$attempt_id\"" \
  'progress did not survive backend restart'

echo "Compose smoke passed: attempt $attempt_id persisted after restart"
