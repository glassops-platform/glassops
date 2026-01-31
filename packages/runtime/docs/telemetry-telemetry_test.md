---
type: Documentation
domain: runtime
origin: packages/runtime/internal/telemetry/telemetry_test.go
last_modified: 2026-01-31
generated: true
source: packages/runtime/internal/telemetry/telemetry_test.go
generated_at: 2026-01-31T10:06:57.808549
hash: 32936707e300d67e620fffac350f9a49781e0acc63fd0680f537f440505e41a0
---

## Telemetry Package Documentation

This package provides functionality for integrating OpenTelemetry tracing into applications. It manages the initialization, shutdown, and interaction with the OpenTelemetry tracer. The primary goal is to offer a simple and robust way to add distributed tracing without requiring extensive configuration or impacting application behavior when tracing is not fully set up.

**Key Types**

*   `testError`: A custom error type used solely for testing purposes to simulate errors within spans. It implements the `error` interface, allowing it to be returned from functions.

**Interfaces**

This package does not define any custom interfaces. It directly uses the `trace.Span` interface from the `go.opentelemetry.io/otel/trace` package.

**Functions**

*   `Init(ctx context.Context, serviceName string, version string) error`: This function initializes the OpenTelemetry tracer. It takes the context, service name, and version as input. It attempts to configure the tracer based on environment variables (specifically `OTEL_EXPORTER_OTLP_ENDPOINT`). If the endpoint is not set, it initializes a no-op tracer, allowing the application to run without requiring a tracing backend. It returns an error if initialization fails, but is designed to succeed silently when no endpoint is provided.

*   `Shutdown(ctx context.Context) error`: This function shuts down the OpenTelemetry tracer, flushing any pending data. It is safe to call even if `Init` has not been called, preventing panics. It returns an error if shutdown fails, but will not return an error if the tracer was never initialized.

*   `GetCurrentSpan(ctx context.Context) trace.Span`: This function retrieves the currently active OpenTelemetry span from the provided context. If no span is active, it returns a no-op span, ensuring that accessing the span does not cause a panic.

*   `AddSpanEvent(ctx context.Context, name string) `: This function adds an event to the currently active span. It gracefully handles cases where no span is active, preventing panics.

*   `WithSpan(ctx context.Context, name string, fn func(context.Context, trace.Span) (string, error)) (string, error)`: This function creates a new OpenTelemetry span, executes a provided function within that span’s context, and returns the function’s result. It handles cases where the tracer has not been initialized by creating a no-op span. It also propagates any errors returned by the provided function. The function `fn` receives the context and the created span as arguments.

**Error Handling**

The package employs a non-panicking error handling strategy. Functions are designed to return errors when appropriate, but will often succeed with a no-op implementation if a dependency (like a tracing endpoint) is unavailable. Custom error types, like `testError`, are used in tests to simulate specific error conditions.

**Concurrency**

The package itself does not explicitly manage goroutines or channels. However, the underlying OpenTelemetry SDK may use concurrency internally for tasks such as exporting trace data. The `context.Context` is used for managing cancellation and deadlines, which can interact with concurrent operations within the OpenTelemetry SDK.

**Design Decisions**

*   **No-op Tracer:** The package prioritizes graceful degradation. When a tracing endpoint is not configured, it initializes a no-op tracer. This allows applications to run without requiring a tracing backend and avoids runtime errors.
*   **Safe Access to Spans:** Functions like `GetCurrentSpan` and `AddSpanEvent` are designed to be safe to call even when no span is active, preventing panics.
*   **Simplified Interface:** The package provides a minimal set of functions to manage OpenTelemetry tracing, making it easy to integrate into existing applications.
*   **Context Propagation:** The package relies heavily on `context.Context` for propagating tracing information. You should ensure that contexts are properly managed throughout your application.