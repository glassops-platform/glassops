package cli

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/glassops-platform/glassops/internal/core/config"
	"github.com/glassops-platform/glassops/internal/core/exitcode"
	"github.com/glassops-platform/glassops/internal/core/pipeline"
	"github.com/glassops-platform/glassops/internal/core/run"
	"github.com/glassops-platform/glassops/internal/providers"
)

// shellCall is one synthetic exec step call.
type shellCall struct {
	dir, command, resultsDir string
}

// fakeShell replaces shellRunner. Each command's exit code comes from codes, and a command
// in fail is reported as not startable.
func fakeShell(t *testing.T, codes map[string]int, calls *[]shellCall, hook func(string)) {
	t.Helper()
	saved := shellRunner
	t.Cleanup(func() { shellRunner = saved })
	shellRunner = func(ctx context.Context, dir string, env []string, command string, stdout, _ io.Writer) (int, error) {
		c := shellCall{dir: dir, command: command}
		for _, e := range env {
			if v, ok := strings.CutPrefix(e, ResultsDirEnv+"="); ok {
				c.resultsDir = v
			}
		}
		*calls = append(*calls, c)
		io.WriteString(stdout, "output of "+command+"\n")
		if hook != nil {
			hook(command)
		}
		if ctx.Err() != nil {
			return -1, ctx.Err()
		}
		return codes[command], nil
	}
}

// runHarness is a harness whose fake provider plans cs from the full source, and whose
// config at a non-default path holds the given pipelines.
func runHarness(t *testing.T, cs providers.ChangeSet, pipelines string) (*harness, *providers.PlanRequest, string) {
	t.Helper()
	cs.Schema = providers.PlanSchema
	h, req, _ := planHarness(t, cs, nil, gitAnswers(nil))
	cfgPath := filepath.Join(h.dir, "ci", "pipelines.yaml")
	h.write(t, filepath.Join("ci", "pipelines.yaml"), "version: 1\nsalesforce:\n  delta:\n    enabled: false\npipelines:\n"+pipelines)
	return h, req, cfgPath
}

func readResult(t *testing.T, path string) run.Result {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var r run.Result
	if err := json.Unmarshal(data, &r); err != nil {
		t.Fatal(err)
	}
	return r
}

// runSteps decodes run.json's step records.
func runSteps(t *testing.T, r run.Result) []pipeline.Record {
	t.Helper()
	data, err := json.Marshal(r.Facts["run.steps"])
	if err != nil {
		t.Fatal(err)
	}
	var recs []pipeline.Record
	if err := json.Unmarshal(data, &recs); err != nil {
		t.Fatal(err)
	}
	return recs
}

const runPipelineYAML = `  pr:
    - plan
    - step: exec
      run: ./publish.sh
    - help
    - step: exec
      run: notify-failure
      when: failure
    - step: exec
      run: lint
    - no-such-command
    - step: exec
      run: cleanup
      when: always
`

