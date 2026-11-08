// Package exec is the exec provider. It runs the commands in the exec section of
// glassops.yaml for plan, describe, validate, deploy and verify, so any deploy tool gets the
// gate, the PERMIT and the evidence. It is never detected from the project files. A project
// chooses it with provider: exec.
//
// Every command runs with sh -c in the project dir and gets the GLASSOPS_* variables that
// commandEnv sets. A command may write <results dir>/exec-<step>.json in the shape of
// schemas/engine-result.json. Without that file the outcome comes from the exit code alone.
package exec

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	osexec "os/exec"
	"strings"

	"github.com/glassops-platform/glassops/internal/core/config"
	"github.com/glassops-platform/glassops/internal/providers"
)

// Name is the provider name in glassops.yaml.
const Name = config.ExecProvider

// TierProd is the tier whose targets must have exec.describe, so the guard always runs
// against them. It is the guard's prod tier name.
const TierProd = "prod"

// ErrNoDescribe means exec.describe is not set for a target whose tier does not need it. The
// guard cannot run, and the caller records it as skipped with this reason instead of failing.
var ErrNoDescribe = errors.New("exec.describe is not set, so the target cannot report its Id and the guard is skipped")

// ErrDescribeRequired means exec.describe is not set for a prod tier target. The guard must
// run before anything touches a prod target, so this is always an error.
var ErrDescribeRequired = errors.New("exec.describe is required for a prod tier target, so the guard can check its Id")

// ShellFunc runs command with sh -c in dir with env and returns its exit status. A command
// that cannot start, or that was stopped by a signal or a cancel, returns an error.
type ShellFunc func(ctx context.Context, dir string, env []string, command string, stdout, stderr io.Writer) (int, error)

// GitFunc runs git in dir and returns its stdout.
type GitFunc func(ctx context.Context, dir string, args ...string) ([]byte, error)

// Workspace is what the commands need to know about the run that the provider interface does
// not carry. Commands set it with Bind before calling the provider.
type Workspace struct {
	// Dir is the project root. Commands run here.
	Dir string
	// ResultsDir is where a command may write exec-<step>.json. Empty means
	// <Dir>/.glassops/results.
	ResultsDir string
	// PlanPath and PermitPath are passed on as GLASSOPS_PLAN and GLASSOPS_PERMIT. Empty
	// means the default path under <Dir>/.glassops when that file exists.
	PlanPath   string
	PermitPath string
	// Stdout and Stderr receive the commands' output. Nil means the process's stderr, so a
	// command never mixes its output into the step lines on stdout.
	Stdout io.Writer
	Stderr io.Writer
}

// Provider is the exec provider.
type Provider struct {
	// cfg is the exec section, set by Configure.
	cfg config.Exec
	// loaded is the config Configure was given, so ProjectDir can find its project dir.
	loaded *config.Config
	// ws is the run's workspace, set by Bind.
	ws Workspace
	// deployOverride replaces cfg.Deploy when set, from deploy -- <cmd>.
	deployOverride string

	shell   ShellFunc
	git     GitFunc
	environ func() []string
}

// New returns the exec provider wired to the real sh, git and process environment.
func New() *Provider {
	return &Provider{shell: runShell, git: runGit, environ: os.Environ}
}

// NewWithRunners returns an exec provider that runs commands and git through the given
// functions and starts every command from base as its environment. Tests use it.
func NewWithRunners(shell ShellFunc, git GitFunc, base []string) *Provider {
	return &Provider{shell: shell, git: git, environ: func() []string { return append([]string(nil), base...) }}
}

// The exec provider implements these optional capabilities.
var (
	_ providers.Provider     = (*Provider)(nil)
	_ providers.Verifier     = (*Provider)(nil)
	_ providers.Configurable = (*Provider)(nil)
)

func (p *Provider) Name() string { return Name }

// Detect always reports false, so the exec provider is used only when the config names it.
func (p *Provider) Detect(string) (bool, error) { return false, nil }

// Tools lists sh, which every command runs with. doctor checks git for every project.
func (p *Provider) Tools() []providers.Tool {
	return []providers.Tool{{Name: "sh", Version: func() (string, error) {
		path, err := osexec.LookPath("sh")
		if err != nil {
			return "", fmt.Errorf("sh not found: %w", err)
		}
		return path, nil
	}}}
}

// CanonicalID compares Ids exactly.
func (p *Provider) CanonicalID(id string) string { return id }

