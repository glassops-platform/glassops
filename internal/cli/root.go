package cli

import (
	"errors"
	"fmt"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/glassops-platform/glassops/internal/ci"
	"github.com/glassops-platform/glassops/internal/core/exitcode"
)

// globals are the flags every command shares.
type globals struct {
	dir        string
	configPath string
	resultsDir string
	ciName     string
}

func (g *globals) results() string {
	if g.resultsDir != "" {
		return g.resultsDir
	}
	return filepath.Join(g.dir, ".glassops", "results")
}

// NewRootCommand builds the command tree.
func NewRootCommand(app *App) *cobra.Command {
	g := &globals{}
	root := &cobra.Command{
		Use:   "glassops",
		Short: "Governed delivery pipelines that run the same in any CI or on a laptop",
		Long: "glassops runs each delivery step as its own command. Every command writes a\n" +
			"result to .glassops/results/ and exits 0 on pass, 1 on a policy or check\n" +
			"failure, and 2 on a tool or infrastructure error.",
		Version:       app.Version,
		SilenceUsage:  true,
		SilenceErrors: true,
	}
	root.SetOut(app.Stdout)
	root.SetErr(app.Stderr)

	pf := root.PersistentFlags()
	pf.StringVar(&g.dir, "dir", ".", "project root")
	pf.StringVar(&g.configPath, "config", "", "config file (default <dir>/glassops.yaml)")
	pf.StringVar(&g.resultsDir, "results-dir", "", "where results are written (default <dir>/.glassops/results)")
	pf.StringVar(&g.ciName, "ci", "auto", "CI system: auto, github or local")

	root.AddCommand(newDoctorCommand(app, g), newInitCommand(app, g))
	return root
}

// detectCI resolves the CI layer for a command.
func (app *App) detectCI(g *globals) (ci.CI, error) {
	return ci.Detect(g.ciName, app.Getenv, app.Stdout, g.dir)
}

// Execute runs the command tree with args and returns the process exit code.
func Execute(app *App, args []string) int {
	root := NewRootCommand(app)
	root.SetArgs(args)
	err := root.Execute()
	if err == nil {
		return exitcode.OK
	}
	var ee *exitcode.ExitError
	if errors.As(err, &ee) {
		if ee.Err != nil {
			fmt.Fprintln(app.Stderr, "glassops:", ee.Err)
		}
		return ee.Code
	}
	// Flag and argument errors from cobra are usage problems, not check failures.
	fmt.Fprintln(app.Stderr, "glassops:", err)
	return exitcode.Error
}
