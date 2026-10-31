package cli

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/glassops-platform/glassops/internal/ci"
	"github.com/glassops-platform/glassops/internal/core/config"
	"github.com/glassops-platform/glassops/internal/core/exitcode"
	"github.com/glassops-platform/glassops/internal/core/run"
	"github.com/glassops-platform/glassops/internal/providers"
)

// defaultEnvTarget is the target name an ephemeral environment registers as by default.
const defaultEnvTarget = "ephemeral"

// EnvSchema identifies the env.json file env create writes.
const EnvSchema = "glassops.env/v1"

// namePartMax caps the repo and branch parts of a generated environment name, so the run
// part survives when a provider truncates a long name.
const namePartMax = 30

type envCreateOptions struct {
	target  string
	name    string
	ttlDays int
	shape   string
	envFile string
}

type envDeleteOptions struct {
	target  string
	name    string
	envFile string
}

type envSweepOptions struct {
	target string
	maxAge time.Duration
}

// EnvState is env.json: what env create made, so env delete and later steps can find it.
// It never holds login material, only the path of the file that does.
type EnvState struct {
	Schema   string `json:"schema"`
	Provider string `json:"provider"`
	// Target is the environments key the environment was created for.
	Target string `json:"target"`
	// DevHub is the hub target it was created from.
	DevHub   string `json:"devhub"`
	ID       string `json:"id,omitempty"`
	RecordID string `json:"record_id,omitempty"`
	Name     string `json:"name,omitempty"`
	// ExpiresAt is empty when the provider did not report it.
	ExpiresAt *time.Time `json:"expires_at,omitempty"`
	// AuthFile is the path of the 0600 file a later step logs in with.
	AuthFile string    `json:"auth_file,omitempty"`
	Created  time.Time `json:"created"`
	// Ready is false when env create failed after the environment was requested. env delete
	// still removes it.
	Ready bool `json:"ready"`
}

func newEnvCommand(app *App, g *globals) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "env",
		Short: "Create, delete and sweep ephemeral environments",
		Long: "env manages the ephemeral environment configured under environments in\n" +
			"glassops.yaml. Each subcommand logs in to the environment's hub target.",
		Args: cobra.NoArgs,
	}
	cmd.AddCommand(newEnvCreateCommand(app, g), newEnvDeleteCommand(app, g), newEnvSweepCommand(app, g))
	return cmd
}

func newEnvCreateCommand(app *App, g *globals) *cobra.Command {
	o := &envCreateOptions{}
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create an ephemeral environment",
		Long: "env create creates the ephemeral environment for --target from its hub. The login\n" +
			"material is kept in a file only the current user can read, never in a result.\n" +
			"env.json records the environment and that file's path. When the create fails after\n" +
			"the environment was requested, env.json is still written, so env delete can remove it.",
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return app.envCreate(cmd.Context(), g, o)
		},
	}
	f := cmd.Flags()
	f.StringVar(&o.target, "target", defaultEnvTarget, "environment name from the environments section")
	f.StringVar(&o.name, "name", "", "environment name (default glassops-<repo>-<pr>-<run> with the configured prefix)")
	f.IntVar(&o.ttlDays, "ttl-days", 0, "days the environment lives (default ttl_days, else 1)")
	f.StringVar(&o.shape, "shape", "", "source org Id for an org shape (default the configured shape)")
	f.StringVar(&o.envFile, "env-file", "", "where to write env.json (default <dir>/.glassops/env/<target>/env.json)")
	return cmd
}

func newEnvDeleteCommand(app *App, g *globals) *cobra.Command {
	o := &envDeleteOptions{}
	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete an ephemeral environment",
		Long: "env delete deletes the ephemeral environment for --target from its hub. It reads\n" +
			"env.json, or finds the environment by --name. With neither, there is nothing to\n" +
			"delete and it exits 0.",
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return app.envDelete(cmd.Context(), g, o)
		},
	}
	f := cmd.Flags()
	f.StringVar(&o.target, "target", defaultEnvTarget, "environment name from the environments section")
	f.StringVar(&o.name, "name", "", "environment name (default the one env create made)")
	f.StringVar(&o.envFile, "env-file", "", "env.json to read (default <dir>/.glassops/env/<target>/env.json)")
	return cmd
}

func newEnvSweepCommand(app *App, g *globals) *cobra.Command {
	o := &envSweepOptions{}
	cmd := &cobra.Command{
		Use:   "sweep",
		Short: "Delete old ephemeral environments",
		Long: "env sweep deletes every environment on the hub whose name starts with the\n" +
			"configured prefix and that is older than --max-age, and lists what it deleted.",
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return app.envSweep(cmd.Context(), g, o)
		},
	}
	f := cmd.Flags()
	f.StringVar(&o.target, "target", defaultEnvTarget, "environment name from the environments section")
	f.DurationVar(&o.maxAge, "max-age", 0, "delete environments older than this, such as 24h (default ttl_days)")
	return cmd
}

