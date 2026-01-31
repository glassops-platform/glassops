/**
 * OpenTelemetry integration for GlassOps Runtime.
 * Provides tracing and metrics for governance operations.
 */

import { trace, SpanStatusCode, Span, context } from "@opentelemetry/api";
import { NodeSDK } from "@opentelemetry/sdk-node";
import { OTLPTraceExporter } from "@opentelemetry/exporter-trace-otlp-http";
import {
  ATTR_SERVICE_NAME,
  ATTR_SERVICE_VERSION,
} from "@opentelemetry/semantic-conventions";
import { Resource } from "@opentelemetry/resources";

let sdk: NodeSDK | null = null;
const TRACER_NAME = "glassops-runtime";

/**
 * Initialize OpenTelemetry SDK.
 * Only initializes if OTEL_EXPORTER_OTLP_ENDPOINT is set.
 */
export async function initTelemetry(
  serviceName: string = "glassops-runtime",
  serviceVersion: string = "1.0.0",
): Promise<void> {
  const endpoint = process.env.OTEL_EXPORTER_OTLP_ENDPOINT;

  if (!endpoint) {
    // No endpoint configured, telemetry disabled
    return;
  }

  try {
    const resource = new Resource({
      [ATTR_SERVICE_NAME]: serviceName,
      [ATTR_SERVICE_VERSION]: serviceVersion,
    });

    const headersRaw = process.env.OTEL_EXPORTER_OTLP_HEADERS;
    const headers: Record<string, string> = {};

    if (headersRaw) {
      // Parse "Header=Value,Header2=Value2" format
      headersRaw.split(",").forEach((header) => {
        const [key, value] = header.split("=");
        if (key && value) {
          headers[key.trim()] = value.trim();
        }
      });
    }

    const traceExporter = new OTLPTraceExporter({
      url: `${endpoint}/v1/traces`,
      headers,
    });

    sdk = new NodeSDK({
      resource,
      traceExporter,
    });

    await sdk.start();
  } catch (error) {
    // Telemetry initialization failure should not block execution
    console.warn(
      `[Telemetry] Failed to initialize: ${error instanceof Error ? error.message : String(error)}`,
    );
  }
}

/**
 * Shutdown telemetry gracefully.
 */
export async function shutdown(): Promise<void> {
  if (sdk) {
    try {
      await sdk.shutdown();
    } catch {
      // Ignore shutdown errors
    }
    sdk = null;
  }
}

/**
 * Execute an async function within a traced span.
 */
export async function withSpan<T>(
  name: string,
  fn: (span: Span) => Promise<T>,
  attributes?: Record<string, string | number | boolean>,
): Promise<T> {
  const tracer = trace.getTracer(TRACER_NAME);

  return tracer.startActiveSpan(name, async (span) => {
    if (attributes) {
      span.setAttributes(attributes);
    }

    try {
      const result = await fn(span);
      span.setStatus({ code: SpanStatusCode.OK });
      return result;
    } catch (error) {
      const message = error instanceof Error ? error.message : String(error);
      span.setStatus({ code: SpanStatusCode.ERROR, message });
      span.recordException(error instanceof Error ? error : new Error(message));
      throw error;
    } finally {
      span.end();
    }
  });
}

/**
 * Get the current active span (if any).
 */
export function getCurrentSpan(): Span | undefined {
  return trace.getSpan(context.active());
}

/**
 * Add an event to the current span.
 */
export function addSpanEvent(
  name: string,
  attributes?: Record<string, string | number | boolean>,
): void {
  const span = getCurrentSpan();
  if (span) {
    span.addEvent(name, attributes);
  }
}
