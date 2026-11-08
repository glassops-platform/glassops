package cli

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/glassops-platform/glassops/internal/ci"
	"github.com/glassops-platform/glassops/internal/core/exitcode"
	"github.com/glassops-platform/glassops/internal/core/run"
	"github.com/glassops-platform/glassops/internal/providers"
)

// Variables the verify.smoke command gets besides ResultsDirEnv.
const (
	smokeTargetEnv   = "GLASSOPS_TARGET"
	smokeTargetIDEnv = "GLASSOPS_TARGET_ID"
)

type verifyOptions struct {
	target       string
	planPath     string
	deployResult string
}

func newVerifyCommand(app *App, g *globals) *cobra.Command {
	o := &verifyOptions{}
	cmd := &cobra.Command{
		Use:   "verify",
		Short: "Check after a deploy that the target holds what the plan says",
		Long: "verify reads the deploy result and the change set, checks the plan files still\n" +
			"match the plan digest, logs in to the target and runs the org Id and tier guard.\n" +
			"Then the provider checks the deploy: every plan member deployed, the deploy not\n" +
			"check-only, tests run when the test level needs them, active Flows running as\n" +
			"their latest version, and other deploys to the target in the same window. Last it\n" +
			"runs verify.smoke when set. It writes verify.json, exits 1 when a check fails and 2\n" +
			"on a tool error. It never deploys.",
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return app.verify(cmd.Context(), g, o)
		},
	}
	f := cmd.Flags()
	f.StringVar(&o.target, "target", "", "target name from glassops.yaml (required)")
	f.StringVar(&o.planPath, "plan", "", "change set file (default <dir>/.glassops/plan.json)")
	f.StringVar(&o.deployResult, "deploy-result", "", "deploy result to verify (default <results-dir>/deploy.json)")
	_ = cmd.MarkFlagRequired("target")
	return cmd
}

// verify checks a finished deploy. Everything that can be checked without the target, the
// deploy result, the plan and its digest, is checked before the login.
func (app *App) verify(ctx context.Context, g *globals, o *verifyOptions) error {
	c, err := app.detectCI(g)
	if err != nil {
		return &exitcode.ExitError{Code: exitcode.Error, Err: err}
	}
	proj, err := app.loadProject(g)
	if err != nil {
		return err
	}
	res := run.New("verify", app.Now)
	res.Provider = proj.provider.Name()
	res.Target = o.target
	stop := func(check string, err error) error {
		res.AddCheck(check, run.StatusError, err.Error())
		return app.finish(res, c, g)
	}

	verifier, ok := proj.provider.(providers.Verifier)
	if !ok {
		return stop("verify", fmt.Errorf("provider %s cannot verify a deploy", proj.provider.Name()))
	}
	t, err := app.target(proj.cfg, o.target)
	if err != nil {
		return stop("target", err)
	}
	res.Facts["target.source"] = targetSource(t)

	resultPath := o.deployResult
	if resultPath == "" {
		resultPath = filepath.Join(g.results(), "deploy.json")
	}
	dep, err := readDeployResult(resultPath)
	if err != nil {
		return stop("deploy_result", err)
	}
	if dep.Status == run.StatusEmpty {
		res.Status = run.StatusEmpty
		res.AddCheck("verify", run.StatusSkipped, "the deploy deployed nothing, so there is nothing to verify")
		return app.finish(res, c, g)
	}
	rec, err := deployRecord(dep, o.target, proj.provider.Name())
	if err != nil {
		return stop("deploy_result", err)
	}
	res.Facts["verify.deploy_id"] = rec.ID
	res.Facts["verify.deploy_quick"] = rec.Quick
	if rec.Commit != "" {
		res.Facts["verify.commit"] = rec.Commit
	}
	if id, _ := dep.Facts["deploy.permit_id"].(string); id != "" {
		res.Facts["verify.permit_id"] = id
	}

	planPath := o.planPath
	if planPath == "" {
		planPath = filepath.Join(g.dir, ".glassops", "plan.json")
	}
	cs, err := providers.ReadChangeSet(planPath)
	if err != nil {
		return stop("plan", err)
	}
	if err := providers.CheckDigest(g.dir, cs); err != nil {
		return stop("plan", err)
	}
	// deploy records the plan it sent once it writes deploy.plan_sha256. Until then the
	// digest check above is what ties the plan files to plan.json.
	bound := false
	if sum, ok := dep.Facts["deploy.plan_sha256"].(string); ok {
		if sum != cs.SHA256 {
			return stop("plan", fmt.Errorf("the deploy sent plan %s, and %s holds plan %s", short(sum), planPath, short(cs.SHA256)))
		}
		bound = true
	}
	res.Facts["verify.plan_sha256"] = cs.SHA256
	res.Facts["verify.plan_bound_to_deploy"] = bound
	if level, ok := dep.Facts["deploy.test_level"].(string); ok {
		rec.TestLevel = level
	} else {
		rec.TestLevel = testLevelFor(proj.cfg.Salesforce.TestLevel, t.Tier)
	}

	sess, _, err := app.connect(ctx, proj.provider, t)
	if err != nil {
		return stop("target", err)
	}
	defer sess.Close()

	vc := proj.cfg.Verify
	out, err := verifier.Verify(ctx, sess, providers.VerifyRequest{
		Dir:            g.dir,
		Plan:           cs,
		Deploy:         rec,
		WindowStart:    rec.Started,
		WindowEnd:      app.Now(),
		ForeignDeploys: vc.ForeignDeploysMode(),
		Flows:          vc.FlowsMode(),
	})
	if err != nil {
		return stop("verify", err)
	}
	for k, v := range out.Facts {
		res.Facts[k] = v
	}
	var warnings []string
	for _, ch := range out.Checks {
		status := run.Status(ch.Status)
		detail := ch.Detail
		switch ch.Status {
		case providers.VerifyPass, providers.VerifyFail, providers.VerifySkipped:
		case providers.VerifyWarn:
			// A result check has no warning status. A warning passes the step, says so in
			// its detail and is annotated, and verify.warnings lists it.
			status = run.StatusPass
			detail = "warning: " + detail
			warnings = append(warnings, ch.Name)
			c.Annotate(ci.Warning, fmt.Sprintf("verify %s: %s", ch.Name, ch.Detail))
		default:
			status = run.StatusError
			detail = fmt.Sprintf("the provider returned status %q: %s", ch.Status, ch.Detail)
		}
		res.AddCheck(ch.Name, status, detail)
	}
	res.Facts["verify.warnings"] = nonNilStrings(warnings)
	if !out.Passed {
		// A provider that says failed without a failing check still fails the step.
		res.Degrade(run.StatusFail)
	}

	app.verifySmoke(ctx, res, c, g, vc.Smoke, t)
	res.Facts["verify.passed"] = res.Status == run.StatusPass
	return app.finish(res, c, g)
}

