---
type: Documentation
domain: runtime
origin: packages/runtime/internal/telemetry/telemetry.go
last_modified: 2026-01-31
generated: true
source: packages/runtime/internal/telemetry/telemetry.go
generated_at: 2026-01-31T10:06:35.005316
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
*   **`attribute.KeyValue`**: Represents a key-value pair for adding attributes to spans and events.

**Important Functions:**

*   **`Init(ctx context.Context, serviceName, serviceVersion string) error`**:
    This function initializes the OpenTelemetry SDK. It reads the `OTEL_EXPORTER_OTLP_ENDPOINT` environment variable to determine the endpoint for sending trace data. If the environment variable is not set, the function returns `nil`, effectively disabling telemetry. It also parses `OTEL_EXPORTER_OTLP_HEADERS` to add custom headers to the outgoing requests. It creates a resource with service name and version attributes and configures the tracer provider.
*   **`Shutdown(ctx context.Context) error`**:
    This function gracefully shuts down the OpenTelemetry tracer provider, releasing resources. It returns `nil` if the tracer provider was never initialized.
*   **`WithSpan[T any](ctx context.Context, name string, fn func(context.Context, trace.Span) (T, error), attrs ...attribute.KeyValue) (T, error)`**:
    This is a utility function for executing a function within a traced span. It starts a new span with the given name, adds any provided attributes, executes the provided function `fn` with the span context, and then ends the span. It also handles error reporting by setting the span status and recording errors if they occur. The function returns the result of `fn` and any error it returns.
*   **`GetCurrentSpan(ctx context.Context) trace.Span`**:
    This function retrieves the currently active span from the provided context. It returns `nil` if no span is present in the context.
*   **`AddSpanEvent(ctx context.Context, name string, attrs ...attribute.KeyValue)`**:
    This function adds an event to the current span in the provided context. It retrieves the current span and adds an event with the given name and attributes. If no span is present in the context, the function does nothing.

**Error Handling:**

The package uses standard Go error handling patterns. Functions return an `error` value to indicate failure. Errors are recorded on spans using `span.RecordError(err)` when using `WithSpan`, and span status is set accordingly.

**Concurrency:**

The package itself does not explicitly manage goroutines or channels. However, OpenTelemetry is designed to work in concurrent environments, and the underlying SDK handles concurrency internally. The `WithSpan` function is safe to use concurrently with other operations.

**Design Decisions:**

*   **Environment Variable Configuration:** The package relies on environment variables for configuration, making it easy to enable or disable telemetry without modifying code.
*   **Lazy Initialization:** The tracer provider is only initialized if the `OTEL_EXPORTER_OTLP_ENDPOINT` environment variable is set. This avoids unnecessary overhead when telemetry is not needed.
*   **Context Propagation:** The package leverages the OpenTelemetry context propagation mechanism to ensure that spans are correctly associated with requests as they flow through the system.
*   **Generic `WithSpan` Function:** The `WithSpan` function uses generics to provide type safety and flexibility when executing functions within a traced span.