// envFilePath is the env.json path for target.
func envFilePath(g *globals, target, override string) string {
	if override != "" {
		return override
	}
	return filepath.Join(g.dir, ".glassops", "env", target, "env.json")
}

// envSetup is what every env subcommand needs before it logs in to the hub.
type envSetup struct {
	ci   ci.CI
	proj *project
	res  *run.Result
	cfg  config.Environment
	envs providers.Environments
}

// startEnv loads the project and the environment config for target. A problem with either
// is recorded on the result, and ok is false.
func (app *App) startEnv(g *globals, step, target string) (*envSetup, bool, error) {
	c, err := app.detectCI(g)
	if err != nil {
		return nil, false, &exitcode.ExitError{Code: exitcode.Error, Err: err}
	}
	proj, err := app.loadProject(g)
	if err != nil {
		return nil, false, err
	}
	s := &envSetup{ci: c, proj: proj, res: run.New(step, app.Now)}
	s.res.Provider = proj.provider.Name()
	s.res.Target = target
	cfg, found := proj.cfg.Environments[target]
	if !found {
		s.res.AddCheck("config", run.StatusError, fmt.Sprintf("environments.%s is not configured in glassops.yaml", target))
		return s, false, nil
	}
	s.cfg = cfg
	envs, can := proj.provider.(providers.Environments)
	if !can {
		s.res.AddCheck("config", run.StatusError, fmt.Sprintf("provider %s cannot create environments", proj.provider.Name()))
		return s, false, nil
	}
	s.envs = envs
	return s, true, nil
}

// hub logs in to the environment's hub target and runs the guard on it.
func (app *App) hub(ctx context.Context, s *envSetup) (providers.Session, bool) {
	t, err := app.target(s.proj.cfg, s.cfg.DevHub)
	if err != nil {
		s.res.AddCheck("hub", run.StatusError, err.Error())
		return nil, false
	}
	sess, _, err := app.connect(ctx, s.proj.provider, t)
	if err != nil {
		s.res.AddCheck("hub", run.StatusError, err.Error())
		return nil, false
	}
	s.res.AddCheck("hub", run.StatusPass, "logged in to "+t.Name)
	return sess, true
}

func (app *App) envCreate(ctx context.Context, g *globals, o *envCreateOptions) error {
	s, ok, err := app.startEnv(g, "env-create", o.target)
	if err != nil {
		return err
	}
	if !ok {
		return app.finish(s.res, s.ci, g)
	}
	res := s.res
	path := envFilePath(g, o.target, o.envFile)
	res.Facts["env.file"] = path

	if prev, err := readEnvState(path); err == nil {
		res.AddCheck("env.json", run.StatusError, fmt.Sprintf("%s already records environment %q, run env delete first", path, prev.Name))
		return app.finish(res, s.ci, g)
	} else if !errors.Is(err, os.ErrNotExist) {
		res.AddCheck("env.json", run.StatusError, err.Error())
		return app.finish(res, s.ci, g)
	}

	days := o.ttlDays
	if days == 0 {
		days = s.cfg.Days()
	}
	if days < 1 {
		res.AddCheck("ttl", run.StatusError, fmt.Sprintf("--ttl-days must be 1 or more, not %d", days))
		return app.finish(res, s.ci, g)
	}
	shape := o.shape
	if shape == "" {
		shape = s.cfg.Shape
	}
	if m := envRef.FindStringSubmatch(shape); m != nil {
		v, set := app.LookupEnv(m[1])
		if !set || v == "" {
			res.AddCheck("shape", run.StatusError, fmt.Sprintf("the shape is %s, which is not set", shape))
			return app.finish(res, s.ci, g)
		}
		shape = v
	}
	name := o.name
	if name == "" {
		name = envName(s.ci.Context(), s.cfg.NamePrefix(), g.dir, app.Now())
	}

	sess, ok := app.hub(ctx, s)
	if !ok {
		return app.finish(res, s.ci, g)
	}
	defer sess.Close()

	env, cerr := s.envs.CreateEnv(ctx, sess, providers.EnvRequest{
		Name:           name,
		DefinitionFile: s.cfg.Definition,
		Dir:            g.dir,
		TTL:            time.Duration(days) * 24 * time.Hour,
		Shape:          shape,
	})
	for k, v := range env.Facts {
		res.Facts[k] = v
	}
	if env.ID != "" || env.RecordID != "" {
		state := EnvState{
			Schema: EnvSchema, Provider: s.proj.provider.Name(), Target: o.target, DevHub: s.cfg.DevHub,
			ID: env.ID, RecordID: env.RecordID, Name: env.Name, AuthFile: env.AuthFile,
			Created: app.Now().UTC(), Ready: cerr == nil,
		}
		if !env.ExpiresAt.IsZero() {
			t := env.ExpiresAt.UTC()
			state.ExpiresAt = &t
			res.Facts["env.expires_at"] = t.Format(time.RFC3339)
		}
		res.Facts["env.id"] = env.ID
		res.Facts["env.record_id"] = env.RecordID
		res.Facts["env.name"] = env.Name
		res.Facts["env.auth_file"] = env.AuthFile
		if err := writeEnvState(path, state); err != nil {
			res.AddCheck("env.json", run.StatusError, err.Error())
		} else {
			res.Artifacts = append(res.Artifacts, path)
		}
	}
	if cerr != nil {
		detail := cerr.Error()
		if env.ID != "" || env.RecordID != "" {
			detail += ". The request is recorded in env.json, so env delete can remove it"
		}
		res.AddCheck("create", run.StatusError, detail)
		return app.finish(res, s.ci, g)
	}
	detail := fmt.Sprintf("created %s (%s)", env.Name, env.ID)
	if !env.ExpiresAt.IsZero() {
		detail += ", expires " + env.ExpiresAt.UTC().Format("2006-01-02")
	}
	res.AddCheck("create", run.StatusPass, detail)
	for k, v := range map[string]string{"env_id": env.ID, "env_file": path, "auth_file": env.AuthFile} {
		if err := s.ci.SetOutput(k, v); err != nil {
			res.AddError(err)
		}
	}
	return app.finish(res, s.ci, g)
}

