package cli

import (
	"github.com/spf13/cobra"
)

type analyzeOptions struct {
	ruleSelector string
	sarifPath    string
}

func newAnalyzeCommand(app *App, g *globals) *cobra.Command {
	o := &analyzeOptions{}
	cmd := &cobra.Command{
		Use:   "analyze",
		Short: "Run static analysis over the source",
		Long: "analyze runs the provider's static analyzer over the project and records the\n" +
			"violations by severity, so gate can apply analyzer_max_severity.",
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return notImplemented("analyze")
		},
	}
	f := cmd.Flags()
	f.StringVar(&o.ruleSelector, "rule-selector", "", "rules to run (default analyze.rule_selector, else Recommended)")
	f.StringVar(&o.sarifPath, "sarif", "", "SARIF file to write (default <results-dir>/analyze.sarif)")
	return cmd
}
