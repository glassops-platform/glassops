---
type: Documentation
domain: runtime
origin: packages/runtime/internal/telemetry/telemetry_test.go
last_modified: 2026-01-29
generated: true
source: packages/runtime/internal/telemetry/telemetry_test.go
generated_at: 2026-01-29T21:30:02.064899
hash: 32936707e300d67e620fffac350f9a49781e0acc63fd0680f537f440505e41a0
---

## Telemetry Package Documentation

This package provides a lightweight wrapper around OpenTelemetry for application instrumentation. It aims to simplify tracing without imposing strict requirements on the underlying tracing backend. We designed it to be resilient to missing configurations and to function gracefully even when a full tracing setup is not available.

**Package Responsibilities:**

*   Initialization and shutdown of the OpenTelemetry tracer.
*   Providing functions to access the current span.
*   Adding events to the current span.
*   Executing code within the context of a new span.

**Key Types:**

*   `testError`: A custom error type used solely for testing purposes to simulate errors within span execution. It implements the `error` interface.

**Important Functions:**

*   `Init(ctx context.Context, serviceName string, version string) error`: Initializes the OpenTelemetry tracer. It takes a context, service name, and version as input. If the `OTEL_EXPORTER_OTLP_ENDPOINT` environment variable is not set, it initializes a no-op tracer. It returns an error if initialization fails, but is designed to succeed silently when no endpoint is configured.
*   `Shutdown(ctx context.Context) error`: Shuts down the OpenTelemetry tracer. It gracefully handles cases where the tracer was never initialized, preventing panics. It returns an error if shutdown fails, but will not return an error if the tracer was never initialized.
*   `GetCurrentSpan(ctx context.Context) trace.Span`: Returns the currently active OpenTelemetry span from the provided context. If no span is active, it returns a no-op span, ensuring the application does not panic.
*   `AddSpanEvent(ctx context.Context, name string)`: Adds an event to the current span. It handles cases where no span is active without causing a panic.
*   `WithSpan(ctx context.Context, name string, fn func\[context.Context, trace.Span] (string, error)) (string, error)`: Executes the provided function `fn` within the context of a new OpenTelemetry span. The function receives the context and the newly created span as arguments. It handles cases where the tracer is not initialized, creating a no-op span in those scenarios. It returns the result of the function and any error that occurred during its execution.

**Error Handling:**

The package prioritizes graceful error handling. Functions are designed to avoid panics, even in the absence of a configured tracing backend. Errors are returned when appropriate, allowing calling code to handle them. The `testError` type is used in tests to simulate specific error conditions.

**Concurrency:**

The package leverages OpenTelemetry’s built-in concurrency features. Spans are inherently context-bound, and OpenTelemetry handles the concurrent access to span data safely. We do not introduce additional concurrency primitives within this package.

**Design Decisions:**

*   **Resilience to Missing Configuration:** We intentionally designed the package to function without requiring a fully configured tracing backend. This allows applications to be deployed in environments where tracing is not yet set up without causing errors.
*   **No-Op Tracer:** When a tracing endpoint is not configured, a no-op tracer is used. This ensures that the application can continue to function without performance degradation.
*   **Context Propagation:** The package relies on OpenTelemetry’s context propagation mechanisms to ensure that spans are correctly associated with requests as they flow through the application.
*   **Simplified Interface:** We provide a simple and intuitive API for accessing tracing functionality, hiding the complexity of the underlying OpenTelemetry implementation.