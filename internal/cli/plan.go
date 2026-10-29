package cli

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/glassops-platform/glassops/internal/core/exitcode"
	"github.com/glassops-platform/glassops/internal/core/gitref"
	"github.com/glassops-platform/glassops/internal/core/run"
	"github.com/glassops-platform/glassops/internal/providers"
)

type planOptions struct {
	from   string
	to     string
	target string
}

func newPlanCommand(app *App, g *globals) *cobra.Command {
	o := &planOptions{}
	cmd := &cobra.Command{
		Use:   "plan",
		Short: "Work out what changed between two git refs",
		Long: "plan works out the change set between two git refs and writes the provider's\n" +
			"manifests to .glassops/plan/ and the change set to .glassops/plan.json.\n" +
			"With --from auto, a pull request uses the merge base with its base branch. With\n" +
			"--target, auto uses the git tag glassops/deployed/<target>, then the merge base with\n" +
			"the target's branch. A shallow clone is an error, never an empty plan.",
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return app.plan(cmd.Context(), g, o)
		},
	}
	f := cmd.Flags()
	f.StringVar(&o.from, "from", "auto", "base git ref, or auto")
	f.StringVar(&o.to, "to", "auto", "head git ref, or auto (HEAD)")
	f.StringVar(&o.target, "target", "", "target the plan is for, used by --from auto")
	return cmd
}

func (app *App) plan(ctx context.Context, g *globals, o *planOptions) error {
	proj, err := app.loadProject(g)
	if err != nil {
		return err
	}
	c, err := app.detectCI(g)
	if err != nil {
		return &exitcode.ExitError{Code: exitcode.Error, Err: err}
	}
	cfg := proj.cfg
	res := run.New("plan", app.Now)
	res.Provider = proj.provider.Name()

	var branch string
	if o.target != "" {
		t, ok := cfg.Targets[o.target]
		if !ok {
			return &exitcode.ExitError{Code: exitcode.Error, Err: fmt.Errorf("%w: %s", ErrTargetNotConfigured, o.target)}
		}
		branch = t.Branch
		res.Target = o.target
	}

	// delta.enabled false plans the whole source. Nothing then depends on history, so no ref
	// is resolved and a shallow clone does not matter.
	full := cfg.Salesforce.Delta.Enabled != nil && !*cfg.Salesforce.Delta.Enabled
	git := func(dir string, args ...string) (string, error) {
		out, err := app.Run(dir, "git", args...)
		return strings.TrimSpace(string(out)), err
	}

	req := providers.PlanRequest{
		Dir:    g.dir,
		Full:   full,
		OutDir: filepath.Join(g.dir, ".glassops", "plan"),
	}
	if full {
		// Best effort, only to record which commit the plan was made at.
		if sha, err := git(g.dir, "rev-parse", "HEAD"); err == nil {
			req.To = sha
		}
		res.AddCheck("refs", run.StatusSkipped, "full plan, history is not used")
	} else {
		refs, err := gitref.Resolve(git, gitref.Options{
			Dir:        g.dir,
			From:       o.from,
			To:         o.to,
			Target:     o.target,
			Branch:     branch,
			BaseBranch: c.Context().BaseRef,
		})
		if err != nil {
			res.AddCheck("refs", run.StatusError, err.Error())
			res.Errors = append(res.Errors, err.Error())
			return app.finish(res, c, g)
		}
		req.From, req.To = refs.From, refs.To
		res.AddCheck("refs", run.StatusPass, fmt.Sprintf("%s..%s (%s)", short(refs.From), short(refs.To), refs.FromSource))
		res.Facts["plan.from_source"] = refs.FromSource
	}

	cs, err := proj.provider.Plan(ctx, req)
	if err != nil {
		res.AddCheck("plan", run.StatusError, err.Error())
		res.Errors = append(res.Errors, err.Error())
		return app.finish(res, c, g)
	}

	planPath := filepath.Join(g.dir, ".glassops", "plan.json")
	if err := providers.WriteChangeSet(planPath, cs); err != nil {
		res.AddCheck("plan", run.StatusError, err.Error())
		res.Errors = append(res.Errors, err.Error())
		return app.finish(res, c, g)
	}
	res.Artifacts = append(res.Artifacts, planPath)
	for k, v := range cs.Facts {
		res.Facts[k] = v
	}
	res.Facts["plan.full"] = cs.Full
	res.Facts["plan.from"] = cs.From
	res.Facts["plan.to"] = cs.To
	res.Facts["plan.sha256"] = cs.SHA256
	res.Facts["plan.destructive"] = len(cs.Destructive)

	if cs.Empty {
		res.AddCheck("plan", run.StatusEmpty, "nothing to deploy")
		if res.Status == run.StatusPass {
			res.Status = run.StatusEmpty
		}
	} else {
		res.AddCheck("plan", run.StatusPass, fmt.Sprintf("manifest %s, %d destructive", cs.Manifest, len(cs.Destructive)))
	}
	return app.finish(res, c, g)
}

func short(sha string) string {
	if len(sha) > 8 {
		return sha[:8]
	}
	return sha
}
