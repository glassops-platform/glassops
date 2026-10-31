package cli

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/glassops-platform/glassops/internal/core/exitcode"
	"github.com/glassops-platform/glassops/internal/core/run"
)

type describeOptions struct {
	target string
}

func newDescribeCommand(app *App, g *globals) *cobra.Command {
	o := &describeOptions{}
	cmd := &cobra.Command{
		Use:   "describe",
		Short: "Log in to a target, check its Id and tier, and record what it reports",
		Long: "describe logs in to one target, runs the org Id and tier guard, and writes what the\n" +
			"target reports about itself to .glassops/results/describe.json. A guard mismatch\n" +
			"exits 2 before anything else runs.",
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return app.describe(cmd.Context(), g, o)
		},
	}
	cmd.Flags().StringVar(&o.target, "target", "", "target name from glassops.yaml (required)")
	_ = cmd.MarkFlagRequired("target")
	return cmd
}

// describe logs in to one target, runs the guard and records what the target reports. A
// login failure, a guard mismatch or an unreadable target is exit 2 with the error in the
// result.
func (app *App) describe(ctx context.Context, g *globals, o *describeOptions) error {
	c, err := app.detectCI(g)
	if err != nil {
		return &exitcode.ExitError{Code: exitcode.Error, Err: err}
	}
	proj, err := app.loadProject(g)
	if err != nil {
		return err
	}
	t, err := app.target(proj.cfg, o.target)
	if err != nil {
		return &exitcode.ExitError{Code: exitcode.Error, Err: err}
	}

	res := run.New("describe", app.Now)
	res.Provider = proj.provider.Name()
	res.Target = t.Name
	res.Facts["target.source"] = targetSource(t)

	s, info, err := app.connect(ctx, proj.provider, t)
	if err != nil {
		res.AddError(err)
		return app.finish(res, c, g)
	}
	defer s.Close()

	for k, v := range info.Facts {
		res.Facts[k] = v
	}
	res.Facts["target.id"] = info.ID
	res.Facts["target.kind"] = string(info.Kind)
	res.AddCheck("describe", run.StatusPass, string(info.Kind)+" "+info.ID)
	return app.finish(res, c, g)
}