// Configure keeps the exec section of cfg.
func (p *Provider) Configure(cfg *config.Config) error {
	p.loaded = cfg
	if cfg == nil {
		p.cfg = config.Exec{}
		return nil
	}
	p.cfg = cfg.Exec
	return nil
}

// ProjectDir returns the dir the CLI loaded cfg from, or "" when it does not know it. The
// CLI sets it, so a command that never calls Bind, such as gate, still runs exec.describe in
// the project dir and never in the process's working dir.
var ProjectDir func(cfg *config.Config) string

// dir is the project dir: the bound one, else the one ProjectDir knows for the config.
func (p *Provider) dir() string {
	if p.ws.Dir != "" {
		return p.ws.Dir
	}
	if ProjectDir != nil && p.loaded != nil {
		return ProjectDir(p.loaded)
	}
	return ""
}

// Bind sets the run's workspace.
func (p *Provider) Bind(ws Workspace) { p.ws = ws }

// OverrideDeploy makes Deploy run command instead of exec.deploy. deploy -- <cmd> sets it.
func (p *Provider) OverrideDeploy(command string) { p.deployOverride = command }

// session is an exec "login". The commands carry their own credentials in the environment,
// so there is nothing to set up or remove.
type session struct{ t providers.Target }

func (s *session) Target() providers.Target { return s.t }
func (s *session) Close() error             { return nil }

// Authenticate returns a session for t. The exec provider logs in to nothing itself. Each
// command reads its own credentials from the environment.
func (p *Provider) Authenticate(_ context.Context, t providers.Target) (providers.Session, error) {
	return &session{t: t}, nil
}

// describeOutput is what exec.describe prints.
type describeOutput struct {
	ID   string `json:"id"`
	Kind string `json:"kind"`
}

// Describe runs exec.describe and reads {"id": "...", "kind": "..."} from its stdout. Without
// exec.describe it returns ErrDescribeRequired for a prod tier target and ErrNoDescribe for
// any other.
func (p *Provider) Describe(ctx context.Context, s providers.Session) (providers.TargetInfo, error) {
	t := s.Target()
	if strings.TrimSpace(p.cfg.Describe) == "" {
		if t.Tier == TierProd {
			return providers.TargetInfo{}, ErrDescribeRequired
		}
		return providers.TargetInfo{}, ErrNoDescribe
	}
	dir := p.dir()
	if dir == "" {
		return providers.TargetInfo{}, errors.New("exec provider: the project dir is not bound, so exec.describe cannot run")
	}
	env, err := p.commandEnv(dir, &t, "", nil)
	if err != nil {
		return providers.TargetInfo{}, err
	}
	var out strings.Builder
	code, err := p.shell(ctx, dir, env, p.cfg.Describe, &out, p.stderr())
	if err != nil {
		return providers.TargetInfo{}, fmt.Errorf("exec.describe: %w", err)
	}
	if code != 0 {
		return providers.TargetInfo{}, fmt.Errorf("exec.describe: exit %d", code)
	}
	dec := json.NewDecoder(strings.NewReader(out.String()))
	dec.DisallowUnknownFields()
	var d describeOutput
	if err := dec.Decode(&d); err != nil {
		return providers.TargetInfo{}, fmt.Errorf("exec.describe: stdout is not {\"id\": \"...\", \"kind\": \"...\"}: %w", err)
	}
	if dec.More() {
		return providers.TargetInfo{}, errors.New("exec.describe: stdout holds more than one JSON value")
	}
	if strings.TrimSpace(d.ID) == "" {
		return providers.TargetInfo{}, errors.New("exec.describe: id is empty")
	}
	kind, err := parseKind(d.Kind)
	if err != nil {
		return providers.TargetInfo{}, fmt.Errorf("exec.describe: %w", err)
	}
	return providers.TargetInfo{ID: d.ID, Kind: kind, Facts: map[string]any{"describe": true}}, nil
}

// parseKind accepts the kinds the guard knows.
func parseKind(k string) (providers.Kind, error) {
	switch providers.Kind(k) {
	case providers.KindProduction, providers.KindSandbox, providers.KindScratch, providers.KindUnknown:
		return providers.Kind(k), nil
	}
	return "", fmt.Errorf("kind is %q, want production, sandbox, scratch or unknown", k)
}

