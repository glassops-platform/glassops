#!/bin/sh
# Shared helpers for the Phase 6 end-to-end scripts (sigstore-e2e.sh, opa-e2e.sh,
# otel-e2e.sh, pr-comment-e2e.sh). Source it, do not run it:
#
#   . "$(dirname "$0")/lib-exec-project.sh"
#
# It builds the same exec project shape as exec-e2e.sh: a temp git repository with two
# commits and a target that is a plain directory. No Salesforce org and no secret is used.
# The caller sets E2E_EXTRA_YAML (top level YAML appended to glassops.yaml) before
# e2e_project, and E2E_PLAN_COMMAND (an exec.plan command, default none).
#
# shellcheck shell=sh

# e2e_init sets the scratch directory and the counters. GLASSOPS is the binary to run.
e2e_init() {
  GLASSOPS=${1:-glassops}
  SCRIPTS=$(cd "$(dirname "$0")" && pwd)
  WORK=$(mktemp -d)
  REPO=$WORK/repo
  TARGET=$WORK/target
  mkdir -p "$REPO" "$TARGET/app"
  trap 'rm -rf "$WORK"' EXIT INT TERM
  EXEC_E2E_SCRIPTS=$SCRIPTS
  EXEC_E2E_TARGET_DIR=$TARGET
  EXEC_E2E_PROD_ID=exec-e2e-prod
  GLASSOPS_PERMIT_KEY=$(head -c 32 /dev/urandom | base64 | tr -d '\n')
  export EXEC_E2E_SCRIPTS EXEC_E2E_TARGET_DIR EXEC_E2E_PROD_ID GLASSOPS_PERMIT_KEY
  echo "$EXEC_E2E_PROD_ID" > "$TARGET/.target-id"
  failures=0
}

fail() {
  echo "FAIL: $*" >&2
  failures=$((failures + 1))
}

pass() {
  echo "ok: $*"
}

# e2e_done prints the verdict and exits.
e2e_done() {
  if [ "$failures" -ne 0 ]; then
    echo "$1: $failures failed" >&2
    exit 1
  fi
  echo "$1: all passed"
}

# gops runs a glassops command against the repository with the local CI layer. Flags go
# before any -- so a command after it gets nothing appended. It is not named glassops, so a
# GLASSOPS value of glassops runs the binary on the PATH and not this function.
gops() {
  sub=$1
  shift
  if [ "$sub" = permit ]; then
    sub="permit $1"
    shift
  fi
  # shellcheck disable=SC2086
  "$GLASSOPS" $sub --dir "$REPO" --ci local "$@"
}

# expect_exit runs a glassops command and checks its exit code. The output stays in
# $WORK/out.txt.
expect_exit() {
  want=$1
  what=$2
  shift 2
  set +e
  gops "$@" > "$WORK/out.txt" 2>&1
  got=$?
  set -e
  if [ "$got" -eq "$want" ]; then
    pass "$what (exit $got)"
  else
    fail "$what: exit $got, want $want"
    cat "$WORK/out.txt" >&2
  fi
}

# expect_refusal runs a glassops command that must exit 1, a refusal. Exit 2 is a tool or
# infrastructure error, which this script does not accept as a refusal, so an outage cannot
# pass for a check that worked.
expect_refusal() {
  expect_exit 1 "$@"
}

# permit_json prints one expression over the parsed PERMIT as JSON. The expression sees the
# PERMIT as p.
permit_json() {
  node -e 'const p = JSON.parse(require("fs").readFileSync(process.argv[1], "utf8")); console.log(JSON.stringify(eval(process.argv[2])))' "$REPO/.glassops/permit.json" "$1"
}

# e2e_project builds the repository and glassops.yaml. The second commit changes app/a.txt,
# adds app/b.txt, deletes app/old.txt and changes other/c.txt, which is outside exec.paths.
# E2E_EXTRA_YAML is appended to the config. E2E_PLAN_COMMAND, when set, is the exec.plan
# command.
e2e_project() {
  echo old > "$TARGET/app/old.txt"
  cd "$REPO"
  git init -q -b main
  git config user.email e2e@example.invalid
  git config user.name e2e
  git config core.autocrlf false
  mkdir -p app other
  echo one > app/a.txt
  echo old > app/old.txt
  echo one > other/c.txt
  echo .glassops/ > .gitignore
  {
    cat <<'EOF'
version: 1
provider: exec
exec:
  paths: [app]
  describe: sh "$EXEC_E2E_SCRIPTS/exec-describe.sh"
  deploy: "false"
EOF
    if [ -n "${E2E_PLAN_COMMAND:-}" ]; then
      printf '  plan: %s\n' "$E2E_PLAN_COMMAND"
    fi
    cat <<'EOF'
targets:
  prod:
    id: $EXEC_E2E_PROD_ID
    tier: prod
EOF
    if [ -n "${E2E_EXTRA_YAML:-}" ]; then
      printf '%s\n' "$E2E_EXTRA_YAML"
    fi
  } > glassops.yaml
  git add -A
  git commit -q -m one
  echo two > app/a.txt
  echo new > app/b.txt
  echo two > other/c.txt
  git rm -q app/old.txt
  git add -A
  git commit -q -m two
  cd - > /dev/null
}
