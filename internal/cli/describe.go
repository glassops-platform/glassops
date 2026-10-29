package cli

import (
	"context"

	"github.com/spf13/cobra"
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

// describe is implemented by Phase 2 workstream 2.
func (app *App) describe(ctx context.Context, g *globals, o *describeOptions) error {
	return notImplemented("describe")
}
