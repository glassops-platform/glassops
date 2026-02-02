---
type: Documentation
domain: runtime
origin: packages/runtime/internal/telemetry/telemetry.go
last_modified: 2026-02-01
generated: true
source: packages/runtime/internal/telemetry/telemetry.go
generated_at: 2026-02-01T19:45:51.566999
hash: de6e42c3e0bcb05f5d41072920cb61dfd827fe759cc2b3874316d72a2191e789
---

## Telemetry Package Documentation

This package provides integration with OpenTelemetry for distributed tracing within the glassops runtime environment. It allows for the monitoring and analysis of application performance and behavior.

**Package Responsibilities:**

The primary responsibility of this package is to initialize, manage, and provide access to an OpenTelemetry tracer. It handles the configuration of the tracer based on environment variables, the creation of spans to track operations, and the reporting of trace data to a configured backend.  The package aims to be unobtrusive; if the necessary environment variables for configuration are not present, telemetry is effectively disabled.

**Key Types and Interfaces:**

*   **`trace.Tracer`**:  The core OpenTelemetry tracer interface.  This package uses this interface to start spans, add attributes, and record events.
*   **`trace.Span`**: Represents a single operation within a trace. Spans have a start and end time, and can contain attributes and events.
*   **`sdktrace.TracerProvider`**: Manages the lifecycle of tracers and exporters. This package creates and manages a single `TracerProvider` instance.
*   **`attribute.KeyValue`**: Represents a key-value pair used to add metadata to spans and events.

**Important Functions:**

*   **`Init(ctx context.Context, serviceName, serviceVersion string) error`**:
    This function initializes the OpenTelemetry SDK. It reads the `OTEL_EXPORTER_OTLP_ENDPOINT` environment variable to determine the address of the OpenTelemetry collector. If the environment variable is not set, the function returns `nil`, effectively disabling telemetry. It also parses `OTEL_EXPORTER_OTLP_HEADERS` to add custom headers to the outgoing requests. It creates a resource with service name and version attributes and configures the tracer provider.
*   **`Shutdown(ctx context.Context) error`**:
    This function gracefully shuts down the OpenTelemetry tracer provider, flushing any pending data. It returns an error if shutdown fails; otherwise, it returns `nil`. It does nothing if the tracer provider has not been initialized.
*   **`WithSpan[T any](ctx context.Context, name string, fn func(context.Context, trace.Span) (T, error), attrs ...attribute.KeyValue) (T, error)`**:
    This is a utility function for executing a function within a traced span. It starts a new span with the given name and attributes, calls the provided function `fn` with the context and span, and then ends the span. It records any errors that occur during the execution of `fn` as span events. The function returns the result of `fn` and any error it returns.
*   **`GetCurrentSpan(ctx context.Context) trace.Span`**:
    This function retrieves the currently active span from the provided context. It returns `nil` if no span is present in the context.
*   **`AddSpanEvent(ctx context.Context, name string, attrs ...attribute.KeyValue)`**:
    This function adds an event to the current span in the provided context. It retrieves the current span and adds an event with the given name and attributes. It does nothing if no span is present in the context.

**Error Handling:**

The package uses standard Go error handling patterns. Functions return an `error` value to indicate failure. Errors are recorded as span events when they occur within a `WithSpan` block, providing context for debugging.

**Concurrency:**

The OpenTelemetry SDK itself handles concurrency internally. This package does not explicitly manage goroutines or channels, but relies on the SDK's ability to handle concurrent tracing operations.

**Design Decisions:**

*   **Environment Variable Configuration:** The package relies on environment variables for configuration, making it easy to enable or disable telemetry without modifying code.
*   **Lazy Initialization:** The tracer is initialized only when the `OTEL_EXPORTER_OTLP_ENDPOINT` environment variable is set, avoiding unnecessary overhead when telemetry is not required.
*   **`WithSpan` Utility:** The `WithSpan` function simplifies the process of creating and managing spans, ensuring that spans are always properly ended, even in the event of errors.
*   **Context Propagation:** The package leverages the OpenTelemetry context propagation mechanism to ensure that spans are correctly associated with requests as they flow through the system.