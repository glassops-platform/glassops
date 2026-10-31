package cli

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/glassops-platform/glassops/internal/ci"
	"github.com/glassops-platform/glassops/internal/core/config"
	"github.com/glassops-platform/glassops/internal/core/exitcode"
	"github.com/glassops-platform/glassops/internal/core/gitref"
	"github.com/glassops-platform/glassops/internal/core/permit"
	"github.com/glassops-platform/glassops/internal/core/run"
	"github.com/glassops-platform/glassops/internal/providers"
)

type deployOptions struct {
	target     string
	planPath   string
	permitPath string
	wait       time.Duration
	resumeID   string
}

// minPermitKeyBytes is the shortest HMAC key deploy accepts, after base64 decoding.
const minPermitKeyBytes = 32

// permitFuncs is how deploy reads and checks a PERMIT. Tests replace it.
var permitFuncs = struct {
	read   func(path string) (permit.Permit, error)
	verify func(p permit.Permit, s permit.Signer, want permit.Expect) error
	signer func(key []byte) permit.Signer
}{permit.Read, permit.Verify, permit.NewHMAC}

func newDeployCommand(app *App, g *globals) *cobra.Command {
	o := &deployOptions{}
	cmd := &cobra.Command{
		Use:   "deploy",
		Short: "Deploy the plan to a target under a verified PERMIT",
		Long: "deploy verifies the PERMIT before it logs in, then runs the org Id and tier guard\n" +
			"and deploys the change set in .glassops/plan.json. When the PERMIT holds a\n" +
			"validated deploy Id that the target still accepts, deploy quick deploys it.\n" +
			"Success is read from the target's result body. When --wait runs out, the result\n" +
			"keeps the job Id, and --resume picks the same deploy up again. After a deploy\n" +
			"succeeds, the git tag glassops/deployed/<target> is moved to the deployed commit\n" +
			"and pushed. A failed push is a warning, because the deploy has already happened.",
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return app.deploy(cmd.Context(), g, o)
		},
	}
	f := cmd.Flags()
	f.StringVar(&o.target, "target", "", "target name from glassops.yaml (required)")
	f.StringVar(&o.planPath, "plan", "", "change set file (default <dir>/.glassops/plan.json)")
	f.StringVar(&o.permitPath, "permit", "", "PERMIT file (default <dir>/.glassops/permit.json)")
	f.DurationVar(&o.wait, "wait", 0, "how long to wait for the deploy, such as 30m (default the provider's)")
	f.StringVar(&o.resumeID, "resume", "", "job Id of an earlier deploy to resume")
	_ = cmd.MarkFlagRequired("target")
	return cmd
}

// deploy verifies the PERMIT, logs in, runs the guard and deploys. Nothing touches the target
// until the PERMIT has verified against this target, commit, plan and config.
func (app *App) deploy(ctx context.Context, g *globals, o *deployOptions) error {
	c, err := app.detectCI(g)
	if err != nil {
		return &exitcode.ExitError{Code: exitcode.Error, Err: err}
	}
	proj, err := app.loadProject(g)
	if err != nil {
		return err
	}
	res := run.New("deploy", app.Now)
	res.Provider = proj.provider.Name()
	res.Target = o.target
	stop := func(check string, err error) error {
		res.AddCheck(check, run.StatusError, err.Error())
		return app.finish(res, c, g)
	}

	t, err := app.target(proj.cfg, o.target)
	if err != nil {
		return stop("target", err)
	}
	planPath := o.planPath
	if planPath == "" {
		planPath = filepath.Join(g.dir, ".glassops", "plan.json")
	}
	cs, err := providers.ReadChangeSet(planPath)
	if err != nil {
		return stop("plan", err)
	}
	cfgSHA, err := app.configSHA256(g, proj)
	if err != nil {
		return stop("config", err)
	}
	commit, err := app.deployCommit(g, c)
	if err != nil {
		return stop("commit", err)
	}

	pm, err := app.verifyPermit(o, g, permit.Expect{
		TargetName:   t.Name,
		TargetID:     proj.provider.CanonicalID(t.ID),
		Commit:       commit,
		PlanSHA256:   cs.SHA256,
		ConfigSHA256: cfgSHA,
		Now:          app.Now(),
	})
	if pm.ID != "" {
		res.Facts["deploy.permit_id"] = pm.ID
	}
	if err != nil {
		return stop("permit", err)
	}
	res.AddCheck("permit", run.StatusPass, fmt.Sprintf("PERMIT %s verified for %s at %s", pm.ID, t.Name, short(commit)))

	if cs.Empty && o.resumeID == "" {
		res.Status = run.StatusEmpty
		res.AddCheck("deploy", run.StatusSkipped, "the change set is empty, nothing to deploy")
		return app.finish(res, c, g)
	}

	opts := providers.DeployOptions{
		Dir:       g.dir,
		TestLevel: testLevelFor(proj.cfg.Salesforce.TestLevel, t.Tier),
		Wait:      o.wait,
		ResumeID:  o.resumeID,
	}
	validated := ""
	if pm.Validation != nil && pm.Validation.Passed {
		validated = pm.Validation.DeployID
	}
	mode := proj.cfg.Salesforce.Deploy.QuickMode()
	res.Facts["deploy.quick_mode"] = mode
	if o.resumeID == "" {
		switch mode {
		case config.QuickOff:
		case config.QuickRequire:
			if validated == "" {
				res.AddCheck("quick", run.StatusFail, "salesforce.deploy.quick is require and the PERMIT holds no passed validation to quick deploy")
				res.Degrade(run.StatusFail)
				return app.finish(res, c, g)
			}
			opts.ValidatedID, opts.QuickOnly = validated, true
		default:
			opts.ValidatedID = validated
		}
	}

	sess, _, err := app.connect(ctx, proj.provider, t)
	if err != nil {
		return stop("target", err)
	}
	defer sess.Close()

	out, err := proj.provider.Deploy(ctx, sess, cs, opts)
	if err != nil {
		return stop("deploy", err)
	}
	for k, v := range out.Facts {
		res.Facts[k] = v
	}
	res.Facts["deploy.id"] = out.ID
	res.Facts["deploy.quick"] = out.Quick
	res.Facts["deploy.timed_out"] = out.TimedOut
	res.Facts["deploy.commit"] = commit

	if out.TimedOut {
		return stop("deploy", fmt.Errorf("deploy %s did not finish within the wait. Run glassops deploy --target %s --resume %s", out.ID, t.Name, out.ID))
	}
	for i, p := range out.Problems {
		res.AddCheck(fmt.Sprintf("problem.%d", i+1), run.StatusFail, p)
	}
	if !out.Passed {
		res.Degrade(run.StatusFail)
		return app.finish(res, c, g)
	}
	how := "full deploy"
	if out.Quick {
		how = "quick deploy of validation " + opts.ValidatedID
	}
	res.AddCheck("deploy", run.StatusPass, fmt.Sprintf("deploy %s succeeded, %d components (%s)", out.ID, out.ComponentsDeployed, how))

	app.moveDeployedTag(res, c, g, t.Name, commit)
	return app.finish(res, c, g)
}

