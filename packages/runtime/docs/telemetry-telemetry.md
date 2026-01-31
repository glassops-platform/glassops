---
type: Documentation
domain: runtime
origin: packages/runtime/internal/telemetry/telemetry.go
last_modified: 2026-01-31
generated: true
source: packages/runtime/internal/telemetry/telemetry.go
generated_at: 2026-01-31T09:10:32.155837
hash: de6e42c3e0bcb05f5d41072920cb61dfd827fe759cc2b3874316d72a2191e789
---

## Telemetry Package Documentation

This package provides integration with OpenTelemetry for distributed tracing within the glassops runtime environment. It allows for the monitoring and analysis of application performance and behavior.

**Package Responsibilities:**

The primary responsibility of this package is to initialize, manage, and provide access to an OpenTelemetry tracer. It handles the configuration of the tracer based on environment variables, the creation of spans to track operations, and the reporting of trace data to a configured backend.  The package aims to be non-intrusive, allowing developers to easily add tracing to their code without significant modification.

**Key Types and Interfaces:**

*   **`trace.Tracer`:**  The core OpenTelemetry tracer interface.  This package provides access to a global tracer instance.  It is used to start new spans.
*   **`trace.Span`:** Represents a single operation within a trace. Spans have a start and end time, and can contain attributes and events.
*   **`attribute.KeyValue`:** Represents a key-value pair used to add metadata to spans and events.
*   **`sdktrace.TracerProvider`:**  Manages the lifecycle of tracers and exporters. This package initializes and shuts down the tracer provider.

**Important Functions:**

*   **`Init(ctx context.Context, serviceName, serviceVersion string) error`:**  Initializes the OpenTelemetry SDK. This function checks for the `OTEL_EXPORTER_OTLP_ENDPOINT` environment variable. If set, it configures an OpenTelemetry exporter to send trace data to the specified endpoint. It also sets service name and version attributes. If the environment variable is not set, the function returns `nil`, effectively disabling telemetry.  It also parses optional headers from the `OTEL_EXPORTER_OTLP_HEADERS` environment variable.
*   **`Shutdown(ctx context.Context) error`:**  Gracefully shuts down the OpenTelemetry tracer provider, flushing any pending data.  It handles the case where the tracer provider has not been initialized.
*   **`WithSpan[T any](ctx context.Context, name string, fn func(context.Context, trace.Span) (T, error), attrs ...attribute.KeyValue) (T, error)`:**  A utility function for executing a function within a traced span. It starts a new span with the given name and attributes, executes the provided function with the span context, and ensures the span is ended (regardless of success or failure).  If the function returns an error, the span’s status is set to `Error` and the error is recorded on the span.
*   **`GetCurrentSpan(ctx context.Context) trace.Span`:** Retrieves the currently active span from the provided context. Returns `nil` if no span is active in the context.
*   **`AddSpanEvent(ctx context.Context, name string, attrs ...attribute.KeyValue)`:** Adds an event to the current span, if one exists in the provided context.  Events provide additional information about the operation being traced.

**Error Handling:**

The package employs standard Go error handling practices. Functions return an `error` value to indicate failure. Errors during initialization (e.g., invalid endpoint configuration) are returned by `Init`.  The `WithSpan` function captures errors from the executed function and records them on the span.

**Concurrency:**

The OpenTelemetry SDK itself is designed to be concurrency-safe. This package leverages that inherent safety. The `WithSpan` function is safe to be called from multiple goroutines concurrently.

**Design Decisions:**

*   **Environment Variable Configuration:** The package relies on environment variables (`OTEL_EXPORTER_OTLP_ENDPOINT`, `OTEL_EXPORTER_OTLP_HEADERS`) for configuration, allowing for flexible deployment without code changes.
*   **Lazy Initialization:** Telemetry is only initialized if the `OTEL_EXPORTER_OTLP_ENDPOINT` environment variable is set. This avoids unnecessary overhead when tracing is not required.
*   **`WithSpan` Utility:** The `WithSpan` function simplifies the process of adding tracing to existing code by handling span creation, context propagation, and span completion automatically.
*   **Global Tracer:** A single, global tracer instance is used to maintain consistency and reduce resource consumption.