package policy

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"testing"

	"github.com/glassops-platform/glassops/internal/core/config"
)

// The opa outputs below are synthetic, written here in the shape "opa eval --format json"
// printed for OPA 1.22.0 over examples/rego in the openpolicyagent/opa:latest-static image.

// syntheticOPADeny is one denial.
const syntheticOPADeny = `{"result":[{"expressions":[{"value":[{"detail":"the plan changes the protected profile Admin","rule":"protected_profile"}],"text":"data.glassops.deny","location":{"row":1,"col":1}}]}]}`

// syntheticOPAEmpty is an empty deny set.
const syntheticOPAEmpty = `{"result":[{"expressions":[{"value":[],"text":"data.glassops.deny","location":{"row":1,"col":1}}]}]}`

// syntheticOPAUndefined is what opa prints when no file defines data.glassops.deny.
const syntheticOPAUndefined = `{}`

// syntheticOPAParseError is a parse error, printed with exit 2.
const syntheticOPAParseError = `{"errors":[{"message":"unexpected eof token","code":"rego_parse_error","location":{"file":"/work/policy/deploy.rego","row":3,"col":0}},{"message":"unexpected eof token: expected rule key term","code":"rego_parse_error","location":{"file":"/work/policy/deploy.rego","row":3,"col":0}}]}`

// syntheticOPABool is deny written as a boolean rule instead of a set.
const syntheticOPABool = `{"result":[{"expressions":[{"value":true,"text":"data.glassops.deny","location":{"row":1,"col":1}}]}]}`

// syntheticValue wraps a deny value in the opa eval output shape.
func syntheticValue(value string) string {
	return `{"result":[{"expressions":[{"value":` + value + `,"text":"data.glassops.deny"}]}]}`
}

// TestMain lets the test binary stand in for opa when GLASSOPS_FAKE_OPA is set, so runOPA is
// tested through a real process.
func TestMain(m *testing.M) {
	if mode := os.Getenv("GLASSOPS_FAKE_OPA"); mode != "" {
		os.Exit(fakeOPAMain(mode))
	}
	os.Exit(m.Run())
}

// fakeOPAMain writes its working directory and arguments to GLASSOPS_FAKE_OPA_LOG, then prints
// GLASSOPS_FAKE_OPA_STDOUT and exits as mode says.
func fakeOPAMain(mode string) int {
	wd, _ := os.Getwd()
	if log := os.Getenv("GLASSOPS_FAKE_OPA_LOG"); log != "" {
		_ = os.WriteFile(log, []byte(wd+"\n"+strings.Join(os.Args[1:], "\n")), 0o644)
	}
	fmt.Fprint(os.Stdout, os.Getenv("GLASSOPS_FAKE_OPA_STDOUT"))
	if mode == "ok" {
		return 0
	}
	fmt.Fprintln(os.Stderr, "synthetic opa failure")
	code, _ := strconv.Atoi(mode)
	return code
}

// opaCall is what one fake opa run was given.
type opaCall struct {
	dir   string
	args  []string
	input string
}

// fakeOPA replaces RunOPA with a function that records the call, including the input file's
// contents, and returns out and err.
func fakeOPA(t *testing.T, out string, err error) *[]opaCall {
	t.Helper()
	calls := &[]opaCall{}
	old := RunOPA
	t.Cleanup(func() { RunOPA = old })
	RunOPA = func(_ context.Context, dir string, args ...string) ([]byte, error) {
		c := opaCall{dir: dir, args: args}
		for i, a := range args {
			if a == "--input" && i+1 < len(args) {
				data, rerr := os.ReadFile(args[i+1])
				if rerr != nil {
					t.Errorf("read the opa input: %v", rerr)
				}
				c.input = string(data)
			}
		}
		*calls = append(*calls, c)
		return []byte(out), err
	}
	return calls
}

