#!/bin/sh
# The OPA end-to-end run. It gates an exec project with the Rego files in examples/rego and a
# rule over the changed paths, and checks that
#
#   sh scripts/e2e/opa-e2e.sh [glassops] [source tree]
#
#   1. a plan that touches only unprotected paths passes (exit 0, a passing "rego" check)
#   2. a plan that touches a protected path is denied (exit 1, a failed rego.protected_path
#      check in the PERMIT, and the PERMIT does not allow)
#   3. the same plan gated with no Rego files passes, so the denial is the Rego rule's
#
# examples/rego/deploy.rego looks at Salesforce facts, which an exec project has none of, so it
# loads and stays quiet here. The rule that fires is testdata/e2e/rego/protected-paths.rego.
# The source tree is the checkout, by default two directories above this script. Needs sh,
# git, base64, head, node and opa on the PATH, which the glassops image has. No network.
set -eu

# shellcheck source=/dev/null
. "$(dirname "$0")/lib-exec-project.sh"
e2e_init "${1:-glassops}"
SRC=${2:-$(cd "$(dirname "$0")/../.." && pwd)}
for f in examples/rego/deploy.rego examples/rego/deploy_test.rego testdata/e2e/rego/protected-paths.rego; do
  [ -f "$SRC/$f" ] || { echo "opa e2e: $SRC/$f is missing" >&2; exit 2; }
done

# The examples carry their own tests. This proves the copy in this checkout is the one tested.
if opa test "$SRC/examples/rego/deploy.rego" "$SRC/examples/rego/deploy_test.rego" > "$WORK/opa-test.txt" 2>&1; then
  pass "opa test examples/rego"
else
  fail "opa test examples/rego"
  cat "$WORK/opa-test.txt" >&2
fi

# The command is single quoted on purpose: exec.plan expands EXEC_E2E_SCRIPTS itself.
# shellcheck disable=SC2016,SC2089
E2E_PLAN_COMMAND='sh "$EXEC_E2E_SCRIPTS/exec-plan.sh"'
# shellcheck disable=SC2089
E2E_EXTRA_YAML="policy:
  rego:
    files: [policy/deploy.rego, policy/protected-paths.rego]"
# shellcheck disable=SC2090
export E2E_EXTRA_YAML E2E_PLAN_COMMAND
e2e_project
mkdir -p "$REPO/policy"
cp "$SRC/examples/rego/deploy.rego" "$SRC/testdata/e2e/rego/protected-paths.rego" "$REPO/policy/"
git -C "$REPO" add policy
git -C "$REPO" commit -q -m policy

# rego_outcome prints the outcome of the PERMIT check with the given rule, or "none".
rego_outcome() {
  node -e 'const p = JSON.parse(require("fs").readFileSync(process.argv[1], "utf8")); const c = p.checks.find((x) => x.rule === process.argv[2]); console.log(c ? c.outcome : "none")' "$REPO/.glassops/permit.json" "$1"
}

# 1. The last commit (policy) touches no path under app/, so the plan is empty of protected
# paths. Plan the second commit instead, which changes app/a.txt and app/b.txt.
expect_exit 0 "plan of an unprotected change" plan --from HEAD~2 --to HEAD~1
expect_exit 0 "gate allows an unprotected change" gate --target prod
if [ "$(rego_outcome rego)" = pass ]; then
  pass "the PERMIT holds a passing rego check"
else
  fail "the rego check is $(rego_outcome rego), want pass"
fi
if [ "$(rego_outcome rego.protected_path)" = none ]; then
  pass "no rego.protected_path check"
else
  fail "rego.protected_path is present in a plan with no protected path"
fi

# 2. A commit that changes a protected path.
mkdir -p "$REPO/app/protected"
echo secret > "$REPO/app/protected/key.txt"
git -C "$REPO" add app/protected
git -C "$REPO" commit -q -m protected
expect_exit 0 "plan of a protected change" plan --from HEAD~1 --to HEAD
expect_refusal "gate denies a protected path" gate --target prod
if [ "$(rego_outcome rego.protected_path)" = fail ]; then
  pass "the PERMIT holds a failed rego.protected_path check"
else
  fail "rego.protected_path is $(rego_outcome rego.protected_path), want fail"
fi
if [ "$(node -e 'console.log(JSON.parse(require("fs").readFileSync(process.argv[1], "utf8")).allow)' "$REPO/.glassops/permit.json")" = false ]; then
  pass "the PERMIT does not allow"
else
  fail "the PERMIT allows a protected path"
fi

# 3. The same plan with the Rego files left out of the config.
grep -v '^policy:' "$REPO/glassops.yaml" | grep -v '^  rego:' | grep -v 'files: \[policy' > "$WORK/no-rego.yaml"
if grep -q rego "$WORK/no-rego.yaml"; then
  fail "the config without Rego still names Rego"
fi
expect_exit 0 "gate with no Rego files allows the same plan" gate --target prod --config "$WORK/no-rego.yaml"

e2e_done "opa e2e"
