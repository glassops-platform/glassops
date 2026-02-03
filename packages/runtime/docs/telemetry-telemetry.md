---
type: Documentation
domain: runtime
last_modified: 2026-02-02
generated: true
source: packages/runtime/internal/telemetry/telemetry.go
generated_at: 2026-02-02T22:41:28.958459
hash: de6e42c3e0bcb05f5d41072920cb61dfd827fe759cc2b3874316d72a2191e789
---

## glassops Runtime Telemetry Package Documentation

This package provides integration with OpenTelemetry for distributed tracing within the glassops runtime environment. It allows for the monitoring and analysis of application performance and behavior.

**Package Responsibilities:**

The primary responsibility of this package is to initialize, manage, and provide access to an OpenTelemetry tracer. It handles the configuration of the tracer based on environment variables, the creation of spans to track operations, and the reporting of trace data to a configured backend.

**Key Types and Interfaces:**

- `trace.Tracer`:  The core OpenTelemetry interface for creating and managing traces and spans.  We use this to start new tracing spans.
- `trace.Span`: Represents a single operation within a trace.  Spans are used to measure the duration and attributes of specific code sections.
- `sdktrace.TracerProvider`:  Manages the lifecycle of tracers and exporters.  It is responsible for configuring and shutting down the tracing pipeline.
- `attribute.KeyValue`: Represents a key-value pair used to add metadata to spans and events.

**Important Functions:**

- `Init(ctx context.Context, serviceName, serviceVersion string) error`:
  This function initializes the OpenTelemetry SDK. It checks for the `OTEL_EXPORTER_OTLP_ENDPOINT` environment variable. If this variable is set, it configures an OpenTelemetry exporter to send trace data to the specified endpoint. It also sets service name and version attributes. If the environment variable is not set, telemetry is disabled, and the function returns nil. It also parses optional headers from the `OTEL_EXPORTER_OTLP_HEADERS` environment variable.
- `Shutdown(ctx context.Context) error`:
  This function gracefully shuts down the OpenTelemetry tracer provider, releasing resources and flushing any pending trace data. It returns an error if shutdown fails; otherwise, it returns nil.
- `WithSpan[T any](ctx context.Context, name string, fn func(context.Context, trace.Span) (T, error), attrs ...attribute.KeyValue) (T, error)`:
  This is a utility function for executing a function within a traced span. It starts a new span with the given name and attributes, executes the provided function with the span context, and ensures the span is ended (regardless of success or failure). If the function returns an error, the span's status is set to `Error`; otherwise, it's set to `Ok`.  You pass a function `fn` that accepts a `context.Context` and a `trace.Span` as arguments.
- `GetCurrentSpan(ctx context.Context) trace.Span`:
  This function retrieves the currently active span from the provided context. It returns nil if no span is active in the context.
- `AddSpanEvent(ctx context.Context, name string, attrs ...attribute.KeyValue)`:
  This function adds an event to the current span.  It retrieves the current span from the context and adds an event with the given name and attributes. If no span is present in the context, the function does nothing.

**Error Handling:**

The package employs standard Go error handling practices. Functions return an `error` value to indicate failure. Errors are typically wrapped and propagated to provide context. Within `WithSpan`, errors from the executed function are recorded on the span using `span.RecordError` and the span status is set accordingly.

**Concurrency:**

The OpenTelemetry SDK itself is designed to be concurrency-safe. This package leverages that inherent safety.  The `WithSpan` function is safe to be called concurrently from multiple goroutines.

**Design Decisions:**

- **Conditional Initialization:** Telemetry is only initialized if the `OTEL_EXPORTER_OTLP_ENDPOINT` environment variable is set. This allows for easy disabling of tracing in environments where it is not desired or necessary.
- **Environment Variable Configuration:** The package relies on environment variables for configuration, making it easy to adapt to different deployment environments without code changes.
- **Context Propagation:**  The package utilizes Go's context mechanism to propagate trace context across function calls and goroutines.
- **Generic `WithSpan` Function:** The `WithSpan` function is implemented using generics to provide type safety and flexibility.