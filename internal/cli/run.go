package cli

import (
	"github.com/spf13/cobra"
)

type runOptions struct {
	dryRun bool
}

func newRunCommand(app *App, g *globals) *cobra.Command {
	o := &runOptions{}
	cmd := &cobra.Command{
		Use:   "run <pipeline>",
		Short: "Run a pipeline from glassops.yaml",
		Long: "run runs the steps of one pipeline in order in this process. A step's when decides\n" +
			"whether it runs after an earlier step failed. An exec step runs its run command\n" +
			"with sh -c. With --dry-run it prints the steps without running them.",
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, _ []string) error {
			return notImplemented("run")
		},
	}
	cmd.Flags().BoolVar(&o.dryRun, "dry-run", false, "print the steps without running them")
	return cmd
}