// Validate runs exec.validate. It is an error when exec.validate is not set, because a
// validate that ran nothing must never read as passed.
func (p *Provider) Validate(ctx context.Context, s providers.Session, cs providers.ChangeSet, opts providers.ValidateOptions) (providers.ValidateResult, error) {
	t := s.Target()
	out, err := p.step(ctx, stepValidate, p.cfg.Validate, opts.Dir, &t, cs.Dir, nil)
	if err != nil {
		return providers.ValidateResult{}, err
	}
	return providers.ValidateResult{Passed: out.passed, ID: out.id, Problems: out.problems, Facts: out.facts}, nil
}

// Deploy runs exec.deploy, or the command OverrideDeploy set. The exec provider has no quick
// deploy and no resume, and it waits for the command to finish.
func (p *Provider) Deploy(ctx context.Context, s providers.Session, cs providers.ChangeSet, opts providers.DeployOptions) (providers.DeployResult, error) {
	if opts.ResumeID != "" {
		return providers.DeployResult{}, errors.New("the exec provider cannot resume a deploy: run the deploy again")
	}
	command := p.cfg.Deploy
	if p.deployOverride != "" {
		command = p.deployOverride
	}
	t := s.Target()
	out, err := p.step(ctx, stepDeploy, command, opts.Dir, &t, cs.Dir, nil)
	if err != nil {
		return providers.DeployResult{}, err
	}
	out.facts["exec.deploy_override"] = p.deployOverride != ""
	return providers.DeployResult{
		Passed:             out.passed,
		ID:                 out.id,
		ComponentsDeployed: out.components,
		Problems:           out.problems,
		Facts:              out.facts,
	}, nil
}

// Verify runs exec.verify. It is an error when exec.verify is not set, because a verify that
// ran nothing must never read as passed.
func (p *Provider) Verify(ctx context.Context, s providers.Session, req providers.VerifyRequest) (providers.VerifyResult, error) {
	t := s.Target()
	extra := []string{EnvDeployID + "=" + req.Deploy.ID}
	out, err := p.step(ctx, stepVerify, p.cfg.Verify, req.Dir, &t, req.Plan.Dir, extra)
	if err != nil {
		return providers.VerifyResult{}, err
	}
	check := providers.VerifyCheck{Name: "verify.exec", Status: providers.VerifyPass, Detail: out.summary}
	if !out.passed {
		check.Status = providers.VerifyFail
	}
	checks := []providers.VerifyCheck{check}
	for i, pr := range out.problems {
		checks = append(checks, providers.VerifyCheck{Name: fmt.Sprintf("verify.exec.problem.%d", i+1), Status: providers.VerifyFail, Detail: pr})
	}
	return providers.VerifyResult{Passed: out.passed, Checks: checks, Facts: out.facts}, nil
}

func (p *Provider) stdout() io.Writer {
	if p.ws.Stdout != nil {
		return p.ws.Stdout
	}
	return os.Stderr
}

func (p *Provider) stderr() io.Writer {
	if p.ws.Stderr != nil {
		return p.ws.Stderr
	}
	return os.Stderr
}

// runShell is the real ShellFunc.
func runShell(ctx context.Context, dir string, env []string, command string, stdout, stderr io.Writer) (int, error) {
	cmd := osexec.CommandContext(ctx, "sh", "-c", command)
	cmd.Dir = dir
	cmd.Env = env
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	err := cmd.Run()
	if err == nil {
		return 0, nil
	}
	var ee *osexec.ExitError
	if errors.As(err, &ee) && ee.ExitCode() >= 0 && ctx.Err() == nil {
		return ee.ExitCode(), nil
	}
	if ctx.Err() != nil {
		return -1, fmt.Errorf("stopped: %w", ctx.Err())
	}
	return -1, err
}

// runGit is the real GitFunc.
func runGit(ctx context.Context, dir string, args ...string) ([]byte, error) {
	cmd := osexec.CommandContext(ctx, "git", args...)
	cmd.Dir = dir
	out, err := cmd.Output()
	if err != nil {
		var ee *osexec.ExitError
		if errors.As(err, &ee) && len(ee.Stderr) > 0 {
			return out, fmt.Errorf("git %s: %w: %s", strings.Join(args, " "), err, strings.TrimSpace(string(ee.Stderr)))
		}
		return out, fmt.Errorf("git %s: %w", strings.Join(args, " "), err)
	}
	return out, nil
}
