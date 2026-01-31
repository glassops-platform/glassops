// Package telemetry provides OpenTelemetry integration for tracing.
package telemetry

import (
	"context"
	"os"
	"strings"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.24.0"
	"go.opentelemetry.io/otel/trace"
)

const tracerName = "glassops-runtime"

var (
	tracerProvider *sdktrace.TracerProvider
	tracer         trace.Tracer
)

// Init initializes OpenTelemetry SDK.
// Only initializes if OTEL_EXPORTER_OTLP_ENDPOINT is set.
func Init(ctx context.Context, serviceName, serviceVersion string) error {
	endpoint := os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT")
	if endpoint == "" {
		return nil // Telemetry disabled
	}

	// Parse headers from environment
	headersRaw := os.Getenv("OTEL_EXPORTER_OTLP_HEADERS")
	headers := make(map[string]string)
	if headersRaw != "" {
		for _, header := range strings.Split(headersRaw, ",") {
			parts := strings.SplitN(header, "=", 2)
			if len(parts) == 2 {
				headers[strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1])
			}
		}
	}

	exporter, err := otlptracehttp.New(ctx,
		otlptracehttp.WithEndpoint(strings.TrimPrefix(endpoint, "https://")),
		otlptracehttp.WithHeaders(headers),
	)
	if err != nil {
		return err
	}

	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceNameKey.String(serviceName),
			semconv.ServiceVersionKey.String(serviceVersion),
		),
	)
	if err != nil {
		return err
	}

	tracerProvider = sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
	)

	otel.SetTracerProvider(tracerProvider)
	tracer = otel.Tracer(tracerName)

	return nil
}

// Shutdown gracefully shuts down the telemetry provider.
func Shutdown(ctx context.Context) error {
	if tracerProvider != nil {
		return tracerProvider.Shutdown(ctx)
	}
	return nil
}

// WithSpan executes a function within a traced span.
func WithSpan[T any](ctx context.Context, name string, fn func(context.Context, trace.Span) (T, error), attrs ...attribute.KeyValue) (T, error) {
	if tracer == nil {
		tracer = otel.Tracer(tracerName)
	}

	ctx, span := tracer.Start(ctx, name)
	defer span.End()

	if len(attrs) > 0 {
		span.SetAttributes(attrs...)
	}

	result, err := fn(ctx, span)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
	} else {
		span.SetStatus(codes.Ok, "")
	}

	return result, err
}

// GetCurrentSpan returns the current active span from context.
func GetCurrentSpan(ctx context.Context) trace.Span {
	return trace.SpanFromContext(ctx)
}

// AddSpanEvent adds an event to the current span.
func AddSpanEvent(ctx context.Context, name string, attrs ...attribute.KeyValue) {
	span := trace.SpanFromContext(ctx)
	if span != nil {
		span.AddEvent(name, trace.WithAttributes(attrs...))
	}
}
