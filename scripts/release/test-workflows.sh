#!/usr/bin/env bash
# Runs the "Install the glassops command" step of both reusable workflows with git and docker
# stubbed. Checks the default image rule and the wrapper the step writes.
#   scripts/release/test-workflows.sh
set -uo pipefail
here="$(cd "$(dirname "$0")" && pwd)"
root="$(cd "$here/../.." && pwd)"
py=""
for cand in python3 python; do
  if "$cand" -c 'import sys' >/dev/null 2>&1; then py="$cand"; break; fi
done
[ -n "$py" ] || { echo "no working python3 or python" >&2; exit 2; }

work="$(mktemp -d)"
trap 'rm -rf "$work"' EXIT
mkdir -p "$work/bin"

# git stub: ls-remote prints the canned tags and records the URL, anything else fails.
cat > "$work/bin/git" <<'STUB'
#!/usr/bin/env bash
if [ "$1" = "ls-remote" ]; then
  echo "$3" > "$FAKE_GIT_URL"
  if [ "${FAKE_LS_EXIT:-0}" != 0 ]; then exit "$FAKE_LS_EXIT"; fi
  cat "$FAKE_LS_REMOTE"
fi
echo "unexpected git call: $*" >&2
exit 99
STUB
cat > "$work/bin/docker" <<'STUB'
#!/usr/bin/env bash
printf '%s\n' "$@" > "$FAKE_DOCKER_LOG"
STUB
chmod +x "$work/bin/git" "$work/bin/docker"

sha=1111111111111111111111111111111111111111
other=2222222222222222222222222222222222222222

fail=0
ok() { echo "ok   $1"; }
bad() { echo "FAIL $1"; fail=1; }

# step <workflow> <input-image> <ls-remote file>: runs the install step, sets $code, $image.
step() {
  local wf="$1" input="$2" tags="$3"
  rm -rf "$work/runner" "$work/env" "$work/path"; mkdir -p "$work/runner"
  : > "$work/env"; : > "$work/path"
  "$py" -X utf8 "$here/extract-run.py" "$root/.github/workflows/$wf" "Install the glassops command" > "$work/step.sh" || { bad "$wf: no install step"; return 1; }
  env -i PATH="$work/bin:$PATH" HOME="$work" \
    RUNNER_TEMP="$work/runner" GITHUB_ENV="$work/env" GITHUB_PATH="$work/path" \
    GITHUB_SERVER_URL=https://github.com INPUT_IMAGE="$input" WORKFLOW_SHA="$sha" \
    FAKE_LS_EXIT="${FAKE_LS_EXIT:-0}" FAKE_LS_REMOTE="$tags" FAKE_GIT_URL="$work/url" FAKE_DOCKER_LOG="$work/docker.log" \
    bash -eo pipefail "$work/step.sh" > "$work/stdout" 2>&1
  code=$?
  image="$(sed -n 's/^GLASSOPS_IMAGE=//p' "$work/env")"
}

# Annotated tags list the tag object and then the commit with ^{}. Lightweight tags list the commit.
printf '%s\n' \
  "$sha	refs/tags/v1" \
  "$sha	refs/tags/v1.2.3" \
  "$other	refs/tags/v1.2.2" \
  "$other	refs/tags/v0.4.0" > "$work/light"
printf '%s\n' \
  "aaaa000000000000000000000000000000000000	refs/tags/v1" \
  "$sha	refs/tags/v1^{}" \
  "bbbb000000000000000000000000000000000000	refs/tags/v1.2.3" \
  "$sha	refs/tags/v1.2.3^{}" > "$work/annotated"
printf '%s\n' \
  "$sha	refs/tags/v2.0.0-rc1" \
  "$sha	refs/tags/v2.0.0" \
  "$sha	refs/tags/v2" > "$work/final-over-rc"
printf '%s\n' "$sha	refs/tags/v2.0.0-rc1" > "$work/rc-only"
printf '%s\n' "$sha	refs/tags/v1" "$sha	refs/tags/glassops/deployed/prod" > "$work/major-only"
printf '%s\n' "$other	refs/tags/v1.2.3" > "$work/elsewhere"
: > "$work/none"

