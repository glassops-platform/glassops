package cli

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/glassops-platform/glassops/internal/ci"
	"github.com/glassops-platform/glassops/internal/core/exitcode"
	"github.com/glassops-platform/glassops/internal/core/pipeline"
	"github.com/glassops-platform/glassops/internal/core/run"
	"github.com/glassops-platform/glassops/internal/providers"
)

// ResultsDirEnv is the variable an exec step reads to find the results dir.
const ResultsDirEnv = "GLASSOPS_RESULTS_DIR"

type runOptions struct {
	dryRun bool
}

func newRunCommand(app *App, g *globals) *cobra.Command {
	o := &runOptions{}
	cmd := &cobra.Command{
		Use:   "run <pipeline>",
		Short: "Run a pipeline from glassops.yaml",
		Long: "run runs the steps of one pipeline in order in this process. A step's when decides\n" +
			"whether it runs after an earlier step failed: success (the default) runs only when\n" +
			"nothing before failed, failure runs only when something did, always runs anyway.\n" +
			"After plan writes an empty plan, later steps run only with when always. An exec\n" +
			"step runs its run command with sh -c in the project dir, with " + ResultsDirEnv + "\n" +
			"set, and writes exec.<n>.json. Every step gets this run's --dir, --config,\n" +
			"--results-dir and --ci. The exit code is the worst of all steps. A cancel stops\n" +
			"the current step and still runs the when always steps. With --dry-run it prints\n" +
			"the steps without running them.",
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return app.runPipeline(cmd.Context(), g, o, args[0])
		},
	}
	cmd.Flags().BoolVar(&o.dryRun, "dry-run", false, "print the steps without running them")
	return cmd
}

// shellRunner runs an exec step's command and returns its exit status. A command that
// cannot be started, or that was stopped by a signal or a cancel, returns an error. Tests
// replace it.
var shellRunner = func(ctx context.Context, dir string, env []string, command string, stdout, stderr io.Writer) (int, error) {
	cmd := exec.CommandContext(ctx, "sh", "-c", command)
	cmd.Dir = dir
	cmd.Env = env
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	// A child that keeps the output pipes open after a cancel cannot hold the run forever.
	cmd.WaitDelay = 10 * time.Second
	err := cmd.Run()
	if err == nil {
		return 0, nil
	}
	var ee *exec.ExitError
	if errors.As(err, &ee) && ee.ExitCode() >= 0 && ctx.Err() == nil {
		return ee.ExitCode(), nil
	}
	if ctx.Err() != nil {
		return -1, fmt.Errorf("stopped: %w", ctx.Err())
	}
	return -1, err
}

// cleanupTimeout bounds each when: always step after a cancel. Tests shorten it.
var cleanupTimeout = pipeline.DefaultCleanupTimeout

func (app *App) runPipeline(ctx context.Context, g *globals, o *runOptions, name string) error {
	proj, err := app.loadProject(g)
	if err != nil {
		return err
	}
	steps, err := pipeline.Resolve(proj.cfg, name)
	if err != nil {
		return &exitcode.ExitError{Code: exitcode.Error, Err: fmt.Errorf("pipeline %s: %w", name, err)}
	}
	globalArgs := g.args()

	if o.dryRun {
		fmt.Fprintf(app.Stdout, "pipeline %s, %d steps, each glassops step also gets: %s\n", name, len(steps), strings.Join(globalArgs, " "))
		for _, s := range steps {
			if s.IsExec() {
				fmt.Fprintf(app.Stdout, "%2d. [%s] exec: sh -c %q (result exec.%d.json)\n", s.Index, s.When, s.Exec, s.ExecN)
				continue
			}
			fmt.Fprintf(app.Stdout, "%2d. [%s] glassops %s\n", s.Index, s.When, strings.Join(s.Argv, " "))
		}
		return nil
	}

	c, err := app.detectCI(g)
	if err != nil {
		return &exitcode.ExitError{Code: exitcode.Error, Err: err}
	}
	res := run.New("run", app.Now)
	res.Provider = proj.provider.Name()
	res.Facts["run.pipeline"] = name

	total := len(steps)
	r := &pipeline.Runner{
		Command: app.Execute,
		Globals: globalArgs,
		Exec: func(ctx context.Context, s pipeline.Step) (int, string) {
			return app.execStep(ctx, g, c, s)
		},
		PlanEmpty: func() (bool, error) {
			cs, err := providers.ReadChangeSet(filepath.Join(g.dir, ".glassops", "plan.json"))
			if err != nil {
				return false, fmt.Errorf("read the plan the plan step wrote: %w", err)
			}
			return cs.Empty, nil
		},
		Before: func(s pipeline.Step) {
			fmt.Fprintf(app.Stdout, "==> [%d/%d] %s\n", s.Index, total, stepLabel(s))
		},
		Skipped: func(s pipeline.Step, reason string) {
			fmt.Fprintf(app.Stdout, "==> [%d/%d] %s: skipped, %s\n", s.Index, total, stepLabel(s), reason)
		},
		Now:            app.Now,
		CleanupTimeout: cleanupTimeout,
	}
	out := r.Run(ctx, steps)

	res.Facts["run.steps"] = out.Records
	res.Facts["run.exit_code"] = out.Exit
	res.Facts["run.empty_plan"] = out.EmptyPlan
	res.Facts["run.cancelled"] = out.Cancelled
	for _, rec := range out.Records {
		res.AddCheck(fmt.Sprintf("step.%d.%s", rec.Index, rec.Step), recordStatus(rec), recordDetail(rec))
		if rec.Result != "" {
			res.Artifacts = append(res.Artifacts, rec.Result)
		}
	}
	if out.Cancelled {
		res.AddError(errors.New("the run was cancelled"))
	}
	// The run's status follows the worst exit code, so the run exits with that code.
	res.Degrade(exitStatus(out.Exit))
	return app.finish(res, c, g)
}

