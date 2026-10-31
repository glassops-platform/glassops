package cli

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"sync"

	"github.com/glassops-platform/glassops/internal/core/config"
	"github.com/glassops-platform/glassops/internal/core/exitcode"
	"github.com/glassops-platform/glassops/internal/core/guard"
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
	projectDirs.Store(cfg, g.dir)
	return &project{cfg: cfg, provider: p}, nil
}

// projectDirs maps each config loadProject returned to its project dir, so target can find
// an environment's env.json without every caller passing the dir along.
var projectDirs sync.Map

// target resolves the named target from the config. A $VAR Id is expanded from the
// environment, and an unset variable is an error, so the guard never compares against "".
// A missing target, or one whose Id variable is unset, wraps ErrTargetNotConfigured.
//
// A target named under environments whose env.json exists (written by env create) takes the
// environment's org Id and logs in with its auth file. The config's Id and auth are ignored
// for it while the file exists. With no env.json, the config is used as written.
func (app *App) target(cfg *config.Config, name string) (providers.Target, error) {
	t, ok := cfg.Targets[name]
	if et, found, err := envTarget(cfg, name, t); found || err != nil {
		return et, err
	}
	if !ok {
		return providers.Target{}, fmt.Errorf("%w: %s", ErrTargetNotConfigured, name)
	}
	id := t.ID
	if m := envRef.FindStringSubmatch(id); m != nil {
		v, ok := app.LookupEnv(m[1])
		if !ok || v == "" {
			// An unset Id means the target does not exist yet, for example a scratch org
			// that was never created, so it counts as not configured.
			return providers.Target{}, fmt.Errorf("%w: target %s: id is %s, which is not set", ErrTargetNotConfigured, name, id)
		}
		id = v
	}
	return providers.Target{Name: name, ID: id, Tier: t.Tier, Branch: t.Branch, Auth: t.Auth}, nil
}

// envTarget resolves name from its env.json when name is under environments and the file
// exists. found is false when the config should be used instead.
func envTarget(cfg *config.Config, name string, t config.Target) (providers.Target, bool, error) {
	if _, isEnv := cfg.Environments[name]; !isEnv {
		return providers.Target{}, false, nil
	}
	v, known := projectDirs.Load(cfg)
	if !known {
		return providers.Target{}, false, nil
	}
	dir := v.(string)
	path := filepath.Join(dir, ".glassops", "env", name, "env.json")
	st, err := readEnvState(path)
	if errors.Is(err, os.ErrNotExist) {
		return providers.Target{}, false, nil
	}
	if err != nil {
		return providers.Target{}, true, fmt.Errorf("target %s: %w", name, err)
	}
	switch {
	case st.Target != name:
		return providers.Target{}, true, fmt.Errorf("target %s: %s is for target %s", name, path, st.Target)
	case !st.Ready || st.ID == "" || st.AuthFile == "":
		// A create that failed after the request leaves a record for env delete, but nothing
		// to log in with. A fallback target still applies.
		return providers.Target{}, true, fmt.Errorf("%w: target %s: %s records an environment that was not created", ErrTargetNotConfigured, name, path)
	}
	if err := authFileAllowed(dir, st.AuthFile); err != nil {
		return providers.Target{}, true, fmt.Errorf("target %s: %w", name, err)
	}
	tier := t.Tier
	if tier == "" {
		tier = guard.TierEphemeral
	}
	return providers.Target{
		Name: name, ID: st.ID, Tier: tier, Branch: t.Branch,
		Auth:     config.Auth{Method: "sfdx-url"},
		AuthFile: st.AuthFile,
	}, true, nil
}

// authFileAllowed refuses an auth file outside the OS temp dir and <dir>/.glassops, so an
// edited env.json cannot point a login at an arbitrary file.
func authFileAllowed(dir, file string) error {
	if !filepath.IsAbs(file) {
		return fmt.Errorf("the auth file %s is not an absolute path", file)
	}
	real, err := filepath.EvalSymlinks(file)
	if err != nil {
		return fmt.Errorf("the auth file %s cannot be read: %w", file, err)
	}
	for _, root := range []string{os.TempDir(), filepath.Join(dir, ".glassops")} {
		if r, err := filepath.EvalSymlinks(root); err == nil && within(r, real) {
			return nil
		}
	}
	return fmt.Errorf("the auth file %s is outside the temp dir and .glassops, refusing it", file)
}

// within reports whether path is strictly inside root. Windows paths compare without case.
func within(root, path string) bool {
	if runtime.GOOS == "windows" {
		root, path = strings.ToLower(root), strings.ToLower(path)
	}
	rel, err := filepath.Rel(root, path)
	return err == nil && rel != "." && rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator)) && !filepath.IsAbs(rel)
}