// regoProject writes Rego files under a temp project dir and points RegoDir at it.
func regoProject(t *testing.T, files ...string) (string, config.Rego) {
	t.Helper()
	dir := t.TempDir()
	for _, f := range files {
		path := filepath.Join(dir, f)
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte("package glassops\n"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	old := RegoDir
	t.Cleanup(func() { RegoDir = old })
	RegoDir = func(*config.Config) string { return dir }
	return dir, config.Rego{Files: files}
}

func regoInput() Input {
	return Input{Facts: doc(map[string]any{"target.tier": "prod", "salesforce.profiles": []string{"Admin"}}), Config: &config.Config{}}
}

func TestRegoDenyEntryFails(t *testing.T) {
	dir, cfg := regoProject(t, "policy/deploy.rego", "policy/extra.rego")
	calls := fakeOPA(t, syntheticOPADeny, nil)
	in := regoInput()
	results, err := NewRego(cfg).Evaluate(context.Background(), in)
	if err != nil {
		t.Fatal(err)
	}
	want := []Result{{Rule: "rego.protected_profile", Outcome: Fail, Detail: "the plan changes the protected profile Admin"}}
	if !reflect.DeepEqual(results, want) {
		t.Fatalf("results = %+v", results)
	}
	if len(*calls) != 1 {
		t.Fatalf("opa ran %d times", len(*calls))
	}
	c := (*calls)[0]
	if c.dir != dir {
		t.Fatalf("opa ran in %q, want the project dir %q", c.dir, dir)
	}
	if len(c.args) != 10 {
		t.Fatalf("args = %q", c.args)
	}
	wantArgs := []string{"eval", "--format", "json", "--input", c.args[4], "--data", "policy/deploy.rego", "--data", "policy/extra.rego", "data.glassops.deny"}
	if !reflect.DeepEqual(c.args, wantArgs) || !filepath.IsAbs(c.args[4]) {
		t.Fatalf("args = %q", c.args)
	}
	wantInput, _ := json.Marshal(in.Facts)
	if c.input != string(wantInput) {
		t.Fatalf("opa input = %s, want the facts document %s", c.input, wantInput)
	}
	if _, err := os.Stat(c.args[4]); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("the input file was left behind: %v", err)
	}

	allow, all, err := EvaluateAll(context.Background(), []Evaluator{GoRules{Rules: []Rule{fixed{"a", []Result{{Rule: "a", Outcome: Pass}}}}}, NewRego(cfg)}, in)
	if allow || err != nil || len(all) != 2 || all[0].Rule != "a" || all[1].Rule != "rego.protected_profile" {
		t.Fatalf("allow %v, err %v, results %+v", allow, err, all)
	}
}

func TestRegoArgsExact(t *testing.T) {
	got := regoArgs("in.json", []string{"a.rego", "b.rego"})
	want := []string{"eval", "--format", "json", "--input", "in.json", "--data", "a.rego", "--data", "b.rego", "data.glassops.deny"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("args = %q", got)
	}
	got = regoArgs("", []string{"a.rego"})
	want = []string{"eval", "--format", "json", "--data", "a.rego", "data.glassops.deny"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("args without input = %q", got)
	}
}

func TestRegoEmptyDenyPasses(t *testing.T) {
	_, cfg := regoProject(t, "policy/deploy.rego")
	fakeOPA(t, syntheticOPAEmpty, nil)
	results, err := NewRego(cfg).Evaluate(context.Background(), regoInput())
	if err != nil {
		t.Fatal(err)
	}
	r := one1(t, results)
	if r.Rule != "rego" || r.Outcome != Pass || !strings.Contains(r.Detail, "policy/deploy.rego") || !strings.Contains(r.Detail, "sha256 ") {
		t.Fatalf("result = %+v", r)
	}
	pass := GoRules{Rules: []Rule{fixed{"a", []Result{{Rule: "a", Outcome: Pass}}}}}
	allow, _, err := EvaluateAll(context.Background(), []Evaluator{pass, NewRego(cfg)}, regoInput())
	if !allow || err != nil {
		t.Fatalf("allow %v, err %v", allow, err)
	}
	// The Rego pass cannot stand in for Go rules that decided nothing.
	allow, all, err := EvaluateAll(context.Background(), []Evaluator{GoRules{}, NewRego(cfg)}, regoInput())
	if allow || err != nil || len(all) != 2 || all[0].Rule != "go" || all[0].Outcome != Fail || all[1].Rule != "rego" || all[1].Outcome != Pass {
		t.Fatalf("empty Go rules: allow %v, err %v, results %+v", allow, err, all)
	}
}

func TestRegoDigestFollowsContent(t *testing.T) {
	dir, cfg := regoProject(t, "p.rego")
	fakeOPA(t, syntheticOPAEmpty, nil)
	first, _ := NewRego(cfg).Evaluate(context.Background(), regoInput())
	if err := os.WriteFile(filepath.Join(dir, "p.rego"), []byte("package glassops\n# changed\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	second, _ := NewRego(cfg).Evaluate(context.Background(), regoInput())
	if first[0].Detail == second[0].Detail {
		t.Fatalf("the detail did not change with the policy: %s", first[0].Detail)
	}
}

// TestRegoOnlyAddsDenials pins T19: the Go results are the same with Rego as without it, and
// no Rego output can pass a rule the Go rules failed.
func TestRegoOnlyAddsDenials(t *testing.T) {
	goRules := GoRules{Rules: []Rule{
		fixed{"plan_not_empty", []Result{{Rule: "plan_not_empty", Outcome: Fail, Detail: "empty"}}},
		fixed{"window", []Result{{Rule: "window.prod", Outcome: Pass}}},
	}}
	_, goOnly, _ := EvaluateAll(context.Background(), []Evaluator{goRules}, regoInput())
	_, cfg := regoProject(t, "p.rego")
	for _, out := range []string{
		syntheticOPAEmpty,
		// An entry that tries to name a Go rule and claim a pass still denies, under rego.
		syntheticValue(`[{"rule":"plan_not_empty","detail":"pass","outcome":"pass"}]`),
	} {
		fakeOPA(t, out, nil)
		allow, results, err := EvaluateAll(context.Background(), []Evaluator{goRules, NewRego(cfg)}, regoInput())
		if allow || err != nil {
			t.Fatalf("allow %v, err %v for %s", allow, err, out)
		}
		if !reflect.DeepEqual(results[:len(goOnly)], goOnly) {
			t.Fatalf("Go results changed: %+v, want %+v", results[:len(goOnly)], goOnly)
		}
		for _, r := range results[len(goOnly):] {
			if !strings.HasPrefix(r.Rule, "rego") {
				t.Fatalf("Rego added a result outside rego: %+v", r)
			}
			if r.Rule != "rego" && r.Outcome != Fail {
				t.Fatalf("a Rego denial did not fail: %+v", r)
			}
		}
	}
}

func TestRegoDenyEntries(t *testing.T) {
	_, cfg := regoProject(t, "p.rego")
	long := strings.Repeat("x", 600)
	fakeOPA(t, syntheticValue(`[
		{"rule":"a","detail":"owner admin@example.com asked"},
		{"rule":"b"},
		{"rule":"c","detail":{"count":2}},
		{"rule":"d","detail":"`+long+`"}
	]`), nil)
	results, err := NewRego(cfg).Evaluate(context.Background(), regoInput())
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 4 {
		t.Fatalf("results = %+v", results)
	}
	if strings.Contains(results[0].Detail, "admin@example.com") || !strings.Contains(results[0].Detail, "a***@example.com") {
		t.Fatalf("the email was not masked: %s", results[0].Detail)
	}
	if results[1].Detail != "denied by Rego" {
		t.Fatalf("no detail gave %q", results[1].Detail)
	}
	if results[2].Detail != `{"count":2}` {
		t.Fatalf("object detail gave %q", results[2].Detail)
	}
	if len(results[3].Detail) != maxRegoDetail+3 || !strings.HasSuffix(results[3].Detail, "...") {
		t.Fatalf("long detail is %d bytes", len(results[3].Detail))
	}
	for _, r := range results {
		if r.Outcome != Fail || !strings.HasPrefix(r.Rule, RegoPrefix) {
			t.Fatalf("result = %+v", r)
		}
	}
}

// TestRegoFailsClosed covers every way the evaluation cannot decide. Each is an error, and
// EvaluateAll denies with a failed rego check.
func TestRegoFailsClosed(t *testing.T) {
	cases := []struct {
		name   string
		out    string
		runErr error
		facts  bool
		files  []string
		want   string
		keeps  []string
	}{
		{"missing opa", "", ErrOPANotFound, true, nil, "opa is not installed", nil},
		{"opa parse error", syntheticOPAParseError, errors.New("opa: exit status 2"), true, nil, "deploy.rego:3: unexpected eof token (and 1 more)", nil},
		{"opa fails without output", "", errors.New("opa: exit status 2: boom"), true, nil, "exit status 2: boom", nil},
		{"errors with exit 0", syntheticOPAParseError, nil, true, nil, "unexpected eof token", nil},
		{"undefined deny", syntheticOPAUndefined, nil, true, nil, "data.glassops.deny is undefined", nil},
		{"deny is a boolean", syntheticOPABool, nil, true, nil, "is not a set", nil},
		{"output is not JSON", "not json", nil, true, nil, "read the opa eval output", nil},
		{"two results", `{"result":[{"expressions":[{"value":[]}]},{"expressions":[{"value":[]}]}]}`, nil, true, nil, "returned 2 results", nil},
		{"string entries", syntheticValue(`["no"]`), nil, true, nil, "entry 1", nil},
		{"entry without a rule", syntheticValue(`[{"rule":"ok"},{"detail":"x"}]`), nil, true, nil, "entry 2", []string{"rego.ok"}},
		{"rule name with a space", syntheticValue(`[{"rule":"bad name"}]`), nil, true, nil, "entry 1", nil},
		{"rule name not a string", syntheticValue(`[{"rule":7}]`), nil, true, nil, "entry 1", nil},
		{"no facts", syntheticOPAEmpty, nil, false, nil, "no facts", nil},
		{"missing file", syntheticOPAEmpty, nil, true, []string{"p.rego", "missing.rego"}, "missing.rego does not exist", nil},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, cfg := regoProject(t, "p.rego")
			if tc.files != nil {
				cfg.Files = tc.files
			}
			fakeOPA(t, tc.out, tc.runErr)
			in := regoInput()
			if !tc.facts {
				in.Facts = nil
			}
			ev := NewRego(cfg)
			results, err := ev.Evaluate(context.Background(), in)
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("err = %v, want it to mention %q", err, tc.want)
			}
			var got []string
			for _, r := range results {
				got = append(got, r.Rule)
			}
			if !reflect.DeepEqual(got, tc.keeps) {
				t.Fatalf("results = %v, want %v", got, tc.keeps)
			}
			allow, all, err := EvaluateAll(context.Background(), []Evaluator{GoRules{}, ev}, in)
			if allow || err == nil {
				t.Fatalf("allow %v, err %v", allow, err)
			}
			last := all[len(all)-1]
			if last.Rule != "rego" || last.Outcome != Fail {
				t.Fatalf("last result = %+v", last)
			}
		})
	}
}

