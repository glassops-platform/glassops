---
type: Documentation
domain: runtime-ts
origin: packages/runtime-ts/src/services/telemetry.ts
last_modified: 2026-01-31
generated: true
source: packages/runtime-ts/src/services/telemetry.ts
generated_at: 2026-01-31T10:14:59.681998
hash: effd62924f7ca6ef31b43af94c93a643466e7be582e31b7b472180bb6b2f408e
---

# GlassOps Runtime Telemetry Documentation

This document details the telemetry integration within GlassOps Runtime, providing tracing and metrics for governance operations. It leverages the OpenTelemetry standard for observability.

## Overview

We provide instrumentation to monitor the performance and behavior of GlassOps Runtime. This is achieved through tracing and metrics, exported to a configured OpenTelemetry collector.  Telemetry assists in identifying performance bottlenecks, understanding system interactions, and debugging issues.

## Initialization

The `initTelemetry` function sets up the OpenTelemetry SDK.

```typescript
async function initTelemetry(
  serviceName: string = "glassops-runtime",
  serviceVersion: string = "1.0.0",
): Promise<void>
```

*   **Purpose:** Initializes the OpenTelemetry SDK, establishing a connection to an OpenTelemetry collector.
*   **Parameters:**
    *   `serviceName` (optional):  A string representing the name of the service. Defaults to "glassops-runtime".
    *   `serviceVersion` (optional): A string representing the version of the service. Defaults to "1.0.0".
*   **Behavior:**
    *   Checks for the `OTEL_EXPORTER_OTLP_ENDPOINT` environment variable. If not set, telemetry is disabled.
    *   Configures a resource with the provided service name and version.
    *   Parses optional headers from the `OTEL_EXPORTER_OTLP_HEADERS` environment variable (format: "Header=Value,Header2=Value2").
    *   Creates an OTLPTraceExporter configured with the endpoint URL and headers.
    *   Starts the OpenTelemetry SDK.
    *   Logs warnings if initialization fails, but does not halt execution.

## Shutdown

The `shutdown` function gracefully shuts down the OpenTelemetry SDK.

```typescript
async function shutdown(): Promise<void>
```

*   **Purpose:**  Shuts down the OpenTelemetry SDK, releasing resources.
*   **Behavior:**
    *   If the SDK is initialized, attempts to shut it down.
    *   Ignores any errors that occur during shutdown.
    *   Sets the SDK to `null`.

## Span Management

We offer functions to manage OpenTelemetry spans for tracing specific operations.

### `withSpan`

```typescript
async function withSpan<T>(
  name: string,
  fn: (span: Span) => Promise<T>,
  attributes?: Record<string, string | number | boolean>,
): Promise<T>
```

*   **Purpose:** Executes an asynchronous function within the context of an OpenTelemetry span.
*   **Parameters:**
    *   `name`: A string representing the name of the span.
    *   `fn`: An asynchronous function that accepts a `Span` object as an argument and returns a value of type `T`.
    *   `attributes` (optional): A record of key-value pairs to add as attributes to the span. Values can be strings, numbers, or booleans.
*   **Behavior:**
    *   Starts a new span with the given name.
    *   Adds the provided attributes to the span.
    *   Executes the provided function `fn` within the span's context.
    *   Sets the span status to `OK` if the function completes successfully.
    *   Sets the span status to `ERROR` if the function throws an error, recording the error message and exception.
    *   Ends the span in a `finally` block, ensuring it always completes.
    *   Re-throws any errors that occur within the function.

### `getCurrentSpan`

```typescript
function getCurrentSpan(): Span | undefined
```

*   **Purpose:** Retrieves the currently active span, if one exists.
*   **Returns:** A `Span` object if a span is active, otherwise `undefined`.

### `addSpanEvent`

```typescript
function addSpanEvent(
  name: string,
  attributes?: Record<string, string | number | boolean>,
): void
```

*   **Purpose:** Adds an event to the current active span.
*   **Parameters:**
    *   `name`: A string representing the name of the event.
    *   `attributes` (optional): A record of key-value pairs to add as attributes to the event.
*   **Behavior:**
    *   Retrieves the current active span using `getCurrentSpan`.
    *   If a span is active, adds an event to it with the given name and attributes.



## Configuration

Telemetry is controlled through environment variables:

*   `OTEL_EXPORTER_OTLP_ENDPOINT`:  Specifies the URL of the OpenTelemetry collector.  If not set, telemetry is disabled.
*   `OTEL_EXPORTER_OTLP_HEADERS`: Specifies custom headers to send with telemetry data (format: "Header=Value,Header2=Value2").