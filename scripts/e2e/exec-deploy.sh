#!/bin/sh
# The deploy command the exec provider end-to-end run deploys with:
#   glassops deploy --target T -- sh scripts/e2e/exec-deploy.sh
#
# Its target is a plain directory, EXEC_E2E_TARGET_DIR. It copies every file in the plan's
# changes.txt into that directory, removes every file in deleted.txt, and writes
# exec-deploy.json in the shape of schemas/engine-result.json. It talks to no outside system.
#
# It exits 2 before it touches the target when a GLASSOPS_* variable glassops must set is
# missing or names no file, or when the PERMIT key reached it, because a deploy command that
# can read the key could sign its own PERMIT.
set -eu

fail() {
  echo "exec-deploy: $*" >&2
  exit 2
}

for name in GLASSOPS_PERMIT GLASSOPS_PLAN GLASSOPS_PLAN_DIR GLASSOPS_TARGET GLASSOPS_TARGET_ID GLASSOPS_RESULTS_DIR; do
  eval "value=\${$name:-}"
  [ -n "$value" ] || fail "$name is not set"
done
[ -z "${GLASSOPS_PERMIT_KEY+x}" ] || fail "GLASSOPS_PERMIT_KEY reached the deploy command"
[ -f "$GLASSOPS_PERMIT" ] || fail "GLASSOPS_PERMIT names no file"
[ -f "$GLASSOPS_PLAN" ] || fail "GLASSOPS_PLAN names no file"
[ -f "$GLASSOPS_PLAN_DIR/changes.txt" ] || fail "GLASSOPS_PLAN_DIR holds no changes.txt"
[ -d "$GLASSOPS_RESULTS_DIR" ] || fail "GLASSOPS_RESULTS_DIR is not a directory"
: "${EXEC_E2E_TARGET_DIR:?EXEC_E2E_TARGET_DIR is not set}"
[ -d "$EXEC_E2E_TARGET_DIR" ] || fail "EXEC_E2E_TARGET_DIR is not a directory"

count=0
while IFS= read -r file; do
  [ -n "$file" ] || continue
  mkdir -p "$EXEC_E2E_TARGET_DIR/$(dirname "$file")"
  cp "$file" "$EXEC_E2E_TARGET_DIR/$file"
  count=$((count + 1))
done < "$GLASSOPS_PLAN_DIR/changes.txt"

if [ -f "$GLASSOPS_PLAN_DIR/deleted.txt" ]; then
  while IFS= read -r file; do
    [ -n "$file" ] || continue
    rm -f "$EXEC_E2E_TARGET_DIR/$file"
    count=$((count + 1))
  done < "$GLASSOPS_PLAN_DIR/deleted.txt"
fi

# The target name and Id come from glassops.yaml. The run's own config keeps them to letters,
# digits, dots, dashes and underscores, so they need no JSON escaping here.
cat > "$GLASSOPS_RESULTS_DIR/exec-deploy.json" <<EOF
{
  "schema": "glassops.engine-result/v1",
  "step": "deploy",
  "passed": true,
  "id": "exec-e2e-$GLASSOPS_TARGET_ID",
  "components_deployed": $count,
  "facts": {"exec.e2e_target": "$GLASSOPS_TARGET"}
}
EOF
echo "exec-deploy: $count files deployed to target $GLASSOPS_TARGET"
