package telemetry

import (
	"context"
	"testing"

	"github.com/glassops-platform/glassops/internal/core/config"
	"github.com/glassops-platform/glassops/internal/core/run"
)

func TestExportIsOptIn(t *testing.T) {
	rec := run.New("run", nil)
	if err := Export(context.Background(), nil, rec); err != nil {
		t.Fatalf("nil config: %v", err)
	}
	if err := Export(context.Background(), &config.Config{}, rec); err != nil {
		t.Fatalf("no endpoint: %v", err)
	}
	on := &config.Config{Telemetry: config.Telemetry{OTLP: config.OTLP{Endpoint: "https://otel.example.com"}}}
	if err := Export(context.Background(), on, rec); err == nil {
		t.Fatal("the stub exported")
	}
}