func TestRunPipelineEndToEnd(t *testing.T) {
	var calls []shellCall
	fakeShell(t, map[string]int{"lint": 1}, &calls, nil)
	h, req, cfgPath := runHarness(t, providers.ChangeSet{Manifest: "package.xml"}, runPipelineYAML)
	results := filepath.Join(t.TempDir(), "out")

	code := Execute(h.app, []string{"run", "pr", "--dir", h.dir, "--config", cfgPath, "--results-dir", results, "--ci", "local"})
	if code != exitcode.Fail {
		t.Fatalf("exit %d, want the worst step's 1, output:\n%s", code, h.out)
	}
	// plan read the --config file, which turns the delta off.
	if !req.Full {
		t.Fatalf("plan did not read the run's --config: %+v", *req)
	}

	r := readResult(t, filepath.Join(results, "run.json"))
	if r.Status != run.StatusFail || r.Facts["run.pipeline"] != "pr" {
		t.Fatalf("run result %+v", r)
	}
	recs := runSteps(t, r)
	type got struct {
		step    string
		exit    int
		skipped string
	}
	var have []got
	for _, rec := range recs {
		g := got{step: rec.Step, skipped: rec.SkipReason}
		if rec.ExitCode != nil {
			g.exit = *rec.ExitCode
		} else {
			g.exit = -1
		}
		have = append(have, g)
	}
	want := []got{
		{"plan", 0, ""},
		{"exec", 0, ""},
		{"help", 0, ""},
		{"exec", -1, pipeline.SkipNoFailure},
		{"exec", 1, ""},
		{"no-such-command", -1, pipeline.SkipEarlierFailed},
		{"exec", 0, ""},
	}
	if !reflect.DeepEqual(have, want) {
		t.Fatalf("steps\n%+v\nwant\n%+v", have, want)
	}
	wantArgv := []string{"--dir=" + h.dir, "--config=" + cfgPath, "--results-dir=" + results, "--ci=local", "plan"}
	if !reflect.DeepEqual(recs[0].Argv, wantArgv) {
		t.Fatalf("plan argv = %q, want %q", recs[0].Argv, wantArgv)
	}

	// Every step wrote to the run's results dir, and nothing to the default one.
	readResult(t, filepath.Join(results, "plan.json"))
	if _, err := os.Stat(filepath.Join(h.dir, ".glassops", "results")); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("a step wrote to the default results dir: %v", err)
	}

	// Exec steps ran in the project dir with GLASSOPS_RESULTS_DIR set, and their output passed through.
	absResults, _ := filepath.Abs(results)
	gotCmds := []string{}
	for _, c := range calls {
		gotCmds = append(gotCmds, c.command)
		if c.dir != h.dir || c.resultsDir != absResults {
			t.Fatalf("exec call %+v, want dir %s and results %s", c, h.dir, absResults)
		}
	}
	if !reflect.DeepEqual(gotCmds, []string{"./publish.sh", "lint", "cleanup"}) {
		t.Fatalf("exec calls = %v", gotCmds)
	}
	if !strings.Contains(h.out.String(), "output of ./publish.sh") {
		t.Fatalf("exec output not passed through:\n%s", h.out)
	}

	// Each exec step has its own envelope, numbered among the exec steps.
	for n, tc := range map[string]struct {
		run    string
		status run.Status
	}{
		"exec.1": {"./publish.sh", run.StatusPass},
		"exec.3": {"lint", run.StatusFail},
		"exec.4": {"cleanup", run.StatusPass},
	} {
		er := readResult(t, filepath.Join(results, n+".json"))
		if er.Step != n || er.Status != tc.status || er.Facts["exec.run"] != tc.run || len(er.Checks) != 1 || er.Checks[0].Name != "command" || !strings.Contains(er.Checks[0].Detail, tc.run) {
			t.Fatalf("%s envelope %+v", n, er)
		}
	}
	if _, err := os.Stat(filepath.Join(results, "exec.2.json")); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("the skipped exec step wrote a result: %v", err)
	}
}

func TestRunEmptyPlanSkipsLaterSteps(t *testing.T) {
	var calls []shellCall
	fakeShell(t, map[string]int{}, &calls, nil)
	h, _, cfgPath := runHarness(t, providers.ChangeSet{Empty: true}, runPipelineYAML)

	if code := h.exec("run", "pr", "--config", cfgPath); code != exitcode.OK {
		t.Fatalf("exit %d, output:\n%s", code, h.out)
	}
	if len(calls) != 1 || calls[0].command != "cleanup" {
		t.Fatalf("exec calls %+v, want only the always step", calls)
	}
	r := h.result(t, "run")
	if r.Status != run.StatusPass || r.Facts["run.empty_plan"] != true {
		t.Fatalf("run result %+v", r)
	}
	for _, rec := range runSteps(t, r)[1:6] {
		if !rec.Skipped || rec.SkipReason != pipeline.SkipEmptyPlan {
			t.Fatalf("record %+v", rec)
		}
	}
	if checkStatus(r, "step.2.exec") != run.StatusSkipped || checkStatus(r, "step.7.exec") != run.StatusPass {
		t.Fatalf("checks %+v", r.Checks)
	}
}

func TestRunCancelStillRunsAlways(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	var calls []shellCall
	// SIGTERM arrives while ./publish.sh runs.
	fakeShell(t, map[string]int{}, &calls, func(c string) {
		if c == "./publish.sh" {
			cancel()
		}
	})
	saved := cleanupTimeout
	cleanupTimeout = 5 * time.Second
	t.Cleanup(func() { cleanupTimeout = saved })
	h, _, cfgPath := runHarness(t, providers.ChangeSet{Manifest: "package.xml"}, runPipelineYAML)

	code := ExecuteContext(ctx, h.app, []string{"run", "pr", "--dir", h.dir, "--config", cfgPath, "--ci", "local"})
	if code != exitcode.Error {
		t.Fatalf("exit %d, output:\n%s", code, h.out)
	}
	var cmds []string
	for _, c := range calls {
		cmds = append(cmds, c.command)
	}
	if !reflect.DeepEqual(cmds, []string{"./publish.sh", "cleanup"}) {
		t.Fatalf("exec calls %v", cmds)
	}
	r := h.result(t, "run")
	if r.Facts["run.cancelled"] != true || r.Status != run.StatusError {
		t.Fatalf("run result %+v", r)
	}
	if h.result(t, "exec.1").Status != run.StatusError || h.result(t, "exec.4").Status != run.StatusPass {
		t.Fatal("exec envelopes do not show the stopped step and the cleanup")
	}
	recs := runSteps(t, r)
	if recs[2].SkipReason != pipeline.SkipCancelled || recs[4].SkipReason != pipeline.SkipCancelled {
		t.Fatalf("records %+v", recs)
	}
}