for wf in glassops-pr.yml glassops-deploy.yml; do
  echo "== $wf"
  step "$wf" "" "$work/light"
  [ "$code" = 0 ] && [ "$image" = "ghcr.io/glassops-platform/glassops:v1.2.3" ] && ok "lightweight tags -> v1.2.3" || bad "lightweight: code=$code image=$image"
  [ "$(cat "$work/url")" = "https://github.com/glassops-platform/glassops.git" ] && ok "looks tags up in the public repo" || bad "ls-remote url: $(cat "$work/url")"

  step "$wf" "" "$work/annotated"
  [ "$code" = 0 ] && [ "$image" = "ghcr.io/glassops-platform/glassops:v1.2.3" ] && ok "annotated tags -> v1.2.3" || bad "annotated: code=$code image=$image"

  step "$wf" "" "$work/final-over-rc"
  [ "$image" = "ghcr.io/glassops-platform/glassops:v2.0.0" ] && ok "a final release wins over a prerelease on the same commit" || bad "final over rc: $image"

  step "$wf" "" "$work/rc-only"
  [ "$image" = "ghcr.io/glassops-platform/glassops:v2.0.0-rc1" ] && ok "a prerelease tag is used when it is the only one" || bad "rc only: $image"

  for case in major-only elsewhere none; do
    step "$wf" "" "$work/$case"
    if [ "$code" != 0 ] && [ -z "$image" ] && grep -q "::error::" "$work/stdout"; then ok "$case -> error, no image"; else bad "$case: code=$code image=$image"; fi
  done

  FAKE_LS_EXIT=1 step "$wf" "" "$work/light"
  [ "$code" != 0 ] && [ -z "$image" ] && ok "git failure -> error, no image" || bad "git failure: code=$code image=$image"

  step "$wf" "ghcr.io/glassops-platform/glassops@sha256:abc" "$work/none"
  [ "$code" = 0 ] && [ "$image" = "ghcr.io/glassops-platform/glassops@sha256:abc" ] && ok "the image input wins and needs no lookup" || bad "input: code=$code image=$image"

  # The wrapper.
  step "$wf" "example.com/g:1" "$work/none"
  wrapper="$work/runner/glassops-bin/glassops"
  [ -x "$wrapper" ] && ok "writes an executable wrapper" || { bad "no wrapper"; continue; }
  if grep -qF '\n' "$wrapper"; then bad "wrapper holds a literal backslash-n"; else ok "no literal backslash-n in the wrapper"; fi
  mkdir -p "$work/ws"
  env -i PATH="$work/bin:$PATH" HOME="$work" GITHUB_WORKSPACE="$work/ws" GITHUB_OUTPUT="$work/o" \
    GLASSOPS_IMAGE="example.com/g:1" GLASSOPS_PASS_ENV="GLASSOPS_PERMIT_KEY EXTRA_NAME" \
    FAKE_DOCKER_LOG="$work/docker.log" bash "$wrapper" report --x > /dev/null 2>&1
  for n in GITHUB_ACTIONS GITHUB_REPOSITORY GITHUB_SHA GITHUB_REF GITHUB_EVENT_NAME GITHUB_ACTOR GITHUB_TRIGGERING_ACTOR \
    GITHUB_RUN_ID GITHUB_BASE_REF GITHUB_API_URL GITHUB_SERVER_URL GITHUB_TOKEN SF_JWT_KEY SF_CLIENT_ID SF_USERNAME \
    SF_DEVHUB_ORG_ID SF_PROD_USERNAME SF_PROD_ORG_ID GLASSOPS_PERMIT_KEY EXTRA_NAME GITHUB_OUTPUT; do
    awk -v v="$n" 'prev=="-e" && $0==v {f=1} {prev=$0} END {exit !f}' "$work/docker.log" || bad "wrapper does not pass $n"
  done
  stray="$(awk 'prev=="-e" && $0 !~ /^(GITHUB_|SF_|GLASSOPS_PERMIT_KEY$|EXTRA_NAME$)/ {print} {prev=$0}' "$work/docker.log")"
  [ -z "$stray" ] && ok "wrapper passes the expected names and no stray one" || bad "stray -e: $stray"
  tail -3 "$work/docker.log" | tr '\n' ' ' | grep -q "example.com/g:1 report --x" && ok "wrapper runs the image with the arguments" || bad "wrapper tail: $(tail -3 "$work/docker.log" | tr '\n' ' ')"
  grep -q "^GITHUB_STEP_SUMMARY$" "$work/docker.log" && bad "passes GITHUB_STEP_SUMMARY that was not set" || ok "unset CI files are not mounted"
done

[ "$fail" = 0 ] && echo "ok: reusable workflow install step" || echo "FAILED"
exit "$fail"
