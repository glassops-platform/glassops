package cli

import (
	"github.com/spf13/cobra"
)

type permitVerifyOptions struct {
	target     string
	permitPath string
	planPath   string
}

func newPermitCommand(app *App, g *globals) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "permit",
		Short: "Work with PERMITs",
		Args:  cobra.NoArgs,
	}
	cmd.AddCommand(newPermitVerifyCommand(app, g))
	return cmd
}

func newPermitVerifyCommand(app *App, g *globals) *cobra.Command {
	o := &permitVerifyOptions{}
	cmd := &cobra.Command{
		Use:   "verify",
		Short: "Check that a PERMIT allows this deploy",
		Long: "permit verify checks the PERMIT's signature and that it allows a deploy of this\n" +
			"plan, at this commit, with this config, to this target, before it expires. It\n" +
			"exits 0 when the PERMIT holds and 1 when it does not, so any deploy tool can honour\n" +
			"a PERMIT. It needs GLASSOPS_PERMIT_KEY and never logs in to the target.",
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return notImplemented("permit verify")
		},
	}
	f := cmd.Flags()
	f.StringVar(&o.target, "target", "", "target name from glassops.yaml (required)")
	f.StringVar(&o.permitPath, "permit", "", "PERMIT file (default <dir>/.glassops/permit.json)")
	f.StringVar(&o.planPath, "plan", "", "change set file (default <dir>/.glassops/plan.json)")
	_ = cmd.MarkFlagRequired("target")
	return cmd
}
