#!/bin/sh
# The exec.plan command of the Rego end-to-end run. glassops runs it after it has written the
# plan files, with GLASSOPS_PLAN_DIR set. It reports the changed files as the fact
# exec.changed_files, so a Rego rule can match a path. It talks to no outside system.
set -eu

: "${GLASSOPS_PLAN_DIR:?GLASSOPS_PLAN_DIR is not set}"
: "${GLASSOPS_RESULTS_DIR:?GLASSOPS_RESULTS_DIR is not set}"

node -e '
const fs = require("fs");
const dir = process.argv[1];
const files = fs.readFileSync(dir + "/changes.txt", "utf8").split("\n").filter(Boolean);
const result = {
  schema: "glassops.engine-result/v1",
  step: "plan",
  passed: true,
  facts: { "exec.changed_files": files },
};
fs.writeFileSync(process.argv[2] + "/exec-plan.json", JSON.stringify(result) + "\n");
' "$GLASSOPS_PLAN_DIR" "$GLASSOPS_RESULTS_DIR"
