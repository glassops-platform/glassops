---
type: Documentation
domain: runtime-ts
origin: packages/runtime-ts/src/services/telemetry.test.ts
last_modified: 2026-01-31
generated: true
source: packages/runtime-ts/src/services/telemetry.test.ts
generated_at: 2026-01-31T10:14:34.428258
hash: 56d6c19ce86ce7a7db9db3a9c9f6f32bb0af562b686006d765c5360258bd239b
---

## OpenTelemetry Integration Service Documentation

This document details the functionality of the OpenTelemetry integration service, providing an overview for both developers and operators. This service facilitates the collection and export of tracing data for application monitoring and performance analysis.

### Overview

The service integrates with OpenTelemetry to instrument applications, capturing trace information. This data is then exported to a configured backend, such as an OpenTelemetry Collector, for storage and visualization. The primary functions of this service are initialization, span management, event logging, and graceful shutdown.

### Key Components

*   **`initTelemetry(serviceName, version)`**: Initializes the OpenTelemetry SDK. This function checks for the `OTEL_EXPORTER_OTLP_ENDPOINT` environment variable. If present, it configures and starts the SDK, connecting to the specified endpoint. It also supports configuring headers via the `OTEL_EXPORTER_OTLP_HEADERS` environment variable. If the endpoint is not set, initialization is skipped.
*   **`withSpan(spanName, callback, attributes)`**: Executes a provided asynchronous function within the context of a new OpenTelemetry span. This function automatically starts and ends the span, propagating tracing context. The `callback` function receives the active span as an argument, allowing for custom attribute addition or manual span control.
*   **`addSpanEvent(eventName, attributes)`**: Adds an event to the currently active span. This allows for recording significant occurrences within a trace. The function handles cases where no span is active without error.
*   **`getCurrentSpan()`**: Returns the currently active OpenTelemetry span, if one exists. Returns `undefined` if no span is active.
*   **`shutdown()`**: Gracefully shuts down the OpenTelemetry SDK, ensuring that all pending data is exported before termination. This function can be called multiple times without error.

### Configuration

The service relies on environment variables for configuration:

*   **`OTEL_EXPORTER_OTLP_ENDPOINT`**:  Specifies the URL of the OpenTelemetry collector or other OTLP receiver.  If this variable is not set, telemetry will not be initialized. Example: `http://localhost:4318`.
*   **`OTEL_EXPORTER_OTLP_HEADERS`**: Specifies HTTP headers to be included in the trace export requests. Headers should be comma-separated key-value pairs. Example: `Authorization=Basic 123,Custom-Header=Values`.

### Usage

1.  **Initialization:** Call `initTelemetry()` with your service name and version after environment variables are set.
2.  **Tracing:** Wrap critical sections of code with `withSpan()` to create spans and capture timing information.
3.  **Event Logging:** Use `addSpanEvent()` to record significant events within your spans.
4.  **Shutdown:** Call `shutdown()` before your application exits to ensure all telemetry data is flushed.

### Error Handling

The service is designed to be resilient. Functions like `addSpanEvent()` and `shutdown()` will not throw errors if called in unexpected states (e.g., no active span, SDK not initialized). `withSpan()` propagates errors thrown by the wrapped function.