package cli

import (
	"time"

	"github.com/spf13/cobra"
)

type testOptions struct {
	target    string
	level     string
	tests     []string
	wait      time.Duration
	junitPath string
}

func newTestCommand(app *App, g *globals) *cobra.Command {
	o := &testOptions{}
	cmd := &cobra.Command{
		Use:   "test",
		Short: "Run tests on a target and check coverage",
		Long: "test logs in to the target, runs the org Id and tier guard, and runs the tests at\n" +
			"the configured test level with coverage. It exits 1 when a test fails, the run does\n" +
			"not finish, or coverage is below test.coverage.",
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return notImplemented("test")
		},
	}
	f := cmd.Flags()
	f.StringVar(&o.target, "target", "", "target name from glassops.yaml (required)")
	f.StringVar(&o.level, "level", "", "test level (default test.level for the tier, else test.level.default)")
	f.StringSliceVar(&o.tests, "tests", nil, "tests to run, comma separated, for a level that names them")
	f.DurationVar(&o.wait, "wait", 0, "how long to wait for the run, such as 30m (default the provider's)")
	f.StringVar(&o.junitPath, "junit", "", "JUnit XML file to write (default <results-dir>/test.junit.xml)")
	_ = cmd.MarkFlagRequired("target")
	return cmd
}
