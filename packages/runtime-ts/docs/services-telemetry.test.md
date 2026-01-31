---
type: Documentation
domain: runtime-ts
origin: packages/runtime-ts/src/services/telemetry.test.ts
last_modified: 2026-01-31
generated: true
source: packages/runtime-ts/src/services/telemetry.test.ts
generated_at: 2026-01-31T09:18:05.651699
hash: 56d6c19ce86ce7a7db9db3a9c9f6f32bb0af562b686006d765c5360258bd239b
---

## OpenTelemetry Runtime Service Documentation

This document details the functionality of the OpenTelemetry integration service, providing an overview for both developers and operators. This service facilitates distributed tracing within applications.

**Overview**

The service provides tools for initializing OpenTelemetry, creating spans to track operation execution, adding events to spans, retrieving the current span context, and gracefully shutting down the telemetry system. It is designed to be configurable through environment variables.

**Initialization (initTelemetry)**

The `initTelemetry` function sets up the OpenTelemetry SDK.  Initialization is conditional: it only proceeds if the `OTEL_EXPORTER_OTLP_ENDPOINT` environment variable is defined. This variable specifies the endpoint for exporting trace data.

*   **Parameters:**
    *   `serviceName` (string): The name of the service being traced.
    *   `serviceVersion` (string): The version of the service.
*   **Configuration:**
    *   `OTEL_EXPORTER_OTLP_ENDPOINT`: Required. Specifies the OTLP endpoint (e.g., `http://localhost:4318`).
    *   `OTEL_EXPORTER_OTLP_HEADERS`: Optional. Allows setting custom headers for the OTLP exporter in the format “Key=Value,Key2=Value2”.

**Span Management**

*   **withSpan(spanName, callback, attributes):**  This function executes a provided asynchronous callback function within the context of a new OpenTelemetry span. It automatically handles span creation, completion, and error propagation.
    *   `spanName` (string): The name of the span.
    *   `callback` (function): The asynchronous function to execute within the span.  The current span object is passed as an argument to the callback.
    *   `attributes` (object, optional):  Key-value pairs to attach as attributes to the span.
*   **addSpanEvent(eventName, attributes):** Records an event on the current span.  This function does not error if no span is active.
    *   `eventName` (string): The name of the event.
    *   `attributes` (object, optional): Key-value pairs to attach as attributes to the event.
*   **getCurrentSpan():** Returns the currently active span, or `undefined` if no span is active.

**Shutdown (shutdown)**

The `shutdown` function gracefully shuts down the OpenTelemetry SDK, flushing any remaining data and releasing resources. It is safe to call multiple times or without prior initialization.

**Error Handling**

The `withSpan` function propagates any errors thrown by the wrapped callback function.  Other functions are designed to be non-fatal and will not throw errors in common scenarios (e.g., attempting to add an event to a non-existent span).

**Dependencies**

This service relies on the OpenTelemetry SDK and exporter packages. Mock implementations are used in testing to isolate the service’s logic.

**Usage**

You can initialize the service by setting the `OTEL_EXPORTER_OTLP_ENDPOINT` environment variable and calling `initTelemetry`.  Then, use `withSpan` to wrap critical sections of your code for tracing. Remember to call `shutdown` when your application is exiting to ensure all trace data is exported.