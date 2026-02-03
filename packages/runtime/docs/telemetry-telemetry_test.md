---
type: Documentation
domain: runtime
last_modified: 2026-02-02
generated: true
source: packages/runtime/internal/telemetry/telemetry_test.go
generated_at: 2026-02-02T22:41:47.307296
hash: 32936707e300d67e620fffac350f9a49781e0acc63fd0680f537f440505e41a0
---

## Telemetry Package Documentation

This package provides functionality for integrating OpenTelemetry tracing into applications. It handles initialization, shutdown, and basic operations for interacting with spans and events. The primary goal is to offer a simple and resilient telemetry solution, even in environments where a full tracing backend is not configured.

**Key Responsibilities:**

*   Initialization of the OpenTelemetry tracer.
*   Graceful shutdown of the OpenTelemetry tracer.
*   Retrieval of the current span from a given context.
*   Adding events to the current span.
*   Executing a function within the context of a new span.

**Key Types:**

*   `testError`: A custom error type used solely for testing purposes to simulate errors within span operations. It implements the `error` interface.

**Important Functions:**

*   `Init(ctx context.Context, serviceName string, version string) error`: This function initializes the OpenTelemetry tracer. It takes the context, service name, and version as input. If the `OTEL_EXPORTER_OTLP_ENDPOINT` environment variable is not set, it initializes the tracer without a configured exporter, resulting in a no-op tracer.  It returns an error if initialization fails, but is designed to succeed silently when no endpoint is provided.
*   `Shutdown(ctx context.Context) error`: This function shuts down the OpenTelemetry tracer. It gracefully stops the tracer and releases resources. It does not panic if the tracer was never initialized. It returns an error if shutdown fails, but will succeed if the tracer was never initialized.
*   `GetCurrentSpan(ctx context.Context) trace.Span`: This function retrieves the current span associated with the given context. If no span is active in the context, it returns a no-op span, preventing panics.
*   `AddSpanEvent(ctx context.Context, name string)`: This function adds an event to the current span. If no span is active, it does nothing and does not return an error.
*   `WithSpan(ctx context.Context, name string, fn func(context.Context, trace.Span) (string, error)) (string, error)`: This function executes the provided function `fn` within the context of a new span. It automatically starts and ends the span. The function takes the context, span name, and a function as input. It returns the result of the function and any error that occurred during its execution. If the tracer is not initialized, it still executes the function without creating a span.

**Error Handling:**

The package employs standard Go error handling practices. Functions return an `error` value to indicate failure. The `testError` type is used in tests to simulate specific error conditions.  The `WithSpan` function propagates errors returned by the provided function.

**Concurrency:**

The OpenTelemetry SDK handles concurrency internally. This package itself does not explicitly manage goroutines or channels, but relies on the concurrent safety of the underlying OpenTelemetry APIs.

**Design Decisions:**

*   **Resilience to Missing Configuration:** The `Init` function is designed to operate without a configured OpenTelemetry endpoint. This allows applications to run in environments where tracing is not yet set up without crashing.
*   **No-Op Behavior:**  Functions like `GetCurrentSpan` and `AddSpanEvent` are designed to gracefully handle cases where a span is not active, preventing panics.
*   **Simplified Interface:** The package provides a minimal set of functions to cover common tracing use cases, keeping the API simple and easy to use.
*   **Testability:** The inclusion of the `testError` type facilitates unit testing of error handling scenarios.