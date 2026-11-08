#!/bin/sh
# The exec provider end-to-end run. It needs no outside system: the target is a temp directory.
#
#   sh scripts/e2e/exec-e2e.sh [path to glassops]
#
# It builds a git repository with two commits, then runs the real binary through plan, gate
# with a throwaway PERMIT key and --deploy-command, permit verify, a tampered PERMIT, a target
# whose describe Id is wrong, a deploy command the PERMIT does not bind, a modified tracked
# file, and deploy -- sh scripts/e2e/exec-deploy.sh. It exits 0 only when every refusal
# refused, the deploy changed the target as the plan says, and the result records it. Needs
# sh, git, base64, head and node on the PATH, which the glassops image has.
set -eu

GLASSOPS=${1:-glassops}
SCRIPTS=$(cd "$(dirname "$0")" && pwd)
WORK=$(mktemp -d)
trap 'rm -rf "$WORK"' EXIT INT TERM
REPO=$WORK/repo
TARGET=$WORK/target
mkdir -p "$REPO" "$TARGET/app"

EXEC_E2E_SCRIPTS=$SCRIPTS
EXEC_E2E_TARGET_DIR=$TARGET
EXEC_E2E_PROD_ID=exec-e2e-prod
GLASSOPS_PERMIT_KEY=$(head -c 32 /dev/urandom | base64 | tr -d '\n')
export EXEC_E2E_SCRIPTS EXEC_E2E_TARGET_DIR EXEC_E2E_PROD_ID GLASSOPS_PERMIT_KEY
echo "$EXEC_E2E_PROD_ID" > "$TARGET/.target-id"
echo old > "$TARGET/app/old.txt"

failures=0
fail() {
  echo "FAIL: $*" >&2
  failures=$((failures + 1))
}
pass() {
  echo "ok: $*"
}

# gops runs a glassops command against the repository with the local CI layer. It is not
# named glassops, so GLASSOPS=glassops runs the binary on the PATH and not this function,
# which would call itself forever. Flags go before any -- so a command after it gets nothing
# appended.
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

