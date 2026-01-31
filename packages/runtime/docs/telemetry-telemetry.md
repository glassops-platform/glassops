---
type: Documentation
domain: runtime
origin: packages/runtime/internal/telemetry/telemetry.go
last_modified: 2026-01-29
generated: true
source: packages/runtime/internal/telemetry/telemetry.go
generated_at: 2026-01-29T21:29:40.173430
hash: de6e42c3e0bcb05f5d41072920cb61dfd827fe759cc2b3874316d72a2191e789
---

## glassops Runtime Telemetry Package Documentation

This package provides integration with OpenTelemetry for distributed tracing within the glassops runtime environment. It allows for the monitoring and analysis of application performance and behavior.

**Package Responsibilities:**

The primary responsibility of this package is to initialize, manage, and provide access to an OpenTelemetry tracer. It handles the configuration of the tracer based on environment variables, the creation of spans to track operations, and the reporting of trace data to a configured backend.

**Key Types and Interfaces:**

*   **`trace.Tracer`**:  The core OpenTelemetry interface for creating and managing spans. This package exposes a global `tracer` instance.
*   **`trace.Span`**: Represents a single operation within a trace. Spans have a start and end time, and can contain attributes and events.
*   **`attribute.KeyValue`**: Represents a key-value pair used to add metadata to spans and events.
*   **`sdktrace.TracerProvider`**: Manages the lifecycle of tracers and exporters. This package maintains a global `tracerProvider` instance.

**Important Functions:**

*   **`Init(ctx context.Context, serviceName, serviceVersion string) error`**:  Initializes the OpenTelemetry SDK. This function checks for the `OTEL_EXPORTER_OTLP_ENDPOINT` environment variable. If set, it configures an OpenTelemetry exporter to send trace data to the specified endpoint. It also sets service name and version attributes. If the environment variable is not set, telemetry is disabled, and the function returns `nil`. It also parses optional headers from the `OTEL_EXPORTER_OTLP_HEADERS` environment variable.
*   **`Shutdown(ctx context.Context) error`**: Gracefully shuts down the OpenTelemetry tracer provider, releasing resources. It returns an error if shutdown fails; otherwise, it returns `nil`. It does nothing if the tracer provider has not been initialized.
*   **`WithSpan\[T any](ctx context.Context, name string, fn func(context.Context, trace.Span) (T, error), attrs ...attribute.KeyValue) (T, error)`**:  Executes a given function `fn` within the context of a new OpenTelemetry span. The function receives the current context and the created span as arguments.  The span is automatically started at the beginning of the function execution and ended when the function returns. Any attributes provided via `attrs` are added to the span. If the function returns an error, the span’s status is set to `Error` and the error is recorded on the span. The function returns the result of `fn` and any error it may have returned. If the tracer is not initialized, it initializes it with the default tracer name.
*   **`GetCurrentSpan(ctx context.Context) trace.Span`**: Retrieves the currently active span from the provided context. Returns `nil` if no span is active in the context.
*   **`AddSpanEvent(ctx context.Context, name string, attrs ...attribute.KeyValue)`**: Adds an event to the current span associated with the provided context. The event is identified by the given `name` and can include additional attributes specified by `attrs`.  If no span is active in the context, this function does nothing.

**Error Handling:**

The package employs standard Go error handling practices. Functions return an `error` value to indicate failure. Errors are recorded on spans when using `WithSpan` to provide context for tracing.

**Concurrency:**

The package itself does not explicitly manage goroutines or channels. However, OpenTelemetry is designed to work in concurrent environments, and the underlying SDK handles concurrency internally.  The `WithSpan` function is safe to use concurrently.

**Design Decisions:**

*   **Lazy Initialization:** The OpenTelemetry SDK is only initialized if the `OTEL_EXPORTER_OTLP_ENDPOINT` environment variable is set. This allows for disabling telemetry in environments where it is not needed or desired, reducing overhead.
*   **Global Tracer:** A global `tracer` instance is used to simplify access to the OpenTelemetry tracer.
*   **Context Propagation:** The package relies on OpenTelemetry’s context propagation mechanism to ensure that spans are correctly associated with requests as they flow through the system.
*   **Attribute Flexibility:** The `WithSpan` and `AddSpanEvent` functions accept a variable number of `attribute.KeyValue` pairs, allowing for flexible metadata to be added to spans and events.