// args are the global flags in the form every step receives, so each one reads the same
// project and writes to the same results dir as the run. The = form keeps an empty value
// as one argument.
func (g *globals) args() []string {
	return []string{
		"--dir=" + g.dir,
		"--config=" + g.configPath,
		"--results-dir=" + g.resultsDir,
		"--ci=" + g.ciName,
	}
}

// execStep runs one exec step with sh -c in the project dir and writes exec.<n>.json. It
// returns the exit code and the result path.
func (app *App) execStep(ctx context.Context, g *globals, c ci.CI, s pipeline.Step) (int, string) {
	step := fmt.Sprintf("exec.%d", s.ExecN)
	res := run.New(step, app.Now)
	res.Facts["exec.run"] = s.Exec
	res.Facts["exec.index"] = s.Index
	path := filepath.Join(g.results(), step+".json")

	results, err := filepath.Abs(g.results())
	if err == nil {
		err = os.MkdirAll(results, 0o755)
	}
	if err != nil {
		res.AddCheck("command", run.StatusError, fmt.Sprintf("%s: results dir: %v", s.Exec, err))
		return app.finishCode(res, c, g), path
	}
	env := append(os.Environ(), ResultsDirEnv+"="+results)
	code, err := shellRunner(ctx, g.dir, env, s.Exec, app.Stdout, app.Stderr)
	if err != nil {
		res.AddCheck("command", run.StatusError, fmt.Sprintf("%s: %v", s.Exec, err))
		return app.finishCode(res, c, g), path
	}
	res.Facts["exec.exit_code"] = code
	res.AddCheck("command", execStatus(code), fmt.Sprintf("%s: exit %d", s.Exec, code))
	return app.finishCode(res, c, g), path
}

// finishCode is finish for a step inside run. It prints any error and returns only the
// exit code, which is all the runner needs.
func (app *App) finishCode(res *run.Result, c ci.CI, g *globals) int {
	err := app.finish(res, c, g)
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
	fmt.Fprintln(app.Stderr, "glassops:", err)
	return exitcode.Error
}

// execStatus reads an exec command's exit status the way glassops reads its own: 0 pass,
// 1 fail, anything else an error.
func execStatus(code int) run.Status {
	switch code {
	case 0:
		return run.StatusPass
	case 1:
		return run.StatusFail
	}
	return run.StatusError
}

func exitStatus(code int) run.Status {
	switch code {
	case exitcode.OK:
		return run.StatusPass
	case exitcode.Fail:
		return run.StatusFail
	}
	return run.StatusError
}

func recordStatus(rec pipeline.Record) run.Status {
	if rec.Skipped {
		return run.StatusSkipped
	}
	if rec.Error != "" {
		return run.StatusError
	}
	return exitStatus(*rec.ExitCode)
}

func recordDetail(rec pipeline.Record) string {
	if rec.Skipped {
		return "skipped, " + rec.SkipReason
	}
	d := fmt.Sprintf("exit %d in %s", *rec.ExitCode, time.Duration(rec.DurationMS)*time.Millisecond)
	if rec.Error != "" {
		d += ", " + rec.Error
	}
	return d
}

func stepLabel(s pipeline.Step) string {
	if s.IsExec() {
		return fmt.Sprintf("exec.%d: %s", s.ExecN, s.Exec)
	}
	return strings.Join(s.Argv, " ")
}
