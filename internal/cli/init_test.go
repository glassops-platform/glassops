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
	"github.com/glassops-platform/glassops/internal/core/exitcode"
	"github.com/glassops-platform/glassops/internal/core/run"
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

func TestInitBitbucketPipelinesGolden(t *testing.T) {
	got, err := render(bitbucketPipelinesTemplate, initData{Provider: "salesforce"})
	if err != nil {
		t.Fatal(err)
	}
	checkGolden(t, "bitbucket-pipelines.yml", got)
}

func TestInitGitLabPipelinesGolden(t *testing.T) {
	got, err := render(gitlabPipelinesTemplate, initData{Provider: "salesforce"})
	if err != nil {
		t.Fatal(err)
	}
	checkGolden(t, "gitlab-ci.yml", got)
}

// The sample pipeline files in examples are the rendered templates, and the sample configs run
// the default pipelines, so neither can drift from what init writes.
func TestOtherCIExamplesMatchInit(t *testing.T) {
	cases := []struct{ dir, file, tmpl string }{
		{"bitbucket", "bitbucket-pipelines.yml", bitbucketPipelinesTemplate},
		{"gitlab", ".gitlab-ci.yml", gitlabPipelinesTemplate},
	}
	for _, tc := range cases {
		t.Run(tc.dir, func(t *testing.T) {
			base := filepath.Join("..", "..", "examples", tc.dir)
			want, err := render(tc.tmpl, initData{Provider: "salesforce"})
			if err != nil {
				t.Fatal(err)
			}
			got, err := os.ReadFile(filepath.Join(base, tc.file))
			if err != nil {
				t.Fatal(err)
			}
			if lf(got) != lf(want) {
				t.Errorf("examples/%s/%s differs from the init template", tc.dir, tc.file)
			}
			data, err := os.ReadFile(filepath.Join(base, "glassops.yaml"))
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
		})
	}
}

// Both files are YAML. The pipelines run the same commands as the GitHub workflow, clone the
// full history, and name the permit key in comments only: the key is a CI-side variable the
// runner injects, so no script line or variables block carries a value for it.
func TestInitOtherPipelinesAreSound(t *testing.T) {
	bb, err := render(bitbucketPipelinesTemplate, initData{})
	if err != nil {
		t.Fatal(err)
	}
	var bbDoc struct {
		Clone     struct{ Depth string } `yaml:"clone"`
		Pipelines struct {
			PullRequests map[string][]struct {
				Step struct {
					Script     []string `yaml:"script"`
					Deployment string   `yaml:"deployment"`
				} `yaml:"step"`
			} `yaml:"pull-requests"`
			Branches map[string][]struct {
				Step struct {
					Script     []string `yaml:"script"`
					Deployment string   `yaml:"deployment"`
				} `yaml:"step"`
			} `yaml:"branches"`
		} `yaml:"pipelines"`
	}
	if err := yaml.Unmarshal(bb, &bbDoc); err != nil {
		t.Fatalf("bitbucket pipelines are not YAML: %v", err)
	}
	if bbDoc.Clone.Depth != "full" {
		t.Errorf("bitbucket clone depth %q, want full", bbDoc.Clone.Depth)
	}
	pr := bbDoc.Pipelines.PullRequests["**"]
	dep := bbDoc.Pipelines.Branches["main"]
	if len(pr) != 1 || len(dep) != 1 {
		t.Fatalf("bitbucket steps: pr %d, main %d", len(pr), len(dep))
	}
	if !contains(pr[0].Step.Script, "glassops run pr") || pr[0].Step.Deployment != "" {
		t.Errorf("bitbucket pr step: %+v", pr[0].Step)
	}
	if !contains(dep[0].Step.Script, "glassops run deploy") || dep[0].Step.Deployment != "production" {
		t.Errorf("bitbucket deploy step: %+v", dep[0].Step)
	}
	if strings.Contains(strings.ReplaceAll(string(bb), "\n#", ""), "GLASSOPS_PERMIT_KEY:") {
		t.Error("bitbucket pipeline sets the permit key")
	}

	gl, err := render(gitlabPipelinesTemplate, initData{})
	if err != nil {
		t.Fatal(err)
	}
	var glDoc map[string]interface{}
	if err := yaml.Unmarshal(gl, &glDoc); err != nil {
		t.Fatalf("gitlab pipelines are not YAML: %v", err)
	}
	vars, _ := glDoc["variables"].(map[string]interface{})
	if vars["GIT_DEPTH"] != "0" {
		t.Errorf("gitlab GIT_DEPTH = %v, want 0", vars["GIT_DEPTH"])
	}
	for job, want := range map[string]string{"glassops-pr": "glassops run pr", "glassops-deploy": "glassops run deploy"} {
		j, _ := glDoc[job].(map[string]interface{})
		var script []string
		for _, s := range j["script"].([]interface{}) {
			script = append(script, s.(string))
		}
		if !contains(script, want) {
			t.Errorf("gitlab job %s script %v lacks %q", job, script, want)
		}
	}
	if strings.Contains(strings.ReplaceAll(string(gl), "\n#", ""), "GLASSOPS_PERMIT_KEY:") {
		t.Error("gitlab pipeline sets the permit key")
	}
}

// init under the Bitbucket and GitLab layers writes the CI file for that CI and exits 0. Before
// the layers existed, every command on those CIs exited 2 because outputs errored.
func TestInitOnBitbucketAndGitLab(t *testing.T) {
	cases := []struct{ ci, file, check string }{
		{"bitbucket", "bitbucket-pipelines.yml", "ci.bitbucket"},
		{"gitlab", ".gitlab-ci.yml", "ci.gitlab"},
	}
	for _, tc := range cases {
		t.Run(tc.ci, func(t *testing.T) {
			h := newHarness(t)
			h.write(t, "sfdx-project.json", "{}")
			if code := Execute(h.app, []string{"init", "--dir", h.dir, "--ci", tc.ci}); code != exitcode.OK {
				t.Fatalf("exit %d, output:\n%s", code, h.out)
			}
			if _, err := os.Stat(filepath.Join(h.dir, tc.file)); err != nil {
				t.Fatalf("%s not written: %v", tc.file, err)
			}
			if got := checkStatus(h.result(t, "init"), tc.check); got != run.StatusPass {
				t.Fatalf("check %s = %q", tc.check, got)
			}
			// The layer wrote its outputs file for the command's own outputs.
			if _, err := os.Stat(filepath.Join(h.dir, ".glassops", "outputs.env")); err != nil {
				t.Fatalf("outputs file: %v", err)
			}
		})
	}
}
