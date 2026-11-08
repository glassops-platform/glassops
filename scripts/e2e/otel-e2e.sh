#!/bin/sh
# The OTel end-to-end run, the glassops side. It builds an exec project with a pipeline of
# three steps and runs it with glassops run, exporting the run as trace spans to the OTLP
# endpoint in E2E_OTLP_ENDPOINT (default http://otel-collector:4318). The workflow starts an
# OpenTelemetry Collector with a file exporter there, then checks the file with
# scripts/e2e/otel-assert.js. This script checks only that the run itself passed.
#
#   E2E_OTLP_ENDPOINT=http://otel-collector:4318 sh scripts/e2e/otel-e2e.sh [glassops]
#
# Needs sh, git, base64, head and node on the PATH, which the glassops image has, and a
# network that reaches the endpoint. No secret and no Salesforce org.
set -eu

# shellcheck source=/dev/null
. "$(dirname "$0")/lib-exec-project.sh"
e2e_init "${1:-glassops}"

# shellcheck disable=SC2089
E2E_EXTRA_YAML="telemetry:
  otlp:
    endpoint: ${E2E_OTLP_ENDPOINT:-http://otel-collector:4318}
    service_name: glassops-e2e
pipelines:
  otel:
    - plan --from HEAD~1 --to HEAD
    - gate --target prod
    - step: report
      when: always"
# shellcheck disable=SC2090
export E2E_EXTRA_YAML
e2e_project

expect_exit 0 "glassops run otel" run otel
if grep -qi "telemetry" "$WORK/out.txt" && grep -i "telemetry" "$WORK/out.txt" | grep -qi "warn\|fail\|error"; then
  fail "the run reported a telemetry problem"
  cat "$WORK/out.txt" >&2
fi

e2e_done "otel e2e"
