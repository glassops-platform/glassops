package cli

import (
	"github.com/spf13/cobra"
)

type reportOptions struct {
	sarifPath   string
	summaryPath string
}

func newReportCommand(app *App, g *globals) *cobra.Command {
	o := &reportOptions{}
	cmd := &cobra.Command{
		Use:   "report",
		Short: "Summarise every result for the CI and write SARIF",
		Long: "report reads every result in .glassops/results/, writes a markdown summary to\n" +
			"the CI's run summary, and writes SARIF for policy denials and validate failures.\n" +
			"With --summary the markdown also goes to a file.",
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return notImplemented("report")
		},
	}
	f := cmd.Flags()
	f.StringVar(&o.sarifPath, "sarif", "", "SARIF file to write (default <dir>/.glassops/report.sarif)")
	f.StringVar(&o.summaryPath, "summary", "", "also write the markdown summary to this file")
	return cmd
}
