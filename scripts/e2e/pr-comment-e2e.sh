#!/bin/sh
# The PR comment end-to-end run. On a pull_request run, report --pr-comment creates one
# comment, and a second report updates that same comment, so the pull request ends with one
# GlassOps comment and not two.
#
#   sh scripts/e2e/pr-comment-e2e.sh [glassops]
#
# It runs plan and gate on a dummy exec project (local CI layer, throwaway key), then report
# with the GitHub CI layer against the real pull request in GITHUB_REPOSITORY. Needs
# GITHUB_TOKEN with pull-requests: write, GITHUB_REPOSITORY, GITHUB_REF (refs/pull/N/merge),
# GITHUB_EVENT_NAME=pull_request and network access to the GitHub API. Needs sh, git, base64,
# head and node on the PATH, which the glassops image has. It posts a public comment on the
# pull request. The comment holds only the dummy project's results, and the script checks the
# comment holds no email address, org Id or actor login. It first deletes the report
# comments an earlier run of this job left on the same pull request, so "creates" is true
# every time. It leaves the final comment in place.
set -eu

: "${GITHUB_TOKEN:?GITHUB_TOKEN is not set}"
: "${GITHUB_REPOSITORY:?GITHUB_REPOSITORY is not set}"
: "${GITHUB_REF:?GITHUB_REF is not set}"
if [ "${GITHUB_EVENT_NAME:-}" != pull_request ]; then
  echo "pr-comment e2e: GITHUB_EVENT_NAME is ${GITHUB_EVENT_NAME:-unset}, this job runs on pull_request only" >&2
  exit 2
fi
case "$GITHUB_REF" in
  refs/pull/*/merge) E2E_PR_NUMBER=$(printf '%s' "$GITHUB_REF" | cut -d/ -f3) ;;
  *)
    echo "pr-comment e2e: GITHUB_REF is not refs/pull/<n>/merge" >&2
    exit 2
    ;;
esac
export E2E_PR_NUMBER

# shellcheck source=/dev/null
. "$(dirname "$0")/lib-exec-project.sh"
e2e_init "${1:-glassops}"
CHECK="node $SCRIPTS/pr-comment-check.js"
e2e_project

# greport runs report with the GitHub CI layer. The step summary file is not set, so the
# summary goes nowhere, and the comment is the one output under test.
greport() {
  "$GLASSOPS" report --dir "$REPO" --ci github --pr-comment > "$WORK/report.txt" 2>&1
}

expect_exit 0 "plan" plan --from HEAD~1 --to HEAD
expect_exit 0 "gate" gate --target prod

if $CHECK clear; then
  pass "earlier report comments cleared"
else
  fail "could not clear earlier report comments"
fi

if greport; then
  pass "report --pr-comment (first)"
else
  fail "the first report failed"
  cat "$WORK/report.txt" >&2
fi
if grep -q "::warning::" "$WORK/report.txt"; then
  fail "the first report warned"
  grep "::warning::" "$WORK/report.txt" >&2
fi
if $CHECK snapshot "$WORK/first.json"; then
  pass "the first report created exactly one comment"
else
  fail "the first report did not leave exactly one clean comment"
fi

# The PERMIT check adds a result, so the second report's comment differs from the first and
# an update can be told from nothing happening. The second report finds the comment by its
# marker.
expect_exit 0 "permit verify" permit verify --target prod
if greport; then
  pass "report --pr-comment (second)"
else
  fail "the second report failed"
  cat "$WORK/report.txt" >&2
fi
if grep -q "::warning::" "$WORK/report.txt"; then
  fail "the second report warned"
  grep "::warning::" "$WORK/report.txt" >&2
fi
if $CHECK same "$WORK/first.json"; then
  pass "the second report updated the same comment"
else
  fail "the second report did not update the first comment in place"
fi

e2e_done "pr-comment e2e"
