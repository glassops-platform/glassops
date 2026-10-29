package cli

import (
	"context"
	"errors"
	"fmt"

	"github.com/glassops-platform/glassops/internal/core/guard"
	"github.com/glassops-platform/glassops/internal/providers"
)

// ErrLogin marks a failure to log in, as opposed to a guard mismatch. validate uses it to
// decide whether a fallback target applies.
var ErrLogin = errors.New("login failed")

// connect logs in to t, reads what the target reports and runs the guard. On any error the
// session is already closed. A guard mismatch comes back as *guard.MismatchError, and the
// caller must stop before doing anything else with the target.
func (app *App) connect(ctx context.Context, p providers.Provider, t providers.Target) (providers.Session, providers.TargetInfo, error) {
	s, err := p.Authenticate(ctx, t)
	if err != nil {
		return nil, providers.TargetInfo{}, fmt.Errorf("%w: target %s: %w", ErrLogin, t.Name, err)
	}
	info, err := p.Describe(ctx, s)
	if err != nil {
		_ = s.Close()
		return nil, providers.TargetInfo{}, fmt.Errorf("describe target %s: %w", t.Name, err)
	}
	if err := guard.Check(t, info, p.CanonicalID); err != nil {
		_ = s.Close()
		return nil, info, err
	}
	return s, info, nil
}