func TestRunDryRunRunsNothing(t *testing.T) {
	var calls []shellCall
	fakeShell(t, map[string]int{}, &calls, nil)
	h, req, cfgPath := runHarness(t, providers.ChangeSet{}, runPipelineYAML)

	if code := h.exec("run", "pr", "--config", cfgPath, "--dry-run"); code != exitcode.OK {
		t.Fatalf("exit %d, output:\n%s", code, h.out)
	}
	out := h.out.String()
	for _, want := range []string{
		"pipeline pr, 7 steps",
		"--dir=" + h.dir,
		" 1. [success] glassops plan",
		` 2. [success] exec: sh -c "./publish.sh" (result exec.1.json)`,
		` 4. [failure] exec: sh -c "notify-failure" (result exec.2.json)`,
		` 7. [always] exec: sh -c "cleanup" (result exec.4.json)`,
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("dry run output missing %q:\n%s", want, out)
		}
	}
	if len(calls) != 0 || req.OutDir != "" {
		t.Fatalf("dry run ran something: exec %v, plan %+v", calls, *req)
	}
	if _, err := os.Stat(filepath.Join(h.dir, ".glassops")); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("dry run wrote files: %v", err)
	}
}

func TestRunRefusesBadPipelines(t *testing.T) {
	cases := map[string]struct {
		pipeline string
		yaml     string
		want     string
	}{
		"unknown":   {"nightly", "  pr:\n    - plan\n", `unknown pipeline "nightly", configured: pr`},
		"nested":    {"pr", "  pr:\n    - run pr\n", "run cannot be a pipeline step"},
		"bad quote": {"pr", "  pr:\n    - check --script \"x\n", "unterminated"},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			h, _, cfgPath := runHarness(t, providers.ChangeSet{}, tc.yaml)
			if code := h.exec("run", tc.pipeline, "--config", cfgPath); code != exitcode.Error {
				t.Fatalf("exit %d", code)
			}
			if !strings.Contains(h.out.String(), tc.want) {
				t.Fatalf("output missing %q:\n%s", tc.want, h.out)
			}
		})
	}
}

// TestDefaultPipelinesUseRealCommands resolves every default step against the real command
// tree, so a default that names a missing command or flag fails here.
func TestDefaultPipelinesUseRealCommands(t *testing.T) {
	root := NewRootCommand(newHarness(t).app)
	for name, steps := range config.DefaultPipelines() {
		resolved, err := pipeline.Parse(steps)
		if err != nil {
			t.Fatalf("%s: %v", name, err)
		}
		for _, s := range resolved {
			c, rest, err := root.Find(s.Argv)
			if err != nil || c == root {
				t.Fatalf("%s step %d %q: no such command: %v", name, s.Index, s.Raw, err)
			}
			if err := c.ParseFlags(rest); err != nil {
				t.Fatalf("%s step %d %q: %v", name, s.Index, s.Raw, err)
			}
			if args := c.Flags().Args(); len(args) > 0 {
				t.Fatalf("%s step %d %q: stray arguments %v", name, s.Index, s.Raw, args)
			}
		}
	}
}