func (app *App) envDelete(ctx context.Context, g *globals, o *envDeleteOptions) error {
	s, ok, err := app.startEnv(g, "env-delete", o.target)
	if err != nil {
		return err
	}
	if !ok {
		return app.finish(s.res, s.ci, g)
	}
	res := s.res
	path := envFilePath(g, o.target, o.envFile)
	res.Facts["env.file"] = path

	var env providers.Env
	fromState := false
	state, err := readEnvState(path)
	switch {
	case err == nil:
		if state.Target != o.target {
			res.AddCheck("env.json", run.StatusError, fmt.Sprintf("%s is for target %s, not %s", path, state.Target, o.target))
			return app.finish(res, s.ci, g)
		}
		if state.DevHub != s.cfg.DevHub {
			res.AddCheck("env.json", run.StatusError, fmt.Sprintf("%s was created from hub %s, but environments.%s.devhub is %s", path, state.DevHub, o.target, s.cfg.DevHub))
			return app.finish(res, s.ci, g)
		}
		if o.name == "" || o.name == state.Name {
			env = providers.Env{ID: state.ID, RecordID: state.RecordID, Name: state.Name, AuthFile: state.AuthFile}
			fromState = true
		} else {
			env = providers.Env{Name: o.name}
		}
	case errors.Is(err, os.ErrNotExist):
		if o.name == "" {
			res.AddCheck("delete", run.StatusSkipped, fmt.Sprintf("no %s and no --name, so nothing is recorded to delete", path))
			return app.finish(res, s.ci, g)
		}
		env = providers.Env{Name: o.name}
	default:
		res.AddCheck("env.json", run.StatusError, err.Error())
		return app.finish(res, s.ci, g)
	}
	res.Facts["env.id"] = env.ID
	res.Facts["env.record_id"] = env.RecordID
	res.Facts["env.name"] = env.Name

	sess, ok := app.hub(ctx, s)
	if !ok {
		return app.finish(res, s.ci, g)
	}
	defer sess.Close()

	if err := s.envs.DeleteEnv(ctx, sess, env); err != nil {
		res.AddCheck("delete", run.StatusError, err.Error())
		return app.finish(res, s.ci, g)
	}
	label := env.Name
	if label == "" {
		label = "record " + env.RecordID
	}
	res.AddCheck("delete", run.StatusPass, "deleted "+label+", or it was already gone")
	if fromState {
		// The provider removes an auth file it made. This also covers one it did not, as long
		// as it sits where an auth file may, so a later step falls back to the config.
		if env.AuthFile != "" && authFileAllowed(g.dir, env.AuthFile) == nil {
			if err := os.Remove(env.AuthFile); err != nil && !errors.Is(err, os.ErrNotExist) {
				res.AddCheck("auth_file", run.StatusError, fmt.Sprintf("remove the auth file: %v", err))
			}
		}
		if err := os.Remove(path); err != nil && !errors.Is(err, os.ErrNotExist) {
			res.AddCheck("env.json", run.StatusError, fmt.Sprintf("remove %s: %v", path, err))
		}
	}
	return app.finish(res, s.ci, g)
}

