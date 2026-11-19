# Telemetry fixtures

`collector-file-export.json` is what `otel/opentelemetry-collector` 0.140.0 wrote with an OTLP HTTP
receiver and a file exporter after `glassops run probe`, a three step exec pipeline whose second
step fails. It holds one trace: the `glassops run` root span with status code 2 and exit code 1,
and one `glassops exec` child per step, the failing one with status code 2. The attributes are the
allowlist in `telemetry.go`, and the resource carries only `service.name`. The collector parsed the
export without a warning, so the hand-written OTLP JSON is accepted as sent.
