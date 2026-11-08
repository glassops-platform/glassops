#!/usr/bin/env bash
# Runs the script of action.yml with docker stubbed, and checks the docker command it builds.
#   scripts/release/test-action.sh
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
mkdir -p "$work/bin" "$work/ws"
"$py" -X utf8 "$here/extract-run.py" "$root/action.yml" run > "$work/action.sh" || exit 2

cat > "$work/bin/docker" <<'STUB'
#!/usr/bin/env bash
printf '%s\n' "$@" > "$FAKE_DOCKER_LOG"
exit "${FAKE_DOCKER_EXIT:-0}"
STUB
chmod +x "$work/bin/docker"

fail=0
ok() { echo "ok   $1"; }
bad() { echo "FAIL $1"; fail=1; }

# run_action <exit> VAR=value... : runs the script in a clean environment.
run_action() {
  local want_exit="$1"; shift
  : > "$work/out"; rm -f "$work/docker.log"
  env -i PATH="$work/bin:$PATH" HOME="$work" \
    GITHUB_WORKSPACE="$work/ws" GITHUB_OUTPUT="$work/out" GITHUB_STEP_SUMMARY="$work/summary" \
    FAKE_DOCKER_LOG="$work/docker.log" FAKE_DOCKER_EXIT="$want_exit" \
    INPUT_COMMAND="run pr" INPUT_IMAGE="" INPUT_ENV="" INPUT_TOKEN="" INPUT_WORKDIR="." ACTION_REF="v1" \
    "$@" bash "$work/action.sh" > "$work/stdout" 2>&1
  echo $? > "$work/code"
}
has() { grep -qxF -- "$1" "$work/docker.log"; }
has_pair() { # flag value: the value on the line after the flag
  awk -v f="$1" -v v="$2" 'prev==f && $0==v {found=1} {prev=$0} END {exit !found}' "$work/docker.log"
}

# 1. The names glassops reads are each passed as -e NAME, and nothing else is.
run_action 0
for n in GITHUB_ACTIONS GITHUB_REPOSITORY GITHUB_REF GITHUB_SHA GITHUB_EVENT_NAME GITHUB_ACTOR GITHUB_RUN_ID \
  GITHUB_BASE_REF GITHUB_TRIGGERING_ACTOR GITHUB_API_URL GITHUB_SERVER_URL; do
  has_pair -e "$n" && ok "passes $n" || bad "does not pass $n"
done
stray="$(awk 'prev=="-e" && $0 !~ /^GITHUB_(ACTIONS|REPOSITORY|REF|SHA|EVENT_NAME|ACTOR|RUN_ID|BASE_REF|TRIGGERING_ACTOR|API_URL|SERVER_URL|OUTPUT|STEP_SUMMARY)$/ {print} {prev=$0}' "$work/docker.log")"
[ -z "$stray" ] && ok "no stray -e name" || bad "stray -e name: $stray"
if grep -qF '\n' "$work/action.sh"; then bad "action.yml script holds a literal backslash-n"; else ok "no literal backslash-n in the script"; fi

# 2. Mounts, working directory and the command.
has_pair -v "$work/ws:/work" && ok "mounts the workspace" || bad "workspace mount"
has_pair -w "/work/." && ok "working directory" || bad "working directory"
has_pair -e GITHUB_OUTPUT && has_pair -v "$work/out:$work/out" && ok "mounts GITHUB_OUTPUT at its runner path" || bad "GITHUB_OUTPUT mount"
tail -2 "$work/docker.log" | tr '\n' ' ' | grep -q "run pr" && ok "ends with the command" || bad "command order"

# 3. Image choice from the action ref.
image_for() { run_action 0 "ACTION_REF=$1" ${2:+"INPUT_IMAGE=$2"}; grep -E '^ghcr.io|^[a-z./-]+:[A-Za-z0-9.-]+$' "$work/docker.log" | head -1; }
[ "$(image_for v1)" = "ghcr.io/glassops-platform/glassops:v1" ] && ok "v1 -> :v1" || bad "v1 -> $(image_for v1)"
[ "$(image_for v1.2.3)" = "ghcr.io/glassops-platform/glassops:v1.2.3" ] && ok "v1.2.3 -> :v1.2.3" || bad "v1.2.3"
[ "$(image_for v1.0.0-rc1)" = "ghcr.io/glassops-platform/glassops:v1.0.0-rc1" ] && ok "prerelease ref keeps its tag" || bad "prerelease ref"
[ "$(image_for main)" = "ghcr.io/glassops-platform/glassops:latest" ] && ok "branch -> :latest" || bad "branch"
[ "$(image_for 0123456789abcdef0123456789abcdef01234567)" = "ghcr.io/glassops-platform/glassops:latest" ] && ok "sha -> :latest" || bad "sha"
[ "$(image_for v1 example.com/x/y:z)" = "example.com/x/y:z" ] && ok "explicit image wins" || bad "explicit image"

# 4. The token, extra env names and the exit code.
run_action 0 INPUT_TOKEN=secret-value "INPUT_ENV=$(printf 'SF_JWT_KEY\n  SF_CLIENT_ID  \n\n')"
has_pair -e GITHUB_TOKEN && ok "passes GITHUB_TOKEN by name" || bad "no GITHUB_TOKEN"
has_pair -e SF_JWT_KEY && has_pair -e SF_CLIENT_ID && ok "passes the env input names" || bad "env input names"
grep -q "secret-value" "$work/docker.log" && bad "the token value reached the docker command line" || ok "token value is not on the command line"
run_action 0
has_pair -e GITHUB_TOKEN && bad "GITHUB_TOKEN passed without a token" || ok "no token, none passed"
run_action 1
[ "$(cat "$work/code")" = 1 ] && grep -qx 'exit-code=1' "$work/out" && ok "exit code 1 propagates and is an output" || bad "exit code 1"
run_action 2
[ "$(cat "$work/code")" = 2 ] && grep -qx 'exit-code=2' "$work/out" && ok "exit code 2 propagates" || bad "exit code 2"
run_action 0 "INPUT_COMMAND=gate --target prod"
tail -3 "$work/docker.log" | tr '\n' ' ' | grep -q "gate --target prod" && ok "splits the command on spaces" || bad "command split"

[ "$fail" = 0 ] && echo "ok: action.yml script" || echo "FAILED"
exit "$fail"
