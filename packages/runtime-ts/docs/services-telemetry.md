---
type: Documentation
domain: runtime-ts
origin: packages/runtime-ts/src/services/telemetry.ts
last_modified: 2026-01-29
generated: true
source: packages/runtime-ts/src/services/telemetry.ts
generated_at: 2026-01-29T21:02:05.205324
hash: effd62924f7ca6ef31b43af94c93a643466e7be582e31b7b472180bb6b2f408e
---

## GlassOps Runtime Telemetry Service Documentation

This document details the telemetry service integrated within GlassOps Runtime, providing tracing and metrics capabilities for governance operations. It outlines initialization, usage, and shutdown procedures.

**Overview**

The telemetry service leverages the OpenTelemetry standard for instrumenting applications. This allows for observability and performance analysis of GlassOps Runtime components. Data is exported to a configured OpenTelemetry Collector endpoint.  The service is designed to be non-intrusive; it will not impede functionality if configuration is incomplete or initialization fails.

**Initialization**

The `initTelemetry` function establishes the OpenTelemetry SDK. 

*   **Prerequisites:**  The environment variable `OTEL_EXPORTER_OTLP_ENDPOINT` must be set to a valid OpenTelemetry Collector endpoint URL (e.g., `http://localhost:4317`).
*   **Configuration:**
    *   `serviceName` (optional):  A string representing the name of the service. Defaults to “glassops-runtime”.
    *   `serviceVersion` (optional): A string representing the version of the service. Defaults to “1.0.0”.
    *   `OTEL_EXPORTER_OTLP_HEADERS` (optional): A comma-separated list of headers to include with trace exports (e.g., `Header=Value,Header2=Value2`).
*   **Behavior:** If `OTEL_EXPORTER_OTLP_ENDPOINT` is not set, the telemetry service remains disabled. Initialization failures are logged as warnings and do not halt application execution.

**Usage**

The service provides several functions for interacting with OpenTelemetry:

*   **`withSpan(name, fn, attributes)`:**  Executes an asynchronous function within a traced span.
    *   `name`: A string representing the name of the span.
    *   `fn`: An asynchronous function that accepts a span object as an argument. This is where you place the code you want to trace.
    *   `attributes` (optional): A record of key-value pairs to attach to the span as attributes. These provide additional context for the trace.
    *   This function automatically handles span start, attribute setting, error handling (recording exceptions and setting span status), and span completion.
*   **`getCurrentSpan()`:** Returns the currently active span, if one exists. This allows you to manually interact with the current span outside of the `withSpan` context.
*   **`addSpanEvent(name, attributes)`:** Adds an event to the current active span.
    *   `name`: A string representing the name of the event.
    *   `attributes` (optional): A record of key-value pairs to attach to the event as attributes.

**Shutdown**

The `shutdown` function gracefully shuts down the OpenTelemetry SDK. You should call this function before your application exits to ensure all pending data is exported.

*   **Behavior:**  Any errors during shutdown are ignored to prevent application termination issues.

**Important Considerations**

*   Ensure your OpenTelemetry Collector is properly configured to receive and process trace data.
*   Use descriptive span names and relevant attributes to create meaningful traces.
*   Consider the performance impact of adding telemetry, especially in high-throughput scenarios.
*   You can disable telemetry entirely by not setting the `OTEL_EXPORTER_OTLP_ENDPOINT` environment variable.