package cli

import (
	"github.com/spf13/cobra"
)

type checkOptions struct {
	script string
}

func newCheckCommand(app *App, g *globals) *cobra.Command {
	o := &checkOptions{}
	cmd := &cobra.Command{
		Use:   "check",
		Short: "Run the check scripts from glassops.yaml",
		Long: "check runs each script under check.scripts, or only --script, and exits 1 when\n" +
			"any of them fails.",
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return notImplemented("check")
		},
	}
	cmd.Flags().StringVar(&o.script, "script", "", "run only the script with this name")
	return cmd
}
