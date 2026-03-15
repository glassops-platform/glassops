package telemetry

import (
	"context"
	"os"
	"testing"

	"go.opentelemetry.io/otel/trace"
)

func TestInitWithoutEndpoint(t *testing.T) {
	// Ensure OTEL endpoint is not set
	originalEndpoint := os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT")
	os.Unsetenv("OTEL_EXPORTER_OTLP_ENDPOINT")
	defer func() {
		if originalEndpoint != "" {
			os.Setenv("OTEL_EXPORTER_OTLP_ENDPOINT", originalEndpoint)
		}
	}()

	ctx := context.Background()
	err := Init(ctx, "test-service", "1.0.0")

	if err != nil {
		t.Errorf("Init should succeed silently when no endpoint is set, got error: %v", err)
	}
}

func TestShutdownWithoutInit(t *testing.T) {
	ctx := context.Background()

	// Shutdown should not panic when called without Init
	err := Shutdown(ctx)
	if err != nil {
		t.Errorf("Shutdown should succeed when tracer not initialized, got error: %v", err)
	}
}

func TestGetCurrentSpanNoContext(t *testing.T) {
	ctx := context.Background()

	// Should return a no-op span, not panic
	span := GetCurrentSpan(ctx)
	if span == nil {
		t.Error("GetCurrentSpan should return a span (even if no-op)")
	}
}

func TestAddSpanEventNoSpan(t *testing.T) {
	ctx := context.Background()

	// Should not panic when no span is active
	AddSpanEvent(ctx, "test-event")
}

func TestWithSpanWithoutTracer(t *testing.T) {
	ctx := context.Background()

	// WithSpan should work even without a tracer initialized
	result, err := WithSpan(ctx, "test-span", func(ctx context.Context, span trace.Span) (string, error) {
		return "success", nil
	})

	if err != nil {
		t.Errorf("WithSpan should succeed, got error: %v", err)
	}

	if result != "success" {
		t.Errorf("expected 'success', got '%s'", result)
	}
}

func TestWithSpanError(t *testing.T) {
	ctx := context.Background()

	_, err := WithSpan(ctx, "test-span", func(ctx context.Context, span trace.Span) (string, error) {
		return "", &testError{"span error"}
	})

	if err == nil {
		t.Error("expected error from WithSpan")
	}

	if err.Error() != "span error" {
		t.Errorf("expected 'span error', got '%s'", err.Error())
	}
}

// testError is a simple error type for testing
type testError struct {
	msg string
}

func (e *testError) Error() string {
	return e.msg
}
