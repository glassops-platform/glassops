/**
 * OpenTelemetry integration for GlassOps Runtime.
 * Provides tracing and metrics for governance operations.
 */
import { Span } from "@opentelemetry/api";
/**
 * Initialize OpenTelemetry SDK.
 * Only initializes if OTEL_EXPORTER_OTLP_ENDPOINT is set.
 */
export declare function initTelemetry(serviceName?: string, serviceVersion?: string): Promise<void>;
/**
 * Shutdown telemetry gracefully.
 */
export declare function shutdown(): Promise<void>;
/**
 * Execute an async function within a traced span.
 */
export declare function withSpan<T>(name: string, fn: (span: Span) => Promise<T>, attributes?: Record<string, string | number | boolean>): Promise<T>;
/**
 * Get the current active span (if any).
 */
export declare function getCurrentSpan(): Span | undefined;
/**
 * Add an event to the current span.
 */
export declare function addSpanEvent(name: string, attributes?: Record<string, string | number | boolean>): void;
//# sourceMappingURL=telemetry.d.ts.map