# expect runs a glassops command and checks its exit code.
expect() {
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

# expect_refused runs a glassops command that must fail and leave the target untouched.
expect_refused() {
  what=$1
  shift
  set +e
  gops "$@" > "$WORK/out.txt" 2>&1
  got=$?
  set -e
  if [ "$got" -eq 0 ]; then
    fail "$what: exit 0, want a refusal"
    cat "$WORK/out.txt" >&2
  else
    pass "$what refused (exit $got)"
  fi
  if [ -f "$TARGET/app/a.txt" ] || [ ! -f "$TARGET/app/old.txt" ]; then
    fail "$what: the target changed"
  fi
}

# fact prints one fact of a result file as JSON.
fact() {
  node -e 'const r = JSON.parse(require("fs").readFileSync(process.argv[1], "utf8")); console.log(JSON.stringify((r.facts || {})[process.argv[2]]))' "$1" "$2"
}

check_fact() {
  got=$(fact "$1" "$2")
  if [ "$got" = "$3" ]; then
    pass "$(basename "$1") $2 = $got"
  else
    fail "$(basename "$1") $2 = $got, want $3"
  fi
}

# The repository: the second commit changes app/a.txt, adds app/b.txt, deletes app/old.txt
# and changes other/c.txt, which is outside exec.paths. exec.deploy is false, so only the
# command after -- can deploy.
cd "$REPO"
git init -q -b main
git config user.email exec-e2e@example.invalid
git config user.name exec-e2e
git config core.autocrlf false
mkdir -p app other
echo one > app/a.txt
echo old > app/old.txt
echo one > other/c.txt
echo .glassops/ > .gitignore
cat > glassops.yaml <<'EOF'
version: 1
provider: exec
exec:
  paths: [app]
  describe: sh "$EXEC_E2E_SCRIPTS/exec-describe.sh"
  deploy: "false"
targets:
  prod:
    id: $EXEC_E2E_PROD_ID
    tier: prod
EOF
git add -A
git commit -q -m one
echo two > app/a.txt
echo new > app/b.txt
echo two > other/c.txt
git rm -q app/old.txt
git add -A
git commit -q -m two
cd - > /dev/null

expect 0 "plan" plan --from HEAD~1 --to HEAD
if [ "$(cat "$REPO/.glassops/plan/changes.txt" 2>/dev/null)" = "app/a.txt
app/b.txt" ] && [ "$(cat "$REPO/.glassops/plan/deleted.txt" 2>/dev/null)" = "app/old.txt" ]; then
  pass "plan lists the changed and deleted files under exec.paths only"
else
  fail "plan files are wrong"
fi

# The PERMIT binds the sha256 of this exact argv, as a JSON array.
DEPLOY_COMMAND="[\"sh\",\"$SCRIPTS/exec-deploy.sh\"]"
expect 0 "gate runs exec.describe and the guard, and allows" gate --target prod --deploy-command "$DEPLOY_COMMAND"
if grep -q "exec-deploy.sh" "$REPO/.glassops/permit.json"; then
  fail "the PERMIT holds the deploy command text, not only its hash"
fi
if [ "$(node -e 'console.log(JSON.parse(require("fs").readFileSync(process.argv[1], "utf8")).allow)' "$REPO/.glassops/permit.json")" = true ]; then
  pass "the PERMIT allows"
else
  fail "the PERMIT does not allow"
fi
expect 0 "permit verify" permit verify --target prod

# A tampered PERMIT: the target kind is edited. deploy does not compare the kind with what it
# expects, so only the signature can refuse this edit.
cp "$REPO/.glassops/permit.json" "$WORK/permit.json"
sed 's/"kind": "production"/"kind": "sandbox"/' "$WORK/permit.json" > "$REPO/.glassops/permit.json"
if cmp -s "$WORK/permit.json" "$REPO/.glassops/permit.json"; then
  fail "the PERMIT tamper changed nothing"
fi
expect_refused "permit verify of a tampered PERMIT" permit verify --target prod
expect_refused "deploy under a tampered PERMIT" deploy --target prod -- sh "$SCRIPTS/exec-deploy.sh"
cp "$WORK/permit.json" "$REPO/.glassops/permit.json"

# A target that reports another Id.
echo exec-e2e-other > "$TARGET/.target-id"
expect_refused "deploy to a target whose describe Id is wrong" deploy --target prod -- sh "$SCRIPTS/exec-deploy.sh"
echo "$EXEC_E2E_PROD_ID" > "$TARGET/.target-id"

# A command the PERMIT does not bind, including the bound one with an extra argument.
expect_refused "deploy -- a command the PERMIT does not bind" deploy --target prod -- sh -c true
expect_refused "deploy -- the bound command with an extra argument" deploy --target prod -- sh "$SCRIPTS/exec-deploy.sh" extra

# A tracked file changed after gate: the tree is not the PERMIT's commit.
echo changed > "$REPO/app/a.txt"
expect_refused "deploy with a modified tracked file" deploy --target prod -- sh "$SCRIPTS/exec-deploy.sh"
git -C "$REPO" checkout -q -- app/a.txt

expect 0 "deploy -- sh scripts/e2e/exec-deploy.sh" deploy --target prod -- sh "$SCRIPTS/exec-deploy.sh"
if [ "$(cat "$TARGET/app/a.txt" 2>/dev/null)" = two ] && [ "$(cat "$TARGET/app/b.txt" 2>/dev/null)" = new ] && [ ! -e "$TARGET/app/old.txt" ] && [ ! -e "$TARGET/other" ]; then
  pass "the target holds the plan and nothing else"
else
  fail "the target does not match the plan"
fi
DEPLOY=$REPO/.glassops/results/deploy.json
check_fact "$DEPLOY" exec.result_file true
check_fact "$DEPLOY" exec.deploy_override true
check_fact "$DEPLOY" exec.e2e_target '"prod"'
check_fact "$DEPLOY" deploy.id '"exec-e2e-exec-e2e-prod"'

if [ "$failures" -ne 0 ]; then
  echo "exec e2e: $failures failed" >&2
  exit 1
fi
echo "exec e2e: all passed"