// verifySmoke runs verify.smoke with sh -c in the project dir, like an exec step, with the
// results dir, the target name and the target Id in its environment. Exit 0 passes, any other
// exit fails, and a command that cannot start is an error.
func (app *App) verifySmoke(ctx context.Context, res *run.Result, c ci.CI, g *globals, smoke string, t providers.Target) {
	if strings.TrimSpace(smoke) == "" {
		res.AddCheck("verify.smoke", run.StatusSkipped, "verify.smoke is not set")
		return
	}
	results, err := filepath.Abs(g.results())
	if err == nil {
		err = os.MkdirAll(results, 0o755)
	}
	if err != nil {
		res.AddCheck("verify.smoke", run.StatusError, fmt.Sprintf("results dir: %v", err))
		return
	}
	env := append(os.Environ(), ResultsDirEnv+"="+results, smokeTargetEnv+"="+t.Name, smokeTargetIDEnv+"="+t.ID)
	code, err := shellRunner(ctx, g.dir, env, smoke, app.Stdout, app.Stderr)
	if err != nil {
		res.AddCheck("verify.smoke", run.StatusError, fmt.Sprintf("%s: %v", smoke, err))
		return
	}
	res.Facts["verify.smoke_exit_code"] = code
	res.AddCheck("verify.smoke", execStatus(code), fmt.Sprintf("%s: exit %d", smoke, code))
	if code != 0 {
		c.Annotate(ci.Errors, fmt.Sprintf("verify: the smoke command exited %d", code))
	}
}

// readDeployResult reads the result deploy wrote.
func readDeployResult(path string) (run.Result, error) {
	var r run.Result
	data, err := os.ReadFile(path)
	if err != nil {
		return r, fmt.Errorf("read the deploy result: %w", err)
	}
	if err := json.Unmarshal(data, &r); err != nil {
		return r, fmt.Errorf("parse the deploy result %s: %w", path, err)
	}
	if r.Schema != run.Schema || r.Step != "deploy" {
		return r, fmt.Errorf("%s is not a deploy result (schema %q, step %q)", path, r.Schema, r.Step)
	}
	return r, nil
}

// deployRecord turns a deploy result into what the provider verifies. A deploy that did not
// pass, timed out, ran for another target or provider, or has no Id is refused, because
// there is no finished deploy to check.
func deployRecord(r run.Result, target, provider string) (providers.DeployRecord, error) {
	if r.Target != target {
		return providers.DeployRecord{}, fmt.Errorf("the deploy result is for target %q, not %s", r.Target, target)
	}
	if r.Provider != provider {
		return providers.DeployRecord{}, fmt.Errorf("the deploy result is from provider %q, not %s", r.Provider, provider)
	}
	if r.Status != run.StatusPass {
		return providers.DeployRecord{}, fmt.Errorf("the deploy result has status %s, so there is no finished deploy to verify", r.Status)
	}
	if timedOut, _ := r.Facts["deploy.timed_out"].(bool); timedOut {
		return providers.DeployRecord{}, errors.New("the deploy result says the deploy did not finish")
	}
	id, _ := r.Facts["deploy.id"].(string)
	if id == "" {
		return providers.DeployRecord{}, errors.New("the deploy result has no deploy.id")
	}
	if r.Started.IsZero() {
		return providers.DeployRecord{}, errors.New("the deploy result has no start time")
	}
	quick, _ := r.Facts["deploy.quick"].(bool)
	commit, _ := r.Facts["deploy.commit"].(string)
	return providers.DeployRecord{ID: id, Quick: quick, Started: r.Started, Ended: r.Ended, Commit: commit}, nil
}

func nonNilStrings(xs []string) []string {
	if xs == nil {
		return []string{}
	}
	return xs
}
