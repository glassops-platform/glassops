---
type: Documentation
domain: runtime-ts
origin: packages/runtime-ts/src/services/telemetry.test.ts
last_modified: 2026-01-29
generated: true
source: packages/runtime-ts/src/services/telemetry.test.ts
generated_at: 2026-01-29T21:01:43.765138
hash: 56d6c19ce86ce7a7db9db3a9c9f6f32bb0af562b686006d765c5360258bd239b
---

## OpenTelemetry Integration Service Documentation

This document details the functionality of the OpenTelemetry integration service. This service provides tools for tracing and monitoring application performance.

**Overview**

The service facilitates the collection and export of telemetry data, enabling observability into application behavior. It integrates with OpenTelemetry standards for vendor neutrality and flexibility.

**Key Components**

*   **`initTelemetry(serviceName, version)`:** Initializes the OpenTelemetry SDK. This function checks for the `OTEL_EXPORTER_OTLP_ENDPOINT` environment variable. If present, it configures and starts the SDK, connecting to the specified OpenTelemetry Collector endpoint.  It also supports parsing headers from the `OTEL_EXPORTER_OTLP_HEADERS` environment variable. You must provide a service name and version string when calling this function.
*   **`withSpan(spanName, callback, attributes)`:** Executes a provided function within the context of a new OpenTelemetry span. This allows for precise timing and attribution of operations. The `callback` function receives the active span as an argument.  Optional attributes can be added to the span.
*   **`addSpanEvent(eventName, attributes)`:** Adds an event to the currently active span. This is useful for marking significant occurrences within a traced operation. Attributes can be included with the event. This function handles cases where no span is active without error.
*   **`getCurrentSpan()`:** Returns the currently active OpenTelemetry span, if one exists. Returns `undefined` if no span is active.
*   **`shutdown()`:**  Shuts down the OpenTelemetry SDK, ensuring that all pending data is exported and resources are released. This function is designed to be called during application shutdown. It handles being called multiple times and without prior initialization gracefully.

**Configuration**

The service relies on environment variables for configuration:

*   **`OTEL_EXPORTER_OTLP_ENDPOINT`:**  Specifies the URL of the OpenTelemetry Collector endpoint. If this variable is not set, the telemetry system will not initialize.
*   **`OTEL_EXPORTER_OTLP_HEADERS`:** Specifies HTTP headers to include with trace exports. Headers should be formatted as a comma-separated list of `key=value` pairs.

**Usage Notes**

*   We recommend calling `initTelemetry()` early in the application lifecycle, ideally during startup.
*   Always call `shutdown()` during application shutdown to ensure data is flushed.
*   `withSpan()` is the primary mechanism for tracing individual operations.
*   Error handling within `withSpan()` callbacks functions as expected; errors will propagate.
*   The service is designed to be resilient to missing or invalid configuration.