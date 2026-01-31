---
type: Documentation
domain: runtime-ts
origin: packages/runtime-ts/src/services/telemetry.ts
last_modified: 2026-01-31
generated: true
source: packages/runtime-ts/src/services/telemetry.ts
generated_at: 2026-01-31T09:18:21.656583
hash: effd62924f7ca6ef31b43af94c93a643466e7be582e31b7b472180bb6b2f408e
---

## GlassOps Runtime Telemetry Service Documentation

This document details the telemetry service integrated within GlassOps Runtime, providing tracing and metrics capabilities for governance operations. It outlines initialization, usage, and shutdown procedures.

**Overview**

The telemetry service leverages the OpenTelemetry standard for instrumenting applications. This allows for observability and performance analysis of GlassOps Runtime components. Data is exported to a configured OpenTelemetry Collector endpoint.

**Initialization**

The `initTelemetry` function establishes the OpenTelemetry SDK. 

*   **Prerequisites:** To enable telemetry, the environment variable `OTEL_EXPORTER_OTLP_ENDPOINT` must be set to the URL of a compatible OpenTelemetry Collector.
*   **Configuration:**
    *   `serviceName` (optional):  A string representing the name of the service. Defaults to "glassops-runtime".
    *   `serviceVersion` (optional): A string representing the version of the service. Defaults to "1.0.0".
    *   `OTEL_EXPORTER_OTLP_HEADERS` (optional):  A comma-separated list of headers to include in the trace export request (e.g., "Header=Value,Header2=Value2").
*   **Behavior:** If `OTEL_EXPORTER_OTLP_ENDPOINT` is not defined, telemetry remains disabled, and the function completes without error. Initialization failures are logged as warnings and do not interrupt runtime operation.

**Usage: Tracing Operations with `withSpan`**

The `withSpan` function is the primary method for tracing asynchronous operations.

*   **Parameters:**
    *   `name`: A string identifying the operation being traced.
    *   `fn`: An asynchronous function representing the operation to be traced. This function receives the active `Span` object as an argument.
    *   `attributes` (optional): A record of key-value pairs representing additional metadata to associate with the span. Values can be strings, numbers, or booleans.
*   **Functionality:**  `withSpan` automatically starts a span, executes the provided function within that span, records the span’s status based on the function’s success or failure, and ends the span.  Exceptions thrown by the function are re-thrown after being recorded on the span.

**Accessing and Augmenting the Current Span**

*   `getCurrentSpan()`: Returns the currently active `Span` object, if one exists.  Returns `undefined` if no span is active in the current context.
*   `addSpanEvent()`: Adds an event to the current active span.
    *   `name`: A string describing the event.
    *   `attributes` (optional): A record of key-value pairs providing additional details about the event.

**Shutdown**

The `shutdown` function gracefully shuts down the OpenTelemetry SDK. 

*   **Behavior:**  It attempts to shut down the SDK, ignoring any errors that occur during the shutdown process.  After shutdown, the SDK is set to `null`. You should call this function when the application is exiting to ensure proper resource cleanup.