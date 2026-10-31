package cli

import (
	"flag"
	"os"
	"path/filepath"
	"reflect"
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

// The pipelines the template writes are exactly config.DefaultPipelines(), so the two cannot
// drift apart.
func TestInitConfigPipelinesAreTheDefaults(t *testing.T) {
	got, err := render(configTemplate, initData{Provider: "salesforce"})
	if err != nil {
		t.Fatal(err)
	}
	cfg, err := config.Parse(got)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(cfg.Pipelines, config.DefaultPipelines()) {
		t.Errorf("template pipelines %+v, default pipelines %+v", cfg.Pipelines, config.DefaultPipelines())
	}
}

// The same holds for the example consumer repository.
func TestExampleConfigPipelinesAreTheDefaults(t *testing.T) {
	data, err := os.ReadFile(filepath.Join("..", "..", "examples", "consumer", "glassops.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	cfg, err := config.Parse(data)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(cfg.Pipelines, config.DefaultPipelines()) {
		t.Errorf("example pipelines %+v, default pipelines %+v", cfg.Pipelines, config.DefaultPipelines())
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
			Permissions map[string]string        `yaml:"permissions"`
			Steps       []map[string]interface{} `yaml:"steps"`
		} `yaml:"jobs"`
	}
	if err := yaml.Unmarshal(got, &wf); err != nil {
		t.Fatalf("workflow is not YAML: %v", err)
	}
	if wf.Jobs["deploy"].Environment != "glassops-deploy" {
		t.Fatalf("deploy job environment %q", wf.Jobs["deploy"].Environment)
	}
	wantPerms := map[string]string{"contents": "write", "actions": "read", "issues": "read", "pull-requests": "read"}
	if !reflect.DeepEqual(wf.Jobs["deploy"].Permissions, wantPerms) {
		t.Fatalf("deploy job permissions %v, want %v", wf.Jobs["deploy"].Permissions, wantPerms)
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