func TestShellRunnerWithSh(t *testing.T) {
	if _, err := exec.LookPath("sh"); err != nil {
		t.Skip("no sh on this machine")
	}
	dir := t.TempDir()
	var out strings.Builder
	env := append(os.Environ(), ResultsDirEnv+"=/results")
	code, err := shellRunner(context.Background(), dir, env, `printf '%s' "$GLASSOPS_RESULTS_DIR"; exit 3`, &out, io.Discard)
	if err != nil || code != 3 || out.String() != "/results" {
		t.Fatalf("code %d, err %v, out %q", code, err, out.String())
	}
	if code, err := shellRunner(context.Background(), dir, env, "true", io.Discard, io.Discard); err != nil || code != 0 {
		t.Fatalf("true: code %d, err %v", code, err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()
	// No output writers, so no pipe outlives the killed shell.
	if _, err := shellRunner(ctx, dir, env, "sleep 5", nil, nil); err == nil {
		t.Fatal("a cancelled command reported an exit status")
	}
}

func TestExecStatus(t *testing.T) {
	for code, want := range map[int]run.Status{0: run.StatusPass, 1: run.StatusFail, 2: run.StatusFail, 127: run.StatusFail} {
		if got := execStatus(code); got != want {
			t.Errorf("execStatus(%d) = %s, want %s", code, got, want)
		}
	}
}

func TestExecNonZeroExitIsFailAndStartErrorIsError(t *testing.T) {
	var calls []shellCall
	fakeShell(t, map[string]int{"exit3": 3}, &calls, nil)
	saved := shellRunner
	shellRunner = func(ctx context.Context, dir string, env []string, command string, stdout, stderr io.Writer) (int, error) {
		if command == "missing" {
			return -1, errors.New("exec: \"sh\": executable file not found")
		}
		return saved(ctx, dir, env, command, stdout, stderr)
	}
	h, _, cfgPath := runHarness(t, providers.ChangeSet{Manifest: "package.xml"},
		"  pr:\n    - step: exec\n      run: exit3\n    - step: exec\n      run: missing\n      when: always\n")

	code := h.exec("run", "pr", "--config", cfgPath)
	if code != exitcode.Error {
		t.Fatalf("exit %d, output:\n%s", code, h.out)
	}
	if r := h.result(t, "exec.1"); r.Status != run.StatusFail {
		t.Fatalf("exit 3 gave %s, want fail", r.Status)
	}
	if r := h.result(t, "exec.2"); r.Status != run.StatusError {
		t.Fatalf("a command that could not start gave %s, want error", r.Status)
	}
	recs := runSteps(t, h.result(t, "run"))
	if *recs[0].ExitCode != exitcode.Fail || *recs[1].ExitCode != exitcode.Error {
		t.Fatalf("records %+v", recs)
	}
}

func TestRunRemovesOnlyStaleResults(t *testing.T) {
	h, _, cfgPath := runHarness(t, providers.ChangeSet{Manifest: "package.xml"}, "  pr:\n    - help\n")
	dir := filepath.Join(".glassops", "results")
	stale := `{"schema": "glassops.result/v1", "step": "deploy", "status": "pass"}`
	h.write(t, filepath.Join(dir, "run.json"), "not even json")
	h.write(t, filepath.Join(dir, "deploy.json"), stale)
	h.write(t, filepath.Join(dir, "exec.9.json"), stale)
	keep := map[string]string{
		"analyze.sarif":                   `{"version": "2.1.0", "runs": []}`,
		"sarif.json":                      `{"$schema": "https://json.schemastore.org/sarif-2.1.0.json", "version": "2.1.0"}`,
		"other.json":                      `{"schema": "something.else/v1"}`,
		"broken.json":                     `{"schema": "glassops.result/v1"`,
		"notes.txt":                       `{"schema": "glassops.result/v1"}`,
		filepath.Join("nested", "x.json"): stale,
	}
	for rel, content := range keep {
		h.write(t, filepath.Join(dir, rel), content)
	}

	// A dry run removes nothing.
	if code := h.exec("run", "pr", "--config", cfgPath, "--dry-run"); code != exitcode.OK {
		t.Fatalf("dry run exit %d", code)
	}
	if _, err := os.Stat(filepath.Join(h.dir, dir, "deploy.json")); err != nil {
		t.Fatalf("dry run removed a result: %v", err)
	}

	if code := h.exec("run", "pr", "--config", cfgPath); code != exitcode.OK {
		t.Fatalf("exit %d, output:\n%s", code, h.out)
	}
	for _, gone := range []string{"deploy.json", "exec.9.json"} {
		if _, err := os.Stat(filepath.Join(h.dir, dir, gone)); !errors.Is(err, os.ErrNotExist) {
			t.Fatalf("stale %s survived: %v", gone, err)
		}
	}
	for rel, content := range keep {
		data, err := os.ReadFile(filepath.Join(h.dir, dir, rel))
		if err != nil || string(data) != content {
			t.Fatalf("%s was touched: %v", rel, err)
		}
	}
	r := h.result(t, "run")
	if r.Facts["run.stale_removed"] != float64(3) {
		t.Fatalf("run.stale_removed = %v, want 3", r.Facts["run.stale_removed"])
	}
}
