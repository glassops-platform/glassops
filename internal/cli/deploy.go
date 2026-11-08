package cli

import (
	"context"
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
	execprovider "github.com/glassops-platform/glassops/internal/providers/exec"
)

type deployOptions struct {
	target     string
	planPath   string
	permitPath string
	wait       time.Duration
	resumeID   string
	// command is what follows --, which replaces exec.deploy for provider exec.
	command []string
}

// permitFuncs is how deploy reads and checks a PERMIT. Tests replace it.
var permitFuncs = struct {
	read   func(path string) (permit.Permit, error)
	verify func(p permit.Permit, s permit.Signer, want permit.Expect) error
	signer func(key []byte) permit.Signer
}{permit.Read, permit.Verify, permit.NewHMAC}

func newDeployCommand(app *App, g *globals) *cobra.Command {
	o := &deployOptions{}
	cmd := &cobra.Command{
		Use:   "deploy [-- command...]",
		Short: "Deploy the plan to a target under a verified PERMIT",
		Long: "deploy verifies the PERMIT before it logs in, then runs the org Id and tier guard\n" +
			"and deploys the change set in .glassops/plan.json. When the PERMIT holds a\n" +
			"validated deploy Id that the target still accepts, deploy quick deploys it.\n" +
			"Success is read from the target's result body. When --wait runs out, the result\n" +
			"keeps the job Id, and --resume picks the same deploy up again, only when\n" +
			"results/deploy.json records that job timing out under the same PERMIT. deploy\n" +
			"refuses unless git HEAD is the PERMIT's commit and no tracked file is modified.\n" +
			"After a deploy succeeds, the git tag glassops/deployed/<target> is moved to the\n" +
			"deployed commit and pushed. A failed push is a warning, because the deploy has\n" +
			"already happened.\n" +
			"With provider exec, a command after -- runs instead of exec.deploy, with each\n" +
			"argument passed as one word. The PERMIT must bind that command, from gate\n" +
			"--deploy-command, and the result records exec.deploy_override.",
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) > 0 && cmd.ArgsLenAtDash() != 0 {
				return fmt.Errorf("deploy takes no arguments, except a command after --")
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			o.command = args
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
	if len(o.command) > 0 {
		ov, ok := proj.provider.(deployOverrider)
		if !ok {
			return &exitcode.ExitError{Code: exitcode.Error, Err: fmt.Errorf("a command after -- needs provider exec, and this project uses %s", proj.provider.Name())}
		}
		ov.OverrideDeploy(shellJoin(o.command))
	}
	res := run.New("deploy", app.Now)
	res.Provider = proj.provider.Name()
	res.Target = o.target
	stop := func(check string, err error) error {
		res.AddCheck(check, run.StatusError, err.Error())
		return app.finish(res, c, g)
	}
	// The earlier result is read before this run writes its own over it.
	previous, hasPrevious, previousErr := readResult(filepath.Join(g.results(), "deploy.json"))

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
	if err := providers.CheckDigest(g.dir, cs); err != nil {
		return stop("plan", err)
	}
	res.Facts["deploy.plan_sha256"] = cs.SHA256

	pm, err := app.verifyPermit(o, g, proj.cfg, permit.Expect{
		TargetName:   t.Name,
		TargetID:     proj.provider.CanonicalID(t.ID),
		Commit:       commit,
		PlanSHA256:   cs.SHA256,
		ConfigSHA256: cfgSHA,
		// The change set hash covers which plan files are destructive, and the command hash
		// is checked only when a command after -- replaces exec.deploy.
		ChangeSetSHA256:     permit.ChangeSetDigest(cs),
		DeployCommandSHA256: permit.DeployCommandDigest(o.command),
		Now:                 app.Now(),
		MaxTTL:              proj.cfg.Policy.TTL(),
		CanonicalID:         proj.provider.CanonicalID,
	})
	if pm.ID != "" {
		res.Facts["deploy.permit_id"] = pm.ID
	}
	if err != nil {
		return stop("permit", err)
	}
	res.AddCheck("permit", run.StatusPass, fmt.Sprintf("PERMIT %s verified for %s at %s", pm.ID, t.Name, short(commit)))

	// The PERMIT names a commit, and the source deployed is the working tree, so the tree
	// must be that commit with no tracked file changed.
	if err := app.checkTree(g, pm.Commit); err != nil {
		return stop("tree", err)
	}
	res.AddCheck("tree", run.StatusPass, fmt.Sprintf("HEAD is %s and no tracked file is modified", short(pm.Commit)))
	if o.resumeID != "" {
		if previousErr == nil {
			previousErr = checkResume(previous, hasPrevious, o.resumeID, pm.ID, t.Name)
		}
		if previousErr != nil {
			return stop("resume", previousErr)
		}
	}

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
	res.Facts["deploy.test_level"] = opts.TestLevel
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

	permitPath := o.permitPath
	if permitPath == "" {
		permitPath = filepath.Join(g.dir, ".glassops", "permit.json")
	}
	app.bindWorkspace(proj.provider, g, planPath, permitPath)
	sess, err := app.deployConnect(ctx, res, proj.provider, t)
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
	if out.Quick && pm.Validation != nil && pm.Validation.TestLevel != "" {
		// A quick deploy runs no tests of its own. Its tests ran in the validation.
		res.Facts["deploy.test_level"] = pm.Validation.TestLevel
	}

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

// maxTreePaths is how many modified paths a tree refusal names.
const maxTreePaths = 10

// checkTree refuses unless git HEAD in the project dir is commit and no tracked file is
// modified, staged or not. Untracked files, such as .glassops itself, are not looked at. When
// git is missing or the dir is not a repository it refuses, because the tree cannot be shown
// to be the commit the PERMIT names. The error names the commit and the paths, never file
// contents.
func (app *App) checkTree(g *globals, commit string) error {
	out, err := app.Run(g.dir, "git", "rev-parse", "HEAD")
	head := strings.TrimSpace(string(out))
	if err != nil || head == "" {
		return errors.New("cannot read git HEAD, so the source cannot be shown to be the PERMIT's commit. deploy runs in a git checkout of that commit")
	}
	if head != commit {
		return fmt.Errorf("git HEAD is %s, and the PERMIT is for commit %s", short(head), short(commit))
	}
	out, err = app.Run(g.dir, "git", "status", "--porcelain", "--untracked-files=no")
	if err != nil {
		return errors.New("git status failed, so the tracked files cannot be shown to be unchanged")
	}
	var paths []string
	for _, line := range strings.Split(strings.TrimRight(string(out), "\r\n"), "\n") {
		line = strings.TrimRight(line, "\r")
		if len(line) > 3 {
			paths = append(paths, line[3:])
		}
	}
	if len(paths) == 0 {
		return nil
	}
	more := ""
	if len(paths) > maxTreePaths {
		more = fmt.Sprintf(" and %d more", len(paths)-maxTreePaths)
		paths = paths[:maxTreePaths]
	}
	return fmt.Errorf("tracked files differ from commit %s: %s%s", short(commit), strings.Join(paths, ", "), more)
}

// checkResume allows --resume only for the job the last deploy result names, when that run
// timed out under the same PERMIT on the same target.
func checkResume(prev run.Result, ok bool, jobID, permitID, target string) error {
	if !ok {
		return fmt.Errorf("--resume %s: there is no earlier deploy result, so no timed out job to resume", jobID)
	}
	id, _ := prev.Facts["deploy.id"].(string)
	pid, _ := prev.Facts["deploy.permit_id"].(string)
	timedOut, _ := prev.Facts["deploy.timed_out"].(bool)
	switch {
	case prev.Target != target:
		return fmt.Errorf("--resume %s: the last deploy result is for target %q, not %q", jobID, prev.Target, target)
	case id != jobID:
		return fmt.Errorf("--resume %s: the last deploy result names job %q, not this one", jobID, id)
	case pid != permitID:
		return fmt.Errorf("--resume %s: the last deploy ran under another PERMIT", jobID)
	case !timedOut:
		return fmt.Errorf("--resume %s: the last deploy did not time out, so there is nothing to resume", jobID)
	}
	return nil
}

// deployOverrider is implemented by providers whose deploy command can be replaced from the
// command line, which is provider exec.
type deployOverrider interface {
	OverrideDeploy(command string)
}

// deployConnect logs in and runs the guard. A provider exec target whose tier does not need
// exec.describe has no way to report its Id, so the guard is recorded as skipped with the
// reason, as gate --no-connect does. Every other describe or guard failure stops the deploy.
func (app *App) deployConnect(ctx context.Context, res *run.Result, p providers.Provider, t providers.Target) (providers.Session, error) {
	sess, _, err := app.connect(ctx, p, t)
	if err == nil {
		return sess, nil
	}
	if !errors.Is(err, execprovider.ErrNoDescribe) {
		return nil, err
	}
	sess, aerr := p.Authenticate(ctx, t)
	if aerr != nil {
		return nil, fmt.Errorf("%w: target %s: %w", ErrLogin, t.Name, aerr)
	}
	res.AddCheck("guard", run.StatusSkipped, fmt.Sprintf("target %s, tier %s: %s", t.Name, t.Tier, execprovider.ErrNoDescribe))
	res.Facts["deploy.guard"] = "skipped"
	return sess, nil
}

// shellJoin quotes each argument for sh, so the command runs exactly the words given.
func shellJoin(args []string) string {
	quoted := make([]string, len(args))
	for i, a := range args {
		quoted[i] = shellQuote(a)
	}
	return strings.Join(quoted, " ")
}

// shellQuote leaves a word of plain characters as it is and puts anything else in single
// quotes, with each single quote written as '"'"'.
func shellQuote(s string) string {
	if s != "" && strings.Trim(s, "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789_@%+=:,./-") == "" {
		return s
	}
	return "'" + strings.ReplaceAll(s, "'", `'"'"'`) + "'"
}

// verifyPermit reads the PERMIT and checks it against want with the verifier
// policy.permit.signing selects, for hmac the key from the environment. The returned PERMIT
// is filled in whenever it could be read, so its Id can be recorded even when it does not
// verify. The key itself never reaches an error.
func (app *App) verifyPermit(o *deployOptions, g *globals, cfg *config.Config, want permit.Expect) (permit.Permit, error) {
	verifier, err := permit.VerifierFor(cfg, permit.Keys{HMACKey: app.permitKey, NewHMAC: permitFuncs.signer})
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
	if err := permitFuncs.verify(pm, verifier, want); err != nil {
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
	return permit.ParseKey(v)
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
