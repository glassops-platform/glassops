package cli

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/glassops-platform/glassops/internal/ci"
	"github.com/glassops-platform/glassops/internal/core/config"
	"github.com/glassops-platform/glassops/internal/core/exitcode"
	"github.com/glassops-platform/glassops/internal/core/run"
	"github.com/glassops-platform/glassops/internal/providers"
)

type validateOptions struct {
	target         string
	fallbackTarget string
	planPath       string
	checkOnly      bool
}

func newValidateCommand(app *App, g *globals) *cobra.Command {
	o := &validateOptions{}
	cmd := &cobra.Command{
		Use:   "validate",
		Short: "Check-only validate the plan against a target",
		Long: "validate logs in to the target, runs the org Id and tier guard, and runs a\n" +
			"check-only validate of the change set in .glassops/plan.json. It never deploys.\n" +
			"If the target reports anything other than a check-only run, validate exits 2.\n" +
			"With --fallback-target, a target that is not configured or cannot be reached is\n" +
			"replaced by the fallback for the same check-only run.\n" +
			"On a scratch target, a plan that deletes components first gets them deployed from\n" +
			"the plan's base ref, as salesforce.ephemeral.seed says (minimal, full or off).",
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return app.validate(cmd.Context(), g, o)
		},
	}
	f := cmd.Flags()
	f.StringVar(&o.target, "target", "", "target name from glassops.yaml (required)")
	f.StringVar(&o.fallbackTarget, "fallback-target", "", "target to use when --target is not configured or not reachable")
	f.StringVar(&o.planPath, "plan", "", "change set file (default <dir>/.glassops/plan.json)")
	f.BoolVar(&o.checkOnly, "check-only", true, "always true. validate never deploys, and false is refused")
	_ = cmd.MarkFlagRequired("target")
	return cmd
}

// validate runs a check-only validate of the change set against the target, or against the
// fallback when the target is not configured or cannot be logged in to.
func (app *App) validate(ctx context.Context, g *globals, o *validateOptions) error {
	if !o.checkOnly {
		return &exitcode.ExitError{Code: exitcode.Error, Err: errors.New("validate never deploys, so --check-only=false is refused")}
	}
	c, err := app.detectCI(g)
	if err != nil {
		return &exitcode.ExitError{Code: exitcode.Error, Err: err}
	}
	proj, err := app.loadProject(g)
	if err != nil {
		return err
	}
	res := run.New("validate", app.Now)
	res.Provider = proj.provider.Name()
	res.Facts["validate.fallback_used"] = false

	planPath := o.planPath
	if planPath == "" {
		planPath = filepath.Join(g.dir, ".glassops", "plan.json")
	}
	cs, err := providers.ReadChangeSet(planPath)
	if err != nil {
		res.AddCheck("plan", run.StatusError, err.Error())
		return app.finish(res, c, g)
	}
	if cs.Empty {
		res.Status = run.StatusEmpty
		res.AddCheck("plan", run.StatusSkipped, "the change set is empty, nothing to validate")
		return app.finish(res, c, g)
	}

	sess, used, info, fallbackUsed, err := app.validateConnect(ctx, proj, o)
	if err != nil {
		res.AddCheck("target", run.StatusError, err.Error())
		return app.finish(res, c, g)
	}
	defer sess.Close()
	res.Target = used.Name
	res.Facts["target.source"] = targetSource(used)
	res.Facts["validate.fallback_used"] = fallbackUsed

	if info.Kind == providers.KindScratch && len(cs.Destructive) > 0 {
		if !app.seed(ctx, res, c, proj, sess, cs, g) {
			return app.finish(res, c, g)
		}
	}

	level := testLevelFor(proj.cfg.Salesforce.TestLevel, used.Tier)
	out, err := proj.provider.Validate(ctx, sess, cs, providers.ValidateOptions{
		Dir:       g.dir,
		TestLevel: level,
	})
	if err != nil {
		res.AddCheck("validate", run.StatusError, err.Error())
		return app.finish(res, c, g)
	}
	for k, v := range out.Facts {
		res.Facts[k] = v
	}
	// gate binds this validation into a PERMIT only for the same plan, and quick deploy
	// reuses it at the same test level, so both are recorded after the provider's own facts.
	res.Facts["validate.plan_sha256"] = cs.SHA256
	res.Facts["validate.test_level"] = level
	if out.Passed {
		res.AddCheck("validate", run.StatusPass, "check-only validate passed")
	}
	for i, p := range out.Problems {
		res.AddCheck(fmt.Sprintf("problem.%d", i+1), run.StatusFail, p)
	}
	if !out.Passed {
		res.Degrade(run.StatusFail)
	}
	return app.finish(res, c, g)
}

