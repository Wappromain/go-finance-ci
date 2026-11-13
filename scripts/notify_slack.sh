#!/usr/bin/env bash
# Posts a summary of a workflow run to Slack through an incoming webhook.
#
# Required: SLACK_WEBHOOK_URL
# Set by the workflow (all optional when testing by hand):
#   STATUS        success | failure | cancelled
#   TEST_RESULT   result of the test job
#   IMAGE_RESULT  result of the API image job
#   PASSED, FAILED  number of top level tests that passed / failed
#   COMMIT_MSG    first line of the commit message
#   plus the standard GITHUB_* variables Actions provides
#
# Test from your laptop:
#   SLACK_WEBHOOK_URL=https://hooks.slack.com/services/... ./scripts/notify_slack.sh
set -euo pipefail

if [ -z "${SLACK_WEBHOOK_URL:-}" ]; then
  echo "::warning::SLACK_WEBHOOK_URL is not set, skipping Slack notification"
  exit 0
fi

STATUS="${STATUS:-success}"
REPO="${GITHUB_REPOSITORY:-local/go-finance-ci}"
BRANCH="${GITHUB_REF_NAME:-$(git rev-parse --abbrev-ref HEAD 2>/dev/null || echo unknown)}"
SHA="${GITHUB_SHA:-$(git rev-parse HEAD 2>/dev/null || echo 0000000)}"
ACTOR="${GITHUB_ACTOR:-$(whoami)}"
EVENT="${GITHUB_EVENT_NAME:-manual}"
WORKFLOW="${GITHUB_WORKFLOW:-API tests}"
SERVER="${GITHUB_SERVER_URL:-https://github.com}"
RUN_URL="$SERVER/$REPO/actions/runs/${GITHUB_RUN_ID:-0}"
COMMIT_URL="$SERVER/$REPO/commit/$SHA"
COMMIT_MSG="${COMMIT_MSG:-}"
COMMIT_MSG="${COMMIT_MSG%%$'\n'*}"   # keep only the first line
[ -z "$COMMIT_MSG" ] && COMMIT_MSG="(no commit message: $EVENT run)"

case "$STATUS" in
  success)   ICON=":white_check_mark:"; WORD="passed" ;;
  cancelled) ICON=":warning:";          WORD="was cancelled" ;;
  *)         ICON=":x:";                WORD="failed" ;;
esac

label() {   # turns a job result into an emoji + word
  case "$1" in
    success) echo ":white_check_mark: success" ;;
    failure) echo ":x: failure" ;;
    cancelled) echo ":warning: cancelled" ;;
    skipped) echo ":fast_forward: skipped" ;;
    *) echo "${1:-n/a}" ;;
  esac
}

# jq builds the JSON so quotes or symbols in commit messages can't break it.
PAYLOAD=$(jq -n \
  --arg text "$ICON $WORKFLOW $WORD on $REPO@$BRANCH" \
  --arg header "$ICON $WORKFLOW $WORD" \
  --arg repo "$REPO" --arg branch "$BRANCH" \
  --arg sha "${SHA:0:7}" --arg commit_url "$COMMIT_URL" --arg msg "$COMMIT_MSG" \
  --arg actor "$ACTOR" --arg event "$EVENT" \
  --arg tests "$(label "${TEST_RESULT:-}")  (${PASSED:-?} passed, ${FAILED:-?} failed)" \
  --arg image "$(label "${IMAGE_RESULT:-}")" \
  --arg run_url "$RUN_URL" \
  'def esc: gsub("&"; "&amp;") | gsub("<"; "&lt;") | gsub(">"; "&gt;");
  {
    text: $text,
    blocks: [
      { type: "header", text: { type: "plain_text", text: $header } },
      { type: "section", fields: [
          { type: "mrkdwn", text: "*Repository*\n\($repo)" },
          { type: "mrkdwn", text: "*Branch*\n\($branch)" },
          { type: "mrkdwn", text: "*Commit*\n<\($commit_url)|\($sha)> \($msg | esc)" },
          { type: "mrkdwn", text: "*Triggered by*\n\($actor) (\($event))" },
          { type: "mrkdwn", text: "*Tests (Docker)*\n\($tests)" },
          { type: "mrkdwn", text: "*API image*\n\($image)" }
      ]},
      { type: "actions", elements: [
          { type: "button", text: { type: "plain_text", text: "View workflow run" }, url: $run_url }
      ]}
    ]
  }')

echo "==> Posting to Slack"
curl --fail --silent --show-error -X POST \
  -H 'Content-Type: application/json' \
  --data "$PAYLOAD" \
  "$SLACK_WEBHOOK_URL"
echo
echo "==> Slack notification sent"