func (app *App) envSweep(ctx context.Context, g *globals, o *envSweepOptions) error {
	s, ok, err := app.startEnv(g, "env-sweep", o.target)
	if err != nil {
		return err
	}
	if !ok {
		return app.finish(s.res, s.ci, g)
	}
	res := s.res
	maxAge := o.maxAge
	if maxAge == 0 {
		maxAge = time.Duration(s.cfg.Days()) * 24 * time.Hour
	}
	if maxAge < 0 {
		res.AddCheck("max-age", run.StatusError, "--max-age must be above zero")
		return app.finish(res, s.ci, g)
	}
	prefix := s.cfg.NamePrefix()
	res.Facts["env.prefix"] = prefix
	res.Facts["env.max_age"] = maxAge.String()

	sess, ok := app.hub(ctx, s)
	if !ok {
		return app.finish(res, s.ci, g)
	}
	defer sess.Close()

	deleted, serr := s.envs.SweepEnvs(ctx, sess, prefix, maxAge)
	list := make([]map[string]any, 0, len(deleted))
	for i, e := range deleted {
		item := map[string]any{"id": e.ID, "name": e.Name}
		if !e.ExpiresAt.IsZero() {
			item["expires_at"] = e.ExpiresAt.UTC().Format(time.RFC3339)
		}
		list = append(list, item)
		res.AddCheck(fmt.Sprintf("deleted.%d", i+1), run.StatusPass, fmt.Sprintf("%s (%s)", e.Name, e.ID))
	}
	res.Facts["env.deleted"] = list
	res.Facts["env.deleted_count"] = len(deleted)
	if serr != nil {
		res.AddCheck("sweep", run.StatusError, serr.Error())
		return app.finish(res, s.ci, g)
	}
	res.AddCheck("sweep", run.StatusPass, fmt.Sprintf("deleted %d environments named %s-* older than %s", len(deleted), prefix, maxAge))
	return app.finish(res, s.ci, g)
}

var nameJunk = regexp.MustCompile(`[^a-z0-9]+`)

// envName builds <prefix>-<repo>-<pr or branch>-<run>. Outside CI the repo is the project
// dir's name and the run is a UTC timestamp.
func envName(c ci.Context, prefix, dir string, now time.Time) string {
	clean := func(s string, limit int) string {
		s = strings.Trim(nameJunk.ReplaceAllString(strings.ToLower(s), "-"), "-")
		if limit > 0 && len(s) > limit {
			s = strings.TrimRight(s[:limit], "-")
		}
		return s
	}
	repo := c.Repo
	if i := strings.LastIndex(repo, "/"); i >= 0 {
		repo = repo[i+1:]
	}
	if repo == "" {
		if abs, err := filepath.Abs(dir); err == nil {
			repo = filepath.Base(abs)
		}
	}
	ref := c.PRNumber
	if ref == "" {
		ref = strings.TrimPrefix(strings.TrimPrefix(c.Ref, "refs/heads/"), "refs/tags/")
	}
	runID := c.RunID
	if runID == "" {
		runID = now.UTC().Format("20060102t150405")
	}
	parts := []string{prefix}
	for _, p := range []string{clean(repo, namePartMax), clean(ref, namePartMax), clean(runID, 0)} {
		if p != "" {
			parts = append(parts, p)
		}
	}
	return strings.Join(parts, "-")
}

// readEnvState reads env.json. A missing file wraps os.ErrNotExist.
func readEnvState(path string) (EnvState, error) {
	var s EnvState
	data, err := os.ReadFile(path)
	if err != nil {
		return s, err
	}
	if err := json.Unmarshal(data, &s); err != nil {
		return s, fmt.Errorf("parse %s: %w", path, err)
	}
	if s.Schema != EnvSchema {
		return s, fmt.Errorf("%s: schema is %q, want %q", path, s.Schema, EnvSchema)
	}
	return s, nil
}

func writeEnvState(path string, s EnvState) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("create %s: %w", filepath.Dir(path), err)
	}
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return fmt.Errorf("encode env.json: %w", err)
	}
	if err := os.WriteFile(path, append(data, '\n'), 0o644); err != nil {
		return fmt.Errorf("write %s: %w", path, err)
	}
	return nil
}
