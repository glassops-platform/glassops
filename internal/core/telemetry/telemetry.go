// Package telemetry exports a finished run as trace spans, when telemetry.otlp is configured.
// Export is opt-in, and a failed export never changes a run's exit code.
package telemetry

import (
	"context"
	"errors"

	"github.com/glassops-platform/glassops/internal/core/config"
	"github.com/glassops-platform/glassops/internal/core/run"
)

// errNotImplemented is what Export returns for a configured endpoint until it lands.
var errNotImplemented = errors.New("telemetry export is not implemented yet")

// Export sends rec, the run step's result, to the endpoint in cfg.Telemetry. With no
// endpoint configured it does nothing and returns nil. It is not implemented yet, so a
// configured endpoint returns an error.
func Export(ctx context.Context, cfg *config.Config, rec *run.Result) error {
	if cfg == nil || !cfg.Telemetry.Enabled() {
		return nil
	}
	return errNotImplemented
}
