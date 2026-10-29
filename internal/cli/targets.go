package cli

import (
	"errors"
	"fmt"
	"regexp"

	"github.com/glassops-platform/glassops/internal/core/config"
	"github.com/glassops-platform/glassops/internal/core/exitcode"
	"github.com/glassops-platform/glassops/internal/providers"
)

var envRef = regexp.MustCompile(`^\$([A-Z_][A-Z0-9_]*)$`)

// ErrTargetNotConfigured means glassops.yaml has no target by that name.
var ErrTargetNotConfigured = errors.New("target not configured")

// project is the config and provider a command works with.
type project struct {
	cfg      *config.Config
	provider providers.Provider
}

// loadProject loads glassops.yaml and picks its provider, or the detected one. Every error
// is an *exitcode.ExitError with code 2, because a command cannot start without both.
func (app *App) loadProject(g *globals) (*project, error) {
	detected, err := app.Providers.Detect(g.dir)
	if err != nil {
		return nil, &exitcode.ExitError{Code: exitcode.Error, Err: err}
	}
	name := ""
	if detected != nil {
		name = detected.Name()
	}
	loaded, err := config.Load(g.dir, g.configPath, name)
	if err != nil {
		return nil, &exitcode.ExitError{Code: exitcode.Error, Err: err}
	}
	cfg := loaded.Config
	if cfg.Provider == "" {
		return nil, &exitcode.ExitError{Code: exitcode.Error, Err: fmt.Errorf("no provider: no glassops.yaml names one and none was detected in %s", g.dir)}
	}
	p := app.Providers.Get(cfg.Provider)
	if p == nil {
		return nil, &exitcode.ExitError{Code: exitcode.Error, Err: fmt.Errorf("unknown provider %q", cfg.Provider)}
	}
	return &project{cfg: cfg, provider: p}, nil
}

// target resolves the named target from the config. A $VAR Id is expanded from the
// environment, and an unset variable is an error, so the guard never compares against "".
// A missing target wraps ErrTargetNotConfigured.
func (app *App) target(cfg *config.Config, name string) (providers.Target, error) {
	t, ok := cfg.Targets[name]
	if !ok {
		return providers.Target{}, fmt.Errorf("%w: %s", ErrTargetNotConfigured, name)
	}
	id := t.ID
	if m := envRef.FindStringSubmatch(id); m != nil {
		v, ok := app.LookupEnv(m[1])
		if !ok || v == "" {
			return providers.Target{}, fmt.Errorf("target %s: id is %s, which is not set", name, id)
		}
		id = v
	}
	return providers.Target{Name: name, ID: id, Tier: t.Tier, Branch: t.Branch, Auth: t.Auth}, nil
}
