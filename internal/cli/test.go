package cli

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/glassops-platform/glassops/internal/core/exitcode"
	"github.com/glassops-platform/glassops/internal/core/junit"
	"github.com/glassops-platform/glassops/internal/core/run"
	"github.com/glassops-platform/glassops/internal/providers"
)

type testOptions struct {
	target    string
	level     string
	tests     []string
	wait      time.Duration
	junitPath string
	resumeID  string
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
			return app.test(cmd.Context(), g, o)
		},
	}
	f := cmd.Flags()
	f.StringVar(&o.target, "target", "", "target name from glassops.yaml (required)")
	f.StringVar(&o.level, "level", "", "test level (default test.level for the tier, else test.level.default)")
	f.StringSliceVar(&o.tests, "tests", nil, "tests to run, comma separated, for a level that names them")
	f.DurationVar(&o.wait, "wait", 0, "how long to wait for the run, such as 30m (default the provider's)")
	f.StringVar(&o.junitPath, "junit", "", "JUnit XML file to write (default <results-dir>/test.junit.xml)")
	f.StringVar(&o.resumeID, "resume", "", "run Id of an earlier test run to resume instead of starting a new one")
	_ = cmd.MarkFlagRequired("target")
	return cmd
}

// test runs the tests on the target and applies the gate: a failing test, a run that did not
// finish, org coverage below test.coverage.org_min, or a class below test.coverage.class_min
// fails. Coverage that is required but not reported also fails.
func (app *App) test(ctx context.Context, g *globals, o *testOptions) error {
	c, err := app.detectCI(g)
	if err != nil {
		return &exitcode.ExitError{Code: exitcode.Error, Err: err}
	}
	proj, err := app.loadProject(g)
	if err != nil {
		return err
	}
	res := run.New("test", app.Now)
	res.Provider = proj.provider.Name()

	tester, ok := proj.provider.(providers.Tester)
	if !ok {
		res.AddCheck("provider", run.StatusError, fmt.Sprintf("provider %s cannot run tests", proj.provider.Name()))
		return app.finish(res, c, g)
	}
	t, err := app.target(proj.cfg, o.target)
	if err != nil {
		res.AddCheck("target", run.StatusError, err.Error())
		return app.finish(res, c, g)
	}
	res.Target = t.Name
	res.Facts["target.source"] = targetSource(t)
	sess, _, err := app.connect(ctx, proj.provider, t)
	if err != nil {
		res.AddCheck("target", run.StatusError, err.Error())
		return app.finish(res, c, g)
	}
	defer sess.Close()

	level := o.level
	if level == "" {
		level = testLevelFor(proj.cfg.Test.Level, t.Tier)
	}
	out, err := tester.RunTests(ctx, sess, providers.TestRequest{
		Dir: g.dir, TestLevel: level, Tests: o.tests, Wait: o.wait, ResumeID: o.resumeID,
	})
	if err != nil {
		res.AddCheck("tests", run.StatusError, err.Error())
		return app.finish(res, c, g)
	}

	var unreported []string
	if proj.cfg.Test.Coverage.ClassMin != nil {
		var perr error
		out, unreported, perr = addUncoveredClasses(g, proj.cfg.Test.Coverage.ClassScope, out)
		if perr != nil {
			res.AddCheck("coverage.plan", run.StatusFail, perr.Error())
			res.Degrade(run.StatusFail)
		}
	}
	if len(unreported) > 0 {
		res.Facts["test.coverage_unreported"] = unreported
	}

	for k, v := range out.Facts {
		res.Facts[k] = v
	}
	res.Facts["test.ran"] = out.Ran
	res.Facts["test.failed"] = out.Failed
	res.Facts["test.finished"] = out.Finished
	if out.RunID != "" {
		res.Facts["test.run_id"] = out.RunID
	}
	if out.OrgCoverage >= 0 {
		res.Facts["test.org_coverage"] = out.OrgCoverage
	}
	for _, cl := range out.Classes {
		res.Facts["test.coverage."+cl.Name+".percent"] = cl.Percent
	}

	if out.Facts["test.none_found"] == true {
		path, err := app.writeJUnit(g, o, out)
		if err == nil {
			res.Artifacts = append(res.Artifacts, path)
		}
		cov := proj.cfg.Test.Coverage
		if cov.OrgMin != nil || cov.ClassMin != nil {
			res.AddCheck("tests", run.StatusFail, fmt.Sprintf("no Apex tests ran in %s, so coverage cannot be checked", t.Name))
			res.Degrade(run.StatusFail)
		} else {
			res.AddCheck("tests", run.StatusSkipped, "no Apex tests were found")
			res.Status = run.StatusEmpty
		}
		return app.finish(res, c, g)
	}

	if path, err := app.writeJUnit(g, o, out); err != nil {
		res.AddCheck("junit", run.StatusError, err.Error())
	} else {
		res.Artifacts = append(res.Artifacts, path)
		res.AddCheck("junit", run.StatusPass, path)
	}

	applyTestGate(res, proj.cfg.Test.Coverage.OrgMin, proj.cfg.Test.Coverage.ClassMin, out)
	return app.finish(res, c, g)
}

// stringList reads a fact holding a list of strings, as a provider sets it ([]string) or as it
// reads back from JSON ([]any).
func stringList(v any) []string {
	switch l := v.(type) {
	case []string:
		return l
	case []any:
		var out []string
		for _, e := range l {
			if name, ok := e.(string); ok {
				out = append(out, name)
			}
		}
		return out
	}
	return nil
}

