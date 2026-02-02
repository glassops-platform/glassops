---
type: Documentation
domain: runtime
origin: packages/runtime/internal/telemetry/telemetry_test.go
last_modified: 2026-02-01
generated: true
source: packages/runtime/internal/telemetry/telemetry_test.go
generated_at: 2026-02-01T19:46:10.654269
hash: 32936707e300d67e620fffac350f9a49781e0acc63fd0680f537f440505e41a0
---

## Telemetry Package Documentation

This package provides functionality for integrating OpenTelemetry tracing into applications. It manages the initialization, shutdown, and interaction with the OpenTelemetry tracer. The primary goal is to offer a simple and robust way to add distributed tracing without requiring extensive configuration or impacting application behavior when tracing is not enabled.

**Key Types**

*   `testError`: A custom error type used solely for testing purposes to simulate errors within spans. It implements the `error` interface, allowing it to be returned from functions.

**Interfaces**

This package does not define any custom interfaces. It directly uses the `trace.Span` interface from the `go.opentelemetry.io/otel/trace` package.

**Functions**

*   `Init(ctx context.Context, serviceName string, version string) error`: This function initializes the OpenTelemetry tracer. It takes the context, service name, and version as input. If the `OTEL_EXPORTER_OTLP_ENDPOINT` environment variable is not set, `Init` will proceed without an exporter, creating a no-op tracer. This allows the application to run without requiring a tracing backend. It returns an error if initialization fails; however, it is designed to succeed silently when no endpoint is configured.
*   `Shutdown(ctx context.Context) error`: This function shuts down the OpenTelemetry tracer. It gracefully flushes any pending data and releases resources. It is designed to be safe to call even if `Init` has not been called, preventing panics. It returns an error if shutdown fails, but will not return an error if the tracer was never initialized.
*   `GetCurrentSpan(ctx context.Context) trace.Span`: This function retrieves the currently active OpenTelemetry span from the provided context. If no span is active, it returns a no-op span, ensuring that accessing the span does not cause a panic.
*   `AddSpanEvent(ctx context.Context, name string) `: This function adds an event to the currently active span. If no span is active, it does nothing, preventing a panic.
*   `WithSpan(ctx context.Context, name string, fn func(context.Context, trace.Span) (string, error)) (string, error)`: This function creates a new OpenTelemetry span, executes a provided function within that span’s context, and returns the function’s result. It handles the creation and completion of the span automatically. Importantly, it functions correctly even if the OpenTelemetry tracer has not been initialized, creating a no-op span in that case. If the provided function returns an error, `WithSpan` propagates that error.

**Error Handling**

The package employs standard Go error handling practices. Functions return an `error` value to indicate failure. The `testError` type is used in tests to simulate specific error conditions. The design prioritizes preventing panics, even in scenarios where the tracer is not initialized or a span is not active.

**Concurrency**

The package itself does not explicitly manage goroutines or channels. However, the underlying OpenTelemetry SDK may use concurrency internally for tasks such as exporting trace data. The `context.Context` is used for managing cancellation and deadlines, which can interact with concurrent operations within the OpenTelemetry SDK.

**Design Decisions**

*   **No-op Tracer:** The package is designed to function gracefully even when an OpenTelemetry endpoint is not configured. In this case, it creates a no-op tracer, which allows the application to run without any tracing overhead.
*   **Panic Prevention:**  Several functions are designed to prevent panics by returning default values or doing nothing when a span is not active or the tracer is not initialized. This enhances the robustness of the package.
*   **Simplified Interface:** The package provides a simple and easy-to-use interface for interacting with OpenTelemetry, hiding the complexity of the underlying SDK.
*   **Testability:** The inclusion of the `testError` type facilitates unit testing and allows for simulating error conditions within spans.