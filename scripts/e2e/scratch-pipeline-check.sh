#!/bin/sh
# Checks what the e2e pipeline left in its results directory. Runs on the runner, needs node.
#
#   sh scratch-pipeline-check.sh RESULTS_DIR
#
# Exits 0 only when every result the pipeline must produce holds the expected facts: the
# minimal seed, org coverage of at least 75, verify passed on the deploy that deploy made, and
# the run itself exiting 0. It lists every problem, not only the first.
set -eu

DIR=${1:?usage: scratch-pipeline-check.sh RESULTS_DIR}

node -e '
const fs = require("fs"), dir = process.argv[1];
const read = n => {
  try { return JSON.parse(fs.readFileSync(dir + "/" + n + ".json", "utf8")); }
  catch (e) { return null; }
};
const problems = [];
const need = n => {
  const r = read(n);
  if (!r) { problems.push(n + ".json is missing or unreadable"); return {}; }
  return r;
};
const facts = r => r.facts || {};
const check = (r, name) => (r.checks || []).find(c => c.name === name);

const v = need("validate"), d = need("deploy"), vf = need("verify"), t = need("test"), r = need("run");
const vfacts = facts(v), dfacts = facts(d), vffacts = facts(vf), tfacts = facts(t);

if (vfacts["validate.seed.mode"] !== "minimal") problems.push("seed mode " + vfacts["validate.seed.mode"]);
if (!JSON.stringify(vfacts["validate.seed.members"] || []).includes("Account.Tier__c")) problems.push("seed members " + JSON.stringify(vfacts["validate.seed.members"]));

if (d.status !== "pass" || dfacts["deploy.quick"] !== true || !dfacts["deploy.id"]) problems.push("deploy did not quick deploy: status " + d.status + ", quick " + dfacts["deploy.quick"] + ", id " + dfacts["deploy.id"]);

if (vf.status !== "pass") problems.push("verify status " + vf.status);
if (vffacts["verify.passed"] !== true) problems.push("verify.passed is " + vffacts["verify.passed"]);
if (vffacts["verify.deploy_id"] !== dfacts["deploy.id"]) problems.push("verify checked deploy " + vffacts["verify.deploy_id"] + ", deploy made " + dfacts["deploy.id"]);
if (vffacts["verify.check_only"] !== false) problems.push("verify.check_only is " + vffacts["verify.check_only"]);
for (const c of vf.checks || []) {
  if (c.status !== "pass" && c.status !== "skipped") problems.push("verify check " + c.name + " is " + c.status + ": " + c.detail);
}

if (!(tfacts["test.org_coverage"] >= 75)) problems.push("org coverage " + tfacts["test.org_coverage"]);
if (facts(r)["run.exit_code"] !== 0) problems.push("run exit " + facts(r)["run.exit_code"]);

if (problems.length) {
  console.error("unexpected pipeline results:\n  - " + problems.join("\n  - "));
  process.exit(1);
}
console.log("run e2e: settings, minimal seed, validate, deploy " + dfacts["deploy.id"] + ", verify passed, tests at " + tfacts["test.org_coverage"] + "% org coverage, exec step, report, env delete");
' "$DIR"
