package pipeline

import (
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/glassops-platform/glassops/internal/core/config"
)

func TestSplit(t *testing.T) {
	cases := []struct {
		in   string
		want []string
	}{
		{"plan", []string{"plan"}},
		{"  validate   --target ephemeral\t--fallback-target prod ", []string{"validate", "--target", "ephemeral", "--fallback-target", "prod"}},
		{`check --script "unit tests"`, []string{"check", "--script", "unit tests"}},
		{`check --script 'a "b" c'`, []string{"check", "--script", `a "b" c`}},
		{`x "say \"hi\" \\ \n"`, []string{"x", `say "hi" \ \n`}},
		{`x --name=a' 'b`, []string{"x", "--name=a b"}},
		{`x ""`, []string{"x", ""}},
		{`x ''`, []string{"x", ""}},
		{`x a\ b`, []string{"x", "a b"}},
		{`x C:\work\plan.json`, []string{"x", `C:\work\plan.json`}},
		{`x \"quoted\"`, []string{"x", `"quoted"`}},
		{`x $HOME;rm -rf /`, []string{"x", "$HOME;rm", "-rf", "/"}},
		{"", nil},
		{"   ", nil},
	}
	for _, tc := range cases {
		got, err := Split(tc.in)
		if err != nil {
			t.Errorf("Split(%q): %v", tc.in, err)
			continue
		}
		if !reflect.DeepEqual(got, tc.want) {
			t.Errorf("Split(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

func TestSplitUnterminated(t *testing.T) {
	for _, in := range []string{`x "open`, `x 'open`, `x "a\"`} {
		if _, err := Split(in); err == nil {
			t.Errorf("Split(%q) did not fail", in)
		}
	}
}

func TestParse(t *testing.T) {
	steps, err := Parse([]config.Step{
		{Run: "plan"},
		{Run: config.ExecStep, Exec: "./scripts/a.sh"},
		{Run: "validate --target 'my org'", When: "failure"},
		{Run: config.ExecStep, Exec: "echo done", When: "always"},
		{Run: "report", When: "success"},
	})
	if err != nil {
		t.Fatal(err)
	}
	want := []Step{
		{Index: 1, Name: "plan", Argv: []string{"plan"}, When: Success, Raw: "plan"},
		{Index: 2, Name: "exec", Exec: "./scripts/a.sh", ExecN: 1, When: Success, Raw: "exec"},
		{Index: 3, Name: "validate", Argv: []string{"validate", "--target", "my org"}, When: Failure, Raw: "validate --target 'my org'"},
		{Index: 4, Name: "exec", Exec: "echo done", ExecN: 2, When: Always, Raw: "exec"},
		{Index: 5, Name: "report", Argv: []string{"report"}, When: Success, Raw: "report"},
	}
	if !reflect.DeepEqual(steps, want) {
		t.Fatalf("Parse =\n%+v\nwant\n%+v", steps, want)
	}
	if !steps[1].IsExec() || steps[0].IsExec() {
		t.Fatal("IsExec wrong")
	}
}

func TestParseRejects(t *testing.T) {
	cases := []struct {
		name string
		step config.Step
		want string
	}{
		{"empty", config.Step{Run: "  "}, "empty step"},
		{"bad when", config.Step{Run: "plan", When: "sometimes"}, "not success, failure or always"},
		{"exec without run", config.Step{Run: "exec"}, "needs run"},
		{"exec with args", config.Step{Run: "exec ./a.sh", Exec: "./a.sh"}, "takes its command from run"},
		{"run on a command step", config.Step{Run: "plan", Exec: "./a.sh"}, "only for exec steps"},
		{"flag first", config.Step{Run: "--dir x plan"}, "command name"},
		{"nested run", config.Step{Run: "run pr"}, "run cannot be a pipeline step"},
		{"bad quote", config.Step{Run: `check --script "x`}, "unterminated"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := Parse([]config.Step{{Run: "plan"}, tc.step})
			if err == nil || !strings.Contains(err.Error(), tc.want) || !strings.Contains(err.Error(), "step 2") {
				t.Fatalf("err = %v, want step 2 and %q", err, tc.want)
			}
		})
	}
}

func TestParseReportsEveryProblem(t *testing.T) {
	_, err := Parse([]config.Step{{Run: ""}, {Run: "plan"}, {Run: "exec"}})
	if err == nil || !strings.Contains(err.Error(), "step 1") || !strings.Contains(err.Error(), "step 3") {
		t.Fatalf("err = %v", err)
	}
	if _, err := Parse(nil); err == nil {
		t.Fatal("an empty pipeline parsed")
	}
}

func TestResolve(t *testing.T) {
	cfg := &config.Config{Pipelines: map[string][]config.Step{"pr": {{Run: "plan"}}, "deploy": {{Run: "plan"}}}}
	steps, err := Resolve(cfg, "pr")
	if err != nil || len(steps) != 1 {
		t.Fatalf("Resolve: %v, %v", steps, err)
	}
	_, err = Resolve(cfg, "nightly")
	if !errors.Is(err, ErrUnknownPipeline) || !strings.Contains(err.Error(), "deploy, pr") {
		t.Fatalf("err = %v", err)
	}
	_, err = Resolve(&config.Config{}, "pr")
	if !errors.Is(err, ErrUnknownPipeline) || !strings.Contains(err.Error(), "none are configured") {
		t.Fatalf("err = %v", err)
	}
}

func TestDefaultPipelinesParse(t *testing.T) {
	for name, steps := range config.DefaultPipelines() {
		if _, err := Parse(steps); err != nil {
			t.Errorf("default pipeline %s: %v", name, err)
		}
	}
}
