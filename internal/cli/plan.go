package cli

import (
	"context"

	"github.com/spf13/cobra"
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

// plan is implemented by Phase 2 workstream 3.
func (app *App) plan(ctx context.Context, g *globals, o *planOptions) error {
	return notImplemented("plan")
}
