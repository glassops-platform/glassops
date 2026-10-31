package cli

import (
	"github.com/spf13/cobra"
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
			return notImplemented("verify")
		},
	}
	f := cmd.Flags()
	f.StringVar(&o.target, "target", "", "target name from glassops.yaml (required)")
	f.StringVar(&o.planPath, "plan", "", "change set file (default <dir>/.glassops/plan.json)")
	f.StringVar(&o.deployResult, "deploy-result", "", "deploy result to verify (default <results-dir>/deploy.json)")
	_ = cmd.MarkFlagRequired("target")
	return cmd
}
