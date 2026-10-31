package cli

import (
	"github.com/spf13/cobra"
)

type gateOptions struct {
	target     string
	planPath   string
	permitPath string
	noConnect  bool
}

func newGateCommand(app *App, g *globals) *cobra.Command {
	o := &gateOptions{}
	cmd := &cobra.Command{
		Use:   "gate",
		Short: "Evaluate policy for a deploy and write a signed PERMIT",
		Long: "gate reads the change set, earlier results and the CI context, builds the facts\n" +
			"for one target and runs the policy rules over them. It always writes a signed\n" +
			"PERMIT to .glassops/permit.json, which records allow or deny and every rule result.\n" +
			"It logs in to the target to run the org Id and tier guard unless --no-connect is\n" +
			"set. It exits 1 when policy denies the deploy, and 2 when GLASSOPS_PERMIT_KEY is not\n" +
			"set.",
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return notImplemented("gate")
		},
	}
	f := cmd.Flags()
	f.StringVar(&o.target, "target", "", "target name from glassops.yaml (required)")
	f.StringVar(&o.planPath, "plan", "", "change set file (default <dir>/.glassops/plan.json)")
	f.StringVar(&o.permitPath, "permit", "", "where the PERMIT is written (default <dir>/.glassops/permit.json)")
	f.BoolVar(&o.noConnect, "no-connect", false, "do not log in to the target, and use its configured Id")
	_ = cmd.MarkFlagRequired("target")
	return cmd
}