// verifyPermit reads the PERMIT and checks it against want with the key from the
// environment. The returned PERMIT is filled in whenever it could be read, so its Id can be
// recorded even when it does not verify. The key itself never reaches an error.
func (app *App) verifyPermit(o *deployOptions, g *globals, want permit.Expect) (permit.Permit, error) {
	key, err := app.permitKey()
	if err != nil {
		return permit.Permit{}, err
	}
	path := o.permitPath
	if path == "" {
		path = filepath.Join(g.dir, ".glassops", "permit.json")
	}
	pm, err := permitFuncs.read(path)
	if err != nil {
		return permit.Permit{}, fmt.Errorf("read PERMIT: %w", err)
	}
	if err := permitFuncs.verify(pm, permitFuncs.signer(key), want); err != nil {
		return pm, fmt.Errorf("PERMIT %s does not allow this deploy: %w", pm.ID, err)
	}
	return pm, nil
}

// permitKey reads the HMAC key from GLASSOPS_PERMIT_KEY, which holds standard base64.
func (app *App) permitKey() ([]byte, error) {
	v := strings.TrimSpace(app.Getenv(permit.KeyEnv))
	if v == "" {
		return nil, fmt.Errorf("%s is not set, so the PERMIT cannot be verified", permit.KeyEnv)
	}
	key, err := base64.StdEncoding.DecodeString(v)
	if err != nil {
		return nil, fmt.Errorf("%s is not valid base64", permit.KeyEnv)
	}
	if len(key) < minPermitKeyBytes {
		return nil, fmt.Errorf("%s is %d bytes after decoding, want at least %d", permit.KeyEnv, len(key), minPermitKeyBytes)
	}
	return key, nil
}

// configSHA256 is the hash a PERMIT binds the config by. It loads the config again, because
// the project keeps only the parsed config and not the bytes it came from.
func (app *App) configSHA256(g *globals, proj *project) (string, error) {
	loaded, err := config.Load(g.dir, g.configPath, proj.cfg.Provider)
	if err != nil {
		return "", err
	}
	return loaded.SHA256(), nil
}

// deployCommit is the commit being deployed: the CI's commit, or HEAD when the CI reports
// none.
func (app *App) deployCommit(g *globals, c ci.CI) (string, error) {
	if sha := strings.TrimSpace(c.Context().SHA); sha != "" {
		return sha, nil
	}
	out, err := app.Run(g.dir, "git", "rev-parse", "HEAD")
	sha := strings.TrimSpace(string(out))
	if err != nil || sha == "" {
		return "", errors.New("cannot tell which commit is being deployed: the CI reports none and git rev-parse HEAD failed")
	}
	return sha, nil
}

// moveDeployedTag points glassops/deployed/<target> at the deployed commit and pushes it. Any
// failure here is a warning and leaves the step passing, because the deploy has happened.
func (app *App) moveDeployedTag(res *run.Result, c ci.CI, g *globals, target, commit string) {
	git := func(dir string, args ...string) (string, error) {
		out, err := app.Run(dir, "git", args...)
		return strings.TrimSpace(string(out)), err
	}
	res.Facts["deploy.tag_pushed"] = false
	tag, _ := gitref.DeployedTag(target)
	err := gitref.MoveDeployedTag(git, g.dir, target, commit)
	if err == nil {
		err = gitref.PushDeployedTag(git, g.dir, gitref.DefaultRemote, target)
	}
	if err != nil {
		msg := fmt.Sprintf("the deploy succeeded but the tag was not updated: %s. The next plan for %s will not start from this deploy until the tag is pushed", err, target)
		res.AddCheck("tag", run.StatusSkipped, msg)
		c.Annotate(ci.Warning, "deploy: "+msg)
		return
	}
	res.Facts["deploy.tag"] = tag
	res.Facts["deploy.tag_pushed"] = true
	res.AddCheck("tag", run.StatusPass, fmt.Sprintf("%s moved to %s and pushed", tag, short(commit)))
}
