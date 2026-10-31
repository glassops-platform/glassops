package cli

import (
	"time"

	"github.com/spf13/cobra"
)

type deployOptions struct {
	target     string
	planPath   string
	permitPath string
	wait       time.Duration
	resumeID   string
}

func newDeployCommand(app *App, g *globals) *cobra.Command {
	o := &deployOptions{}
	cmd := &cobra.Command{
		Use:   "deploy",
		Short: "Deploy the plan to a target under a verified PERMIT",
		Long: "deploy verifies the PERMIT before it logs in, then runs the org Id and tier guard\n" +
			"and deploys the change set in .glassops/plan.json. When the PERMIT holds a\n" +
			"validated deploy Id that the target still accepts, deploy quick deploys it.\n" +
			"Success is read from the target's result body. When --wait runs out, the result\n" +
			"keeps the job Id, and --resume picks the same deploy up again.",
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return notImplemented("deploy")
		},
	}
	f := cmd.Flags()
	f.StringVar(&o.target, "target", "", "target name from glassops.yaml (required)")
	f.StringVar(&o.planPath, "plan", "", "change set file (default <dir>/.glassops/plan.json)")
	f.StringVar(&o.permitPath, "permit", "", "PERMIT file (default <dir>/.glassops/permit.json)")
	f.DurationVar(&o.wait, "wait", 0, "how long to wait for the deploy, such as 30m (default the provider's)")
	f.StringVar(&o.resumeID, "resume", "", "job Id of an earlier deploy to resume")
	_ = cmd.MarkFlagRequired("target")
	return cmd
}
