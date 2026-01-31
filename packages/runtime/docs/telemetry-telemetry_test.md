---
type: Documentation
domain: runtime
origin: packages/runtime/internal/telemetry/telemetry_test.go
last_modified: 2026-01-31
generated: true
source: packages/runtime/internal/telemetry/telemetry_test.go
generated_at: 2026-01-31T09:10:49.909310
hash: 32936707e300d67e620fffac350f9a49781e0acc63fd0680f537f440505e41a0
---

## Telemetry Package Documentation

This package provides a lightweight wrapper around OpenTelemetry for application tracing and telemetry data collection. It aims to simplify the integration of tracing into applications without imposing strict requirements on the tracing backend. We designed it to function gracefully even when a full tracing setup is not available.

**Package Responsibilities:**

*   Initialization and shutdown of the OpenTelemetry tracer.
*   Providing functions to access the current span within a given context.
*   Adding events to the current span.
*   Executing functions within the context of a new span.

**Key Types:**

*   `testError`: A custom error type used solely for testing purposes to simulate errors within span execution. It implements the `error` interface.

**Important Functions:**

*   `Init(ctx context.Context, serviceName string, version string) error`: Initializes the OpenTelemetry tracer. It takes the context, service name, and service version as input. If the `OTEL_EXPORTER_OTLP_ENDPOINT` environment variable is not set, it initializes a tracer without an exporter, effectively creating a no-op tracer.  It returns an error if initialization fails, but is designed to succeed silently when no endpoint is configured.
*   `Shutdown(ctx context.Context) error`: Shuts down the OpenTelemetry tracer. It gracefully handles cases where the tracer was never initialized, preventing panics. It returns an error if shutdown fails, but will not return an error if the tracer was never initialized.
*   `GetCurrentSpan(ctx context.Context) trace.Span`: Retrieves the currently active span from the provided context. If no span is active, it returns a no-op span, ensuring the application does not panic.
*   `AddSpanEvent(ctx context.Context, name string)`: Adds an event to the current span. It handles cases where no span is active without causing a panic.
*   `WithSpan(ctx context.Context, operationName string, fn func(context.Context, trace.Span) (string, error)) (string, error)`: Executes the provided function `fn` within the context of a new span. It automatically starts and ends the span.  It handles cases where the tracer is not initialized, still executing the function and returning its result. If the function `fn` returns an error, `WithSpan` propagates that error.

**Error Handling:**

The package prioritizes graceful error handling. Functions are designed to avoid panics, even in situations where the OpenTelemetry tracer is not fully configured. Errors are returned when appropriate, allowing the calling code to handle them.  Custom error types, like `testError`, are used in testing to simulate specific error scenarios.

**Concurrency:**

This package itself does not explicitly manage goroutines or channels. However, OpenTelemetry, which this package wraps, is inherently concurrency-safe. The context passed to functions within this package is used to propagate tracing information across concurrent operations managed by the application.

**Design Decisions:**

*   **No-op Tracer:** We chose to support a no-op tracer when a tracing endpoint is not configured. This allows applications to be deployed in environments where tracing is not available without requiring code changes.
*   **Graceful Degradation:** Functions are designed to handle missing spans or uninitialized tracers without panicking, providing a more robust experience.
*   **Simplified Interface:** The package provides a simple and intuitive interface for common tracing operations, reducing the complexity of integrating OpenTelemetry into applications.