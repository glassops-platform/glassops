package cli

import (
	"errors"
	"fmt"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"

	"github.com/glassops-platform/glassops/internal/core/exitcode"
	"github.com/glassops-platform/glassops/internal/core/run"
)

type checkOptions struct {
	script string
}

func newCheckCommand(app *App, g *globals) *cobra.Command {
	o := &checkOptions{}
	cmd := &cobra.Command{
		Use:   "check",
		Short: "Run the check scripts from glassops.yaml",
		Long: "check runs each script under check.scripts, or only --script, with sh -c in the\n" +
			"project directory and exits 1 when any of them fails. It records one check per\n" +
			"script, named script.<name>. It runs your own lint and test commands and does not\n" +
			"replace them. With no scripts configured it does nothing and exits 0.",
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return app.check(g, o)
		},
	}
	cmd.Flags().StringVar(&o.script, "script", "", "run only the script with this name")
	return cmd
}

// check runs the configured scripts in order and records one check for each. A script that
// exits non-zero is a fail. Every script runs even when an earlier one failed, so one run
// shows every problem.
func (app *App) check(g *globals, o *checkOptions) error {
	c, err := app.detectCI(g)
	if err != nil {
		return &exitcode.ExitError{Code: exitcode.Error, Err: err}
	}
	proj, err := app.loadProject(g)
	if err != nil {
		return err
	}
	res := run.New("check", app.Now)
	res.Provider = proj.provider.Name()

	scripts := proj.cfg.Check.Scripts
	if o.script != "" {
		scripts = nil
		for _, s := range proj.cfg.Check.Scripts {
			if s.Name == o.script {
				scripts = append(scripts, s)
			}
		}
		if len(scripts) == 0 {
			res.AddCheck("script."+o.script, run.StatusError, "no check.scripts entry has that name")
			return app.finish(res, c, g)
		}
	}
	res.Facts["check.scripts"] = len(scripts)
	if len(scripts) == 0 {
		res.Status = run.StatusEmpty
		res.AddCheck("scripts", run.StatusSkipped, "check.scripts is empty, nothing to run")
		return app.finish(res, c, g)
	}

	passed := 0
	for _, s := range scripts {
		name := "script." + s.Name
		out, err := app.Run(g.dir, "sh", "-c", s.Run)
		if len(out) > 0 {
			// The script's own output goes to the log, not to the result, so a result never
			// carries whatever a script happened to print.
			fmt.Fprintf(app.Stderr, "--- %s ---\n%s", name, out)
			if !strings.HasSuffix(string(out), "\n") {
				fmt.Fprintln(app.Stderr)
			}
		}
		if err == nil {
			passed++
			res.AddCheck(name, run.StatusPass, "exit 0")
			continue
		}
		var ee *exec.ExitError
		switch {
		case errors.As(err, &ee):
			if len(ee.Stderr) > 0 {
				fmt.Fprintf(app.Stderr, "--- %s stderr ---\n%s\n", name, strings.TrimRight(string(ee.Stderr), "\n"))
			}
			res.AddCheck(name, run.StatusFail, ee.Error())
		case strings.HasPrefix(err.Error(), "exit status "):
			// A runner that reports the script's exit code without an *exec.ExitError.
			res.AddCheck(name, run.StatusFail, err.Error())
		default:
			// sh could not start, which is a tool error and not a failing script.
			res.AddCheck(name, run.StatusError, err.Error())
		}
	}
	res.Facts["check.passed"] = passed
	res.Facts["check.failed"] = len(scripts) - passed
	return app.finish(res, c, g)
}
