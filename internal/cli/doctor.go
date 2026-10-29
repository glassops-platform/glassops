package cli

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/spf13/cobra"

	"github.com/glassops-platform/glassops/internal/ci"
	"github.com/glassops-platform/glassops/internal/core/config"
	"github.com/glassops-platform/glassops/internal/core/exitcode"
	"github.com/glassops-platform/glassops/internal/core/run"
	"github.com/glassops-platform/glassops/internal/providers"
)

type doctorOptions struct {
	auth   bool
	target string
}

func newDoctorCommand(app *App, g *globals) *cobra.Command {
	o := &doctorOptions{}
	cmd := &cobra.Command{
		Use:   "doctor",
		Short: "Check config, tools, the git checkout and required environment variables",
		Long: "doctor checks that this project is ready for glassops:\n" +
			"  - glassops.yaml (when present) against the schema and its rules\n" +
			"  - the provider and the tools it needs, with their versions\n" +
			"  - that the checkout is not shallow, which silently empties a delta\n" +
			"  - that every $VAR the config refers to is set (presence only, never values)\n" +
			"With --auth it also logs in to each target (or only --target) and runs the org Id\n" +
			"and tier guard. Without --auth it never logs in.\n" +
			"It exits 1 when any check fails.",
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return app.doctor(cmd.Context(), g, o)
		},
	}
	cmd.Flags().BoolVar(&o.auth, "auth", false, "also log in to targets and run the org Id and tier guard")
	cmd.Flags().StringVar(&o.target, "target", "", "with --auth, check only this target")
	return cmd
}

func (app *App) doctor(ctx context.Context, g *globals, o *doctorOptions) error {
	c, err := app.detectCI(g)
	if err != nil {
		return &exitcode.ExitError{Code: exitcode.Error, Err: err}
	}
	res := run.New("doctor", app.Now)
	res.Facts["ci"] = c.Name()

	// 1. Provider, from the project files.
	detected, err := app.Providers.Detect(g.dir)
	if err != nil {
		res.AddCheck("provider", run.StatusError, err.Error())
	}

	// 2. Config.
	providerName := ""
	if detected != nil {
		providerName = detected.Name()
	}
	loaded, err := config.Load(g.dir, g.configPath, providerName)
	var cfg *config.Config
	switch {
	case err == nil:
		cfg = loaded.Config
		if loaded.Path == "" {
			res.AddCheck("config", run.StatusPass, "no glassops.yaml, using defaults")
		} else {
			res.AddCheck("config", run.StatusPass, loaded.Path)
		}
	case errors.As(err, new(*config.ValidationError)), errors.Is(err, config.ErrNotFound):
		res.AddCheck("config", run.StatusFail, err.Error())
	default:
		res.AddCheck("config", run.StatusError, err.Error())
	}

	// 3. Provider consistency and its tools.
	var p providers.Provider
	if cfg != nil && cfg.Provider != "" {
		p = app.Providers.Get(cfg.Provider)
		if detected == nil {
			res.AddCheck("provider", run.StatusFail, fmt.Sprintf("config says %s but no %s project was found in %s", cfg.Provider, cfg.Provider, g.dir))
		} else if detected.Name() != cfg.Provider {
			res.AddCheck("provider", run.StatusFail, fmt.Sprintf("config says %s but the project looks like %s", cfg.Provider, detected.Name()))
		} else {
			res.AddCheck("provider", run.StatusPass, p.Name())
		}
	} else if detected != nil {
		p = detected
		res.AddCheck("provider", run.StatusPass, p.Name()+" (detected)")
	} else {
		res.AddCheck("provider", run.StatusSkipped, "no provider detected")
	}
	if p != nil {
		res.Provider = p.Name()
	}

	// 4. Tools. git is always needed.
	gitV, gitErr := app.gitVersion(g.dir)
	app.checkTool(res, "git", gitV, gitErr)
	if p != nil {
		for _, t := range p.Tools() {
			v, err := t.Version()
			app.checkTool(res, t.Name, v, err)
		}
	}

	// 5. Shallow clone.
	out, err := app.Run(g.dir, "git", "rev-parse", "--is-shallow-repository")
	switch {
	case err != nil:
		res.AddCheck("git.checkout", run.StatusFail, "not a git repository")
	case strings.TrimSpace(string(out)) == "true":
		res.AddCheck("git.checkout", run.StatusFail, "shallow clone: deltas would be empty. Fetch full history (GitHub: actions/checkout with fetch-depth: 0)")
	default:
		res.AddCheck("git.checkout", run.StatusPass, "full history")
	}

	// 6. Environment variables named in the config. Presence only, never values.
	if cfg != nil {
		var missing []string
		refs := envRefsFor(cfg, o)
		for _, name := range refs {
			if v, ok := app.LookupEnv(name); !ok || v == "" {
				missing = append(missing, name)
			}
		}
		switch {
		case len(refs) == 0:
			res.AddCheck("env", run.StatusSkipped, "the config names no environment variables")
		case len(missing) > 0:
			res.AddCheck("env", run.StatusFail, "not set: "+strings.Join(missing, ", "))
		default:
			res.AddCheck("env", run.StatusPass, fmt.Sprintf("%d variables set", len(refs)))
		}
	}

	// 7. Logins and the guard, only when asked.
	if o.auth {
		app.doctorAuth(ctx, res, cfg, p, o.target)
	}

	return app.finish(res, c, g)
}