func TestRegoNoFilesIsAnError(t *testing.T) {
	if _, err := NewRego(config.Rego{}).Evaluate(context.Background(), regoInput()); err == nil {
		t.Fatal("no files gave no error")
	}
}

func TestRegoWithoutRegoDirUsesPathsAsWritten(t *testing.T) {
	old := RegoDir
	t.Cleanup(func() { RegoDir = old })
	RegoDir = nil
	abs := filepath.Join(t.TempDir(), "p.rego")
	if err := os.WriteFile(abs, []byte("package glassops\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	calls := fakeOPA(t, syntheticOPAEmpty, nil)
	if _, err := NewRego(config.Rego{Files: []string{abs}}).Evaluate(context.Background(), regoInput()); err != nil {
		t.Fatal(err)
	}
	if (*calls)[0].dir != "" {
		t.Fatalf("dir = %q", (*calls)[0].dir)
	}
}

func TestCheckRego(t *testing.T) {
	dir, cfg := regoProject(t, "p.rego")
	run := func(out string, err error) OPARunner {
		return func(_ context.Context, d string, args ...string) ([]byte, error) {
			if d != dir || strings.Contains(strings.Join(args, " "), "--input") {
				t.Errorf("dir %q, args %q", d, args)
			}
			return []byte(out), err
		}
	}
	if err := CheckRego(context.Background(), run(syntheticOPAEmpty, nil), dir, cfg.Files); err != nil {
		t.Fatalf("a defined set: %v", err)
	}
	for name, tc := range map[string]struct {
		out   string
		err   error
		files []string
		want  string
	}{
		"undefined":    {syntheticOPAUndefined, nil, cfg.Files, "undefined"},
		"parse error":  {syntheticOPAParseError, errors.New("exit status 2"), cfg.Files, "unexpected eof token"},
		"opa failed":   {"", errors.New("exit status 2"), cfg.Files, "exit status 2"},
		"missing file": {syntheticOPAEmpty, nil, []string{"nope.rego"}, "does not exist"},
		"no files":     {syntheticOPAEmpty, nil, nil, "names no files"},
	} {
		if err := CheckRego(context.Background(), run(tc.out, tc.err), dir, tc.files); err == nil || !strings.Contains(err.Error(), tc.want) {
			t.Errorf("%s: err = %v, want %q", name, err, tc.want)
		}
	}
}

// fakeOPAProgram points runOPA at this test binary, which acts as opa in mode.
func fakeOPAProgram(t *testing.T, mode, stdout string) string {
	t.Helper()
	old := opaProgram
	t.Cleanup(func() { opaProgram = old })
	exe, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	opaProgram = exe
	log := filepath.Join(t.TempDir(), "opa.log")
	t.Setenv("GLASSOPS_FAKE_OPA", mode)
	t.Setenv("GLASSOPS_FAKE_OPA_STDOUT", stdout)
	t.Setenv("GLASSOPS_FAKE_OPA_LOG", log)
	return log
}

func TestRunOPAThroughAProcess(t *testing.T) {
	dir := t.TempDir()
	log := fakeOPAProgram(t, "ok", syntheticOPAEmpty)
	out, err := runOPA(context.Background(), dir, "eval", "--format", "json", "data.glassops.deny")
	if err != nil || string(out) != syntheticOPAEmpty {
		t.Fatalf("out %q, err %v", out, err)
	}
	data, err := os.ReadFile(log)
	if err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(string(data), "\n")
	gotDir, _ := filepath.EvalSymlinks(lines[0])
	wantDir, _ := filepath.EvalSymlinks(dir)
	if gotDir != wantDir || !reflect.DeepEqual(lines[1:], []string{"eval", "--format", "json", "data.glassops.deny"}) {
		t.Fatalf("opa saw %q", lines)
	}

	fakeOPAProgram(t, "2", syntheticOPAParseError)
	out, err = runOPA(context.Background(), dir, "eval")
	if err == nil || !strings.Contains(err.Error(), "synthetic opa failure") || string(out) != syntheticOPAParseError {
		t.Fatalf("out %q, err %v", out, err)
	}
}

func TestRunOPAMissingBinary(t *testing.T) {
	old := opaProgram
	t.Cleanup(func() { opaProgram = old })
	opaProgram = "glassops-no-such-opa"
	if _, err := runOPA(context.Background(), "", "version"); !errors.Is(err, ErrOPANotFound) {
		t.Fatalf("err = %v", err)
	}
}

// TestRegoEvaluatorThroughAProcess runs the evaluator end to end with the fake opa process,
// so the input file, the arguments and the exit code all pass through exec.
func TestRegoEvaluatorThroughAProcess(t *testing.T) {
	_, cfg := regoProject(t, "p.rego")
	fakeOPAProgram(t, "ok", syntheticOPADeny)
	results, err := NewRego(cfg).Evaluate(context.Background(), regoInput())
	if err != nil || len(results) != 1 || results[0].Rule != "rego.protected_profile" {
		t.Fatalf("results %+v, err %v", results, err)
	}
	fakeOPAProgram(t, "2", syntheticOPAParseError)
	if _, err := NewRego(cfg).Evaluate(context.Background(), regoInput()); err == nil || !strings.Contains(err.Error(), "unexpected eof token") {
		t.Fatalf("err = %v", err)
	}
}

// TestRegoRealOPA runs the real opa over examples/rego when opa is on PATH. It is skipped
// otherwise, and runs in the image, where opa ships.
func TestRegoRealOPA(t *testing.T) {
	if _, err := exec.LookPath("opa"); err != nil {
		t.Skip("opa is not on PATH")
	}
	root, err := filepath.Abs(filepath.Join("..", "..", "..", "examples", "rego"))
	if err != nil {
		t.Fatal(err)
	}
	old := RegoDir
	t.Cleanup(func() { RegoDir = old })
	RegoDir = func(*config.Config) string { return root }
	cfg := config.Rego{Files: []string{"deploy.rego"}}
	if err := CheckRego(context.Background(), RunOPA, root, cfg.Files); err != nil {
		t.Fatalf("CheckRego: %v", err)
	}
	for name, want := range map[string][]string{"facts-deny.json": {"rego.protected_profile"}, "facts-pass.json": {"rego"}} {
		data, err := os.ReadFile(filepath.Join(root, name))
		if err != nil {
			t.Fatal(err)
		}
		in := Input{Config: &config.Config{}}
		if err := json.Unmarshal(data, &in.Facts); err != nil {
			t.Fatal(err)
		}
		results, err := NewRego(cfg).Evaluate(context.Background(), in)
		if err != nil {
			t.Fatalf("%s: %v", name, err)
		}
		var got []string
		for _, r := range results {
			got = append(got, r.Rule)
		}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("%s: rules %v, want %v (%+v)", name, got, want, results)
		}
	}
	// A file in the wrong package leaves deny undefined, which fails closed.
	wrong := filepath.Join(t.TempDir(), "other.rego")
	if err := os.WriteFile(wrong, []byte("package other\n\ndeny contains \"x\" if true\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := CheckRego(context.Background(), RunOPA, root, []string{wrong}); err == nil || !strings.Contains(err.Error(), "undefined") {
		t.Fatalf("wrong package: %v", err)
	}
}