// addUncoveredClasses counts every Apex class the run reported no coverage for as 0%, because
// sf lists only classes a test touched. Which classes depends on scope. "changed" (the
// default) takes the changed classes in .glassops/plan.json, and when there is no plan it takes
// every class the provider found untouched in the org's coverage
// aggregate (fact test.coverage_untouched), since without a plan nothing says which classes
// changed. "all" takes the plan's classes and every untouched class. A class whose name ends in
// Test or _Test is taken for a test class and skipped, the same reading the gate's coverage_min
// uses. It returns the classes it added. A plan that exists and cannot be read is an error, and
// so is an aggregate the provider could not read when the scope needs it, so a broken input
// never weakens the check.
func addUncoveredClasses(g *globals, scope string, out providers.TestResult) (providers.TestResult, []string, error) {
	var classes []string
	cs, err := providers.ReadChangeSet(filepath.Join(g.dir, ".glassops", "plan.json"))
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		return out, nil, fmt.Errorf("the changed classes could not be read, so unreported classes are not counted at 0%%: %w", err)
	}
	havePlan := err == nil
	if havePlan {
		classes = stringList(cs.Facts["salesforce.apex_classes"])
	}
	if scope == "all" || !havePlan {
		if why, ok := out.Facts["test.coverage_unavailable"]; ok {
			return out, nil, fmt.Errorf("the org's coverage aggregate could not be read, so classes no test touched are not counted at 0%%: %v", why)
		}
		classes = append(append([]string(nil), classes...), stringList(out.Facts["test.coverage_untouched"])...)
	}
	reported := map[string]bool{}
	for _, cl := range out.Classes {
		reported[cl.Name] = true
	}
	var added []string
	for _, name := range classes {
		if reported[name] || strings.HasSuffix(name, "Test") {
			continue
		}
		reported[name] = true
		added = append(added, name)
	}
	sort.Strings(added)
	merged := append([]providers.ClassCoverage(nil), out.Classes...)
	for _, name := range added {
		merged = append(merged, providers.ClassCoverage{Name: name, Percent: 0})
	}
	out.Classes = merged
	return out, added, nil
}

// applyTestGate adds the checks for a finished or unfinished test run.
func applyTestGate(res *run.Result, orgMin, classMin *int, out providers.TestResult) {
	fail := func(name, detail string) {
		res.AddCheck(name, run.StatusFail, detail)
		res.Degrade(run.StatusFail)
	}
	switch {
	case !out.Finished:
		id := out.RunID
		if id == "" {
			id = "unknown"
		}
		fail("run", fmt.Sprintf("the test run did not finish, run Id %s (resume with --resume)", id))
	case out.Failed > 0 || !out.Passed:
		fail("tests", fmt.Sprintf("%d of %d tests failed", out.Failed, out.Ran))
		for i, f := range out.Failures {
			res.AddCheck(fmt.Sprintf("failure.%d", i+1), run.StatusFail, f)
		}
	default:
		res.AddCheck("tests", run.StatusPass, fmt.Sprintf("%d tests passed", out.Ran))
	}

	if orgMin != nil {
		switch {
		case out.OrgCoverage < 0:
			fail("coverage.org", fmt.Sprintf("org coverage was not reported, and test.coverage.org_min is %d%%", *orgMin))
		case out.OrgCoverage < *orgMin:
			fail("coverage.org", fmt.Sprintf("org coverage %d%% is below test.coverage.org_min %d%%", out.OrgCoverage, *orgMin))
		default:
			res.AddCheck("coverage.org", run.StatusPass, fmt.Sprintf("org coverage %d%% meets %d%%", out.OrgCoverage, *orgMin))
		}
	}
	if classMin != nil {
		classes := append([]providers.ClassCoverage(nil), out.Classes...)
		sort.Slice(classes, func(i, j int) bool { return classes[i].Name < classes[j].Name })
		if len(classes) == 0 {
			fail("coverage.class", fmt.Sprintf("no class coverage was reported, and test.coverage.class_min is %d%%", *classMin))
		}
		low := 0
		for _, cl := range classes {
			if cl.Percent < *classMin {
				low++
				fail("coverage.class."+cl.Name, fmt.Sprintf("%s coverage %d%% is below test.coverage.class_min %d%%", cl.Name, cl.Percent, *classMin))
			}
		}
		if len(classes) > 0 && low == 0 {
			res.AddCheck("coverage.class", run.StatusPass, fmt.Sprintf("all %d classes meet %d%%", len(classes), *classMin))
		}
	}
}

// writeJUnit writes the provider's JUnit XML, or builds a summary from the counts when the
// provider gave none, and returns the path.
func (app *App) writeJUnit(g *globals, o *testOptions, out providers.TestResult) (string, error) {
	path := o.junitPath
	if path == "" {
		path = filepath.Join(g.results(), "test.junit.xml")
	}
	data := out.JUnit
	if len(data) == 0 {
		var s junit.Suites
		s.Name = "tests"
		cs := junit.Suite{Name: "tests"}
		for i := 0; i < out.Ran-out.Failed; i++ {
			cs.Cases = append(cs.Cases, junit.Case{Name: fmt.Sprintf("passed.%d", i+1)})
		}
		for i, f := range out.Failures {
			cs.Cases = append(cs.Cases, junit.Case{Name: fmt.Sprintf("failed.%d", i+1), Failure: &junit.Failure{Message: f}})
		}
		s.Add(cs)
		b, err := s.Marshal()
		if err != nil {
			return "", err
		}
		data = b
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return "", fmt.Errorf("write junit: %w", err)
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return "", fmt.Errorf("write junit: %w", err)
	}
	return path, nil
}