// doctorAuth adds one check per target it logs in to, named auth.<target>. A login failure or
// a guard mismatch is an error, which exits 2. Every session is closed.
func (app *App) doctorAuth(ctx context.Context, res *run.Result, cfg *config.Config, p providers.Provider, only string) {
	if cfg == nil || p == nil {
		res.AddCheck("auth", run.StatusSkipped, "no usable config and provider, so there is nothing to log in to")
		return
	}
	var names []string
	if only != "" {
		if _, ok := cfg.Targets[only]; !ok {
			res.AddCheck("auth."+only, run.StatusError, fmt.Errorf("%w: %s", ErrTargetNotConfigured, only).Error())
			return
		}
		names = []string{only}
	} else {
		for name := range cfg.Targets {
			names = append(names, name)
		}
		sort.Strings(names)
	}
	if len(names) == 0 {
		res.AddCheck("auth", run.StatusSkipped, "the config names no targets")
		return
	}
	for _, name := range names {
		check := "auth." + name
		t, err := app.target(cfg, name)
		if err != nil {
			res.AddCheck(check, run.StatusError, err.Error())
			continue
		}
		s, info, err := app.connect(ctx, p, t)
		if err != nil {
			res.AddCheck(check, run.StatusError, err.Error())
			continue
		}
		_ = s.Close()
		res.AddCheck(check, run.StatusPass, string(info.Kind)+" "+info.ID)
	}
}

func (app *App) gitVersion(dir string) (string, error) {
	out, err := app.Run(dir, "git", "--version")
	if err != nil {
		return "", fmt.Errorf("git not found: %w", err)
	}
	// "git version 2.45.0.windows.1"
	f := strings.Fields(string(out))
	if len(f) >= 3 {
		return f[2], nil
	}
	return strings.TrimSpace(string(out)), nil
}

func (app *App) checkTool(res *run.Result, name, version string, err error) {
	if err != nil {
		res.AddCheck("tool."+name, run.StatusFail, err.Error())
		return
	}
	res.Facts["tools."+name] = version
	res.AddCheck("tool."+name, run.StatusPass, version)
}

// finish prints the checks, writes the result and reports to the CI. It turns the status into
// the command's exit code.
func (app *App) finish(res *run.Result, c ci.CI, g *globals) error {
	var md strings.Builder
	fmt.Fprintf(&md, "### glassops %s: %s\n\n| Check | Status | Detail |\n|---|---|---|\n", res.Step, res.Status)
	for _, ch := range res.Checks {
		fmt.Fprintf(app.Stdout, "%-8s %-22s %s\n", ch.Status, ch.Name, ch.Detail)
		fmt.Fprintf(&md, "| %s | %s | %s |\n", ch.Name, ch.Status, strings.ReplaceAll(ch.Detail, "|", "\\|"))
		if ch.Status == run.StatusFail || ch.Status == run.StatusError {
			c.Annotate(ci.Errors, fmt.Sprintf("%s %s: %s", res.Step, ch.Name, ch.Detail))
		}
	}

	path, err := res.Write(g.results(), app.Now)
	if err != nil {
		return &exitcode.ExitError{Code: exitcode.Error, Err: err}
	}
	fmt.Fprintf(app.Stdout, "%s: %s (result: %s)\n", res.Step, res.Status, path)

	if err := c.SetOutput("status", string(res.Status)); err != nil {
		return &exitcode.ExitError{Code: exitcode.Error, Err: err}
	}
	if err := c.SetOutput("result", path); err != nil {
		return &exitcode.ExitError{Code: exitcode.Error, Err: err}
	}
	if c.Name() != "local" {
		if err := c.Summary(md.String()); err != nil {
			return &exitcode.ExitError{Code: exitcode.Error, Err: err}
		}
	}
	if code := res.Status.ExitCode(); code != exitcode.OK {
		return &exitcode.ExitError{Code: code}
	}
	return nil
}

// envRefsFor lists the $VARs the env check covers. With --auth and --target T it covers only
// target T and anything outside targets, so a target that does not exist yet (such as a
// scratch org about to be created) does not fail the check. Otherwise it covers every ref.
func envRefsFor(cfg *config.Config, o *doctorOptions) []string {
	t, ok := cfg.Targets[o.target]
	if !o.auth || o.target == "" || !ok {
		return config.EnvRefs(cfg)
	}
	only := *cfg
	only.Targets = map[string]config.Target{o.target: t}
	outside := *cfg
	outside.Targets = nil
	set := map[string]bool{}
	for _, r := range append(config.EnvRefs(&only), config.EnvRefs(&outside)...) {
		set[r] = true
	}
	out := make([]string, 0, len(set))
	for r := range set {
		out = append(out, r)
	}
	sort.Strings(out)
	return out
}