// validateConnect resolves the target and logs in. A target that is not configured, or a
// login that fails, falls back to --fallback-target when one is set. A guard mismatch or any
// other failure never does.
func (app *App) validateConnect(ctx context.Context, proj *project, o *validateOptions) (providers.Session, providers.Target, providers.TargetInfo, bool, error) {
	var none providers.TargetInfo
	fallback := func(reason error) (providers.Session, providers.Target, providers.TargetInfo, bool, error) {
		if o.fallbackTarget == "" {
			return nil, providers.Target{}, none, false, reason
		}
		ft, err := app.target(proj.cfg, o.fallbackTarget)
		if err != nil {
			return nil, providers.Target{}, none, false, fmt.Errorf("fallback target: %w", err)
		}
		s, info, err := app.connect(ctx, proj.provider, ft)
		if err != nil {
			return nil, ft, none, true, fmt.Errorf("fallback target: %w", err)
		}
		return s, ft, info, true, nil
	}

	t, err := app.target(proj.cfg, o.target)
	if errors.Is(err, ErrTargetNotConfigured) {
		return fallback(err)
	}
	if err != nil {
		return nil, t, none, false, err
	}
	s, info, err := app.connect(ctx, proj.provider, t)
	if errors.Is(err, ErrLogin) && ctx.Err() == nil {
		return fallback(err)
	}
	if err != nil {
		return nil, t, none, false, err
	}
	return s, t, info, false, nil
}

// seed puts the base state a destructive change needs on a scratch target before it is
// validated there, as salesforce.ephemeral.seed says. It records validate.seed.mode,
// validate.seed.fallback and validate.seed.members. It returns false when the validate must
// not run, because the seed failed or could not be attempted.
func (app *App) seed(ctx context.Context, res *run.Result, c ci.CI, proj *project, sess providers.Session, cs providers.ChangeSet, g *globals) bool {
	mode := proj.cfg.Salesforce.Ephemeral.SeedMode()
	res.Facts["validate.seed.mode"] = mode
	res.Facts["validate.seed.fallback"] = false
	res.Facts["validate.seed.members"] = []string{}
	switch {
	case mode == config.SeedOff:
		res.AddCheck("seed", run.StatusSkipped, "seed is off, so the destructive change is validated on the target as it is")
		return true
	case cs.From == "":
		res.Facts["validate.seed.mode"] = "none"
		res.AddCheck("seed", run.StatusSkipped, "the plan has no base ref, so there is nothing to seed")
		return true
	}
	seeder, ok := proj.provider.(providers.Seeder)
	if !ok {
		res.AddCheck("seed", run.StatusError, fmt.Sprintf("provider %s cannot seed a target, set seed to off to validate without it", proj.provider.Name()))
		return false
	}
	out, err := seeder.Seed(ctx, sess, providers.SeedRequest{Dir: g.dir, BaseRef: cs.From, Plan: cs, Mode: mode})
	if err != nil {
		res.AddCheck("seed", run.StatusError, err.Error())
		return false
	}
	for k, v := range out.Facts {
		res.Facts[k] = v
	}
	members := out.Members
	if members == nil {
		members = []string{}
	}
	res.Facts["validate.seed.mode"] = out.Mode
	res.Facts["validate.seed.fallback"] = out.Fallback
	res.Facts["validate.seed.members"] = members
	detail := fmt.Sprintf("seeded %d components from the base ref, mode %s", len(members), out.Mode)
	if out.Fallback {
		detail += ", after the minimal seed failed on a missing dependency"
		c.Annotate(ci.Warning, "validate: the minimal seed failed on a missing dependency, so the whole base ref source was deployed to the target first")
	}
	res.AddCheck("seed", run.StatusPass, detail)
	return true
}

// testLevelFor picks the test level for a tier: the tier's own entry, else "default", else
// empty, which leaves sf on its own default.
func testLevelFor(levels map[string]string, tier string) string {
	if l := levels[tier]; l != "" {
		return l
	}
	return levels["default"]
}
