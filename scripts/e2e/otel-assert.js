// Checks the file an OpenTelemetry Collector file exporter wrote for the OTel end-to-end run
// (scripts/e2e/otel-e2e.sh):
//
//   node scripts/e2e/otel-assert.js <traces.jsonl> <step> [<step> ...]
//
// It passes only when the file holds exactly one trace, one root span "glassops run", exactly
// one child span per step named, in order, each a child of the root, and no span or resource
// attribute outside the allowlist in internal/core/telemetry/telemetry.go. It also refuses a
// file that holds an email address or something shaped like a Salesforce org Id.
"use strict";

const fs = require("fs");

// The attribute names GlassOps may send, from the attr constants in telemetry.go, plus the
// one resource attribute. Keep this list in step with that file.
const ALLOWED = new Set([
  "glassops.pipeline",
  "glassops.provider",
  "glassops.exit_code",
  "glassops.empty_plan",
  "glassops.cancelled",
  "glassops.step",
  "glassops.step.index",
  "glassops.step.when",
  "glassops.step.outcome",
  "glassops.step.skipped",
  "glassops.duration_ms",
  "glassops.target.name",
  "glassops.target.tier",
  "service.name",
]);

const [file, ...steps] = process.argv.slice(2);
if (!file || steps.length === 0) {
  console.error("usage: node otel-assert.js <traces.jsonl> <step> [<step> ...]");
  process.exit(2);
}

const problems = [];
const text = fs.readFileSync(file, "utf8");
if (text.trim() === "") {
  console.error("FAIL: the collector wrote no spans to " + file);
  process.exit(1);
}

const spans = [];
const attrNames = new Set();
for (const line of text.split("\n")) {
  if (line.trim() === "") continue;
  const doc = JSON.parse(line);
  for (const rs of doc.resourceSpans || []) {
    for (const a of (rs.resource || {}).attributes || []) attrNames.add(a.key);
    for (const ss of rs.scopeSpans || []) {
      for (const sp of ss.spans || []) {
        spans.push(sp);
        for (const a of sp.attributes || []) attrNames.add(a.key);
      }
    }
  }
}

const traces = new Set(spans.map((s) => s.traceId));
if (traces.size !== 1) problems.push("want one trace, got " + traces.size);

const roots = spans.filter((s) => !s.parentSpanId);
if (roots.length !== 1 || roots[0].name !== "glassops run") {
  problems.push("want one root span named glassops run, got " + JSON.stringify(roots.map((s) => s.name)));
}
const children = spans.filter((s) => s.parentSpanId);
const got = children.map((s) => s.name);
const want = steps.map((s) => "glassops " + s);
if (JSON.stringify(got) !== JSON.stringify(want)) {
  problems.push("child spans are " + JSON.stringify(got) + ", want " + JSON.stringify(want));
}
if (roots.length === 1) {
  for (const c of children) {
    if (c.parentSpanId !== roots[0].spanId) problems.push("span " + c.name + " is not a child of the root");
  }
}

for (const name of attrNames) {
  if (!ALLOWED.has(name)) problems.push("attribute outside the allowlist: " + name);
}
if (/[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Za-z]{2,}/.test(text)) problems.push("the file holds an email address");
if (/\b00D[A-Za-z0-9]{12,15}\b/.test(text)) problems.push("the file holds something shaped like an org Id");

if (problems.length > 0) {
  for (const p of problems) console.error("FAIL: " + p);
  process.exit(1);
}
console.log("ok: one trace, one root and one span per step (" + steps.join(", ") + "), " +
  attrNames.size + " attribute names, all on the allowlist");
