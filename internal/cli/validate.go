package cli

import (
	"context"

	"github.com/spf13/cobra"
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
			"replaced by the fallback for the same check-only run.",
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

// validate is implemented by Phase 2 workstream 4.
func (app *App) validate(ctx context.Context, g *globals, o *validateOptions) error {
	return notImplemented("validate")
}
