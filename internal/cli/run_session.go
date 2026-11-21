package cli

import (
	"context"
	"errors"
	"fmt"
	"io"

	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/glassops-platform/glassops/internal/ci"
	"github.com/glassops-platform/glassops/internal/core/guard"
	"github.com/glassops-platform/glassops/internal/core/permit"
	"github.com/glassops-platform/glassops/internal/core/pipeline"
	"github.com/glassops-platform/glassops/internal/core/run"
	"github.com/glassops-platform/glassops/internal/providers"
)

// sessionShellRunner runs an exec step's command that holds a session. It is shellRunner in
// its own process group, and on Linux and macOS it kills the whole group when the command
// ends, so nothing the command started keeps the login. Tests replace it.
var sessionShellRunner = func(ctx context.Context, dir string, env []string, command string, stdout, stderr io.Writer) (int, error) {
	cmd := exec.CommandContext(ctx, "sh", "-c", command)
	cmd.Dir = dir
	cmd.Env = env
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	setProcessGroup(cmd)
	cmd.Cancel = func() error { return killProcessGroup(cmd) }
	cmd.WaitDelay = 10 * time.Second
	err := cmd.Run()
	_ = killProcessGroup(cmd)
	if errors.Is(err, exec.ErrWaitDelay) && cmd.ProcessState != nil && ctx.Err() == nil {
		// The command exited but something it started held the output open. That process
		// group is killed above, and the command's own exit code stands.
		return cmd.ProcessState.ExitCode(), nil
	}
	if err == nil {
		return 0, nil
	}
	var ee *exec.ExitError
	if errors.As(err, &ee) && ee.ExitCode() >= 0 && ctx.Err() == nil {
		return ee.ExitCode(), nil
	}
	if ctx.Err() != nil {
		return -1, fmt.Errorf("stopped: %w", ctx.Err())
	}
	return -1, err
}

// sessionStep is what an exec step with a target needs to run its command.
type sessionStep struct {
	sess providers.Session
	// env is added to the command's environment. It points the provider's tools at the
	// session and holds no secret.
	env []string
	// revoker ends the session's token at the target after the command, or is nil when the
	// provider cannot.
	revoker providers.SessionRevoker
}

// revoke ends the session's token so a copy the command made stops working. A failure is a
// warning: the command has run, and the token still expires on its own.
func (ss *sessionStep) revoke(ctx context.Context, c ci.CI, res *run.Result) {
	if ss.revoker == nil {
		return
	}
	// A cancelled run still revokes, with a short budget of its own.
	rctx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 30*time.Second)
	defer cancel()
	if err := ss.revoker.RevokeSession(rctx, ss.sess); err != nil {
		res.Facts["exec.token_revoked"] = false
		msg := fmt.Sprintf("the session's token was not revoked: %v. It still expires on its own", err)
		res.AddCheck("revoke", run.StatusSkipped, msg)
		c.Annotate(ci.Warning, "exec: "+msg)
		return
	}
	res.Facts["exec.token_revoked"] = true
	res.AddCheck("revoke", run.StatusPass, "the session's token was revoked at the target")
}

// execSession prepares an exec step that names a target. On any tier but ephemeral it first
// checks, before any login, that a PERMIT allows this target, commit, plan and config and binds
// the step's command, and that the working tree is the PERMIT's commit. Then it logs in, runs
// the guard and asks the provider for the session's environment. Each failure adds a check to
// res and returns an error. The caller closes the session.
func (app *App) execSession(ctx context.Context, g *globals, c ci.CI, s pipeline.Step, res *run.Result) (*sessionStep, error) {
	fail := func(check string, err error) (*sessionStep, error) {
		res.AddCheck(check, run.StatusError, err.Error())
		return nil, err
	}
	proj, err := app.loadProject(g)
	if err != nil {
		return fail("config", err)
	}
	res.Provider = proj.provider.Name()
	enver, ok := proj.provider.(providers.SessionEnver)
	if !ok {
		return fail("session", fmt.Errorf("provider %s cannot hand a session to an exec step, so target %s is not allowed on it", proj.provider.Name(), s.Target))
	}
	t, err := app.target(proj.cfg, s.Target)
	if err != nil {
		return fail("target", err)
	}
	res.Target = t.Name
	res.Facts["exec.target"] = t.Name
	res.Facts["exec.tier"] = t.Tier

	if t.Tier == guard.TierEphemeral {
		res.AddCheck("permit", run.StatusSkipped, fmt.Sprintf("target %s is tier %s, which needs only the guard", t.Name, guard.TierEphemeral))
	} else {
		cs, err := providers.ReadChangeSet(filepath.Join(g.dir, ".glassops", "plan.json"))
		if err != nil {
			return fail("plan", err)
		}
		if err := providers.CheckDigest(g.dir, cs); err != nil {
			return fail("plan", err)
		}
		cfgSHA, err := app.configSHA256(g, proj)
		if err != nil {
			return fail("config", err)
		}
		commit, err := app.deployCommit(g, c)
		if err != nil {
			return fail("commit", err)
		}
		pm, err := app.verifyPermit(&deployOptions{}, g, proj.cfg, permit.Expect{
			TargetName:        t.Name,
			TargetID:          proj.provider.CanonicalID(t.ID),
			Commit:            commit,
			PlanSHA256:        cs.SHA256,
			ConfigSHA256:      cfgSHA,
			ChangeSetSHA256:   permit.ChangeSetDigest(cs),
			ExecCommandSHA256: permit.ExecCommandDigest(s.Exec),
			Now:               app.Now(),
			MaxTTL:            proj.cfg.Policy.TTL(),
			CanonicalID:       proj.provider.CanonicalID,
		})
		if pm.ID != "" {
			res.Facts["exec.permit_id"] = pm.ID
		}
		if err != nil {
			return fail("permit", err)
		}
		res.AddCheck("permit", run.StatusPass, fmt.Sprintf("PERMIT %s binds this command for %s at %s", pm.ID, t.Name, short(commit)))
		// The command runs from the working tree, so a script file it names is bound only
		// when the tree is the PERMIT's commit.
		if err := app.checkTree(g, pm.Commit); err != nil {
			return fail("tree", err)
		}
		res.AddCheck("tree", run.StatusPass, fmt.Sprintf("HEAD is %s and no tracked file is modified", short(pm.Commit)))
	}

	sess, info, err := app.connect(ctx, proj.provider, t)
	if err != nil {
		return fail("guard", err)
	}
	res.AddCheck("guard", run.StatusPass, fmt.Sprintf("logged in to %s, tier %s, kind %s", t.Name, t.Tier, info.Kind))
	env, err := enver.SessionEnv(sess)
	if err != nil {
		_ = sess.Close()
		return fail("session", err)
	}
	res.Facts["exec.session"] = true
	revoker, _ := proj.provider.(providers.SessionRevoker)
	return &sessionStep{sess: sess, env: env, revoker: revoker}, nil
}

// withoutPermitKey drops GLASSOPS_PERMIT_KEY from env, so a command that holds a session can
// never sign its own PERMIT.
func withoutPermitKey(env []string) []string {
	out := make([]string, 0, len(env))
	for _, kv := range env {
		if !strings.HasPrefix(kv, permit.KeyEnv+"=") {
			out = append(out, kv)
		}
	}
	return out
}
