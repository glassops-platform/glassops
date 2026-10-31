package cli

import (
	"flag"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"

	"github.com/glassops-platform/glassops/internal/core/config"
)

var updateGolden = flag.Bool("update", false, "rewrite the init golden files")

// lf drops carriage returns, because a Windows checkout may convert the line endings.
func lf(b []byte) string { return strings.ReplaceAll(string(b), "\r\n", "\n") }

func checkGolden(t *testing.T, name string, got []byte) {
	t.Helper()
	path := filepath.Join("templates", "golden", name)
	if *updateGolden {
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(lf(got)), 0o644); err != nil {
			t.Fatal(err)
		}
		return
	}
	want, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read golden (run with -update to create it): %v", err)
	}
	if lf(want) != lf(got) {
		t.Fatalf("%s differs from the golden file. Run go test ./internal/cli -run Golden -update and read the diff.\n--- got ---\n%s", name, got)
	}
}

func TestInitConfigGolden(t *testing.T) {
	for name, data := range map[string]initData{
		"glassops.salesforce.yaml": {Provider: "salesforce"},
		"glassops.neutral.yaml":    {},
	} {
		t.Run(name, func(t *testing.T) {
			got, err := render(configTemplate, data)
			if err != nil {
				t.Fatal(err)
			}
			checkGolden(t, name, got)
			if _, err := config.Parse(got); err != nil {
				t.Fatalf("generated config does not parse: %v", err)
			}
		})
	}
}

// The pipelines the template writes are the default pipelines: deploy validates, gates,
// deploys and reports, and pr ends with an env delete that runs even after a failure.
func TestInitConfigPipelines(t *testing.T) {
	got, err := render(configTemplate, initData{Provider: "salesforce"})
	if err != nil {
		t.Fatal(err)
	}
	cfg, err := config.Parse(got)
	if err != nil {
		t.Fatal(err)
	}
	want := map[string][]string{
		"pr": {
			"plan", "check", "analyze", "env create",
			"validate --target ephemeral --fallback-target prod",
			"test --target ephemeral", "report", "env delete",
		},
		"deploy": {
			"plan", "validate --target prod", "gate --target prod", "deploy --target prod", "report",
		},
	}
	for name, steps := range want {
		if len(cfg.Pipelines[name]) != len(steps) {
			t.Fatalf("%s: %d steps, want %d", name, len(cfg.Pipelines[name]), len(steps))
		}
		for i, s := range steps {
			if cfg.Pipelines[name][i].Run != s {
				t.Errorf("%s step %d: %q, want %q", name, i, cfg.Pipelines[name][i].Run, s)
			}
		}
	}
	if cfg.Pipelines["pr"][7].When != "always" {
		t.Error("env delete does not run when always")
	}
}

func TestInitGitHubWorkflowGolden(t *testing.T) {
	got, err := render(githubWorkflowTemplate, initData{Provider: "salesforce"})
	if err != nil {
		t.Fatal(err)
	}
	checkGolden(t, "github-workflow.yml", got)
}

// B-003: the permit key is read in one place only, the step of the deploy job that runs in the
// protected environment.
func TestInitGitHubWorkflowKeepsThePermitKeyInTheEnvironmentJob(t *testing.T) {
	got, err := render(githubWorkflowTemplate, initData{})
	if err != nil {
		t.Fatal(err)
	}
	var wf struct {
		Jobs map[string]struct {
			Environment string                   `yaml:"environment"`
			Steps       []map[string]interface{} `yaml:"steps"`
		} `yaml:"jobs"`
	}
	if err := yaml.Unmarshal(got, &wf); err != nil {
		t.Fatalf("workflow is not YAML: %v", err)
	}
	if wf.Jobs["deploy"].Environment != "glassops-deploy" {
		t.Fatalf("deploy job environment %q", wf.Jobs["deploy"].Environment)
	}
	if wf.Jobs["pr"].Environment != "" {
		t.Fatalf("pr job runs in environment %q", wf.Jobs["pr"].Environment)
	}
	for name, job := range wf.Jobs {
		out, err := yaml.Marshal(job.Steps)
		if err != nil {
			t.Fatal(err)
		}
		has := strings.Contains(string(out), "GLASSOPS_PERMIT_KEY")
		if has && name != "deploy" {
			t.Errorf("job %s references the permit key", name)
		}
		if !has && name == "deploy" {
			t.Error("the deploy job does not pass the permit key")
		}
	}
	if n := strings.Count(string(got), "secrets.GLASSOPS_PERMIT_KEY"); n != 1 {
		t.Errorf("the permit key secret is referenced %d times, want 1", n)
	}
}
