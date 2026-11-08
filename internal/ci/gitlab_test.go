package ci

import (
	"bytes"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

// The variables below are from docs, not captured: GitLab's "Predefined CI/CD variables"
// reference. No GitLab runner has been observed, so everything here is [unverified] against a
// real pipeline. The values are synthetic.

var gitlabPush = map[string]string{
	"GITLAB_CI":          "true",
	"CI_PIPELINE_SOURCE": "push",
	"CI_PROJECT_PATH":    "acme/widgets",
	"CI_COMMIT_SHA":      "0123456789abcdef0123456789abcdef01234567",
	"CI_COMMIT_REF_NAME": "main",
	"CI_COMMIT_BRANCH":   "main",
	"CI_PIPELINE_ID":     "9001",
	"GITLAB_USER_LOGIN":  "dev1",
	// Never read: neither may reach the context.
	"GITLAB_USER_EMAIL": "dev1@example.invalid",
	"CI_JOB_TOKEN":      "job-token-value",
}

// Merge request pipelines do not set CI_COMMIT_BRANCH.
var gitlabMR = map[string]string{
	"GITLAB_CI":                           "true",
	"CI_PIPELINE_SOURCE":                  "merge_request_event",
	"CI_PROJECT_PATH":                     "acme/widgets",
	"CI_COMMIT_SHA":                       "fedcba9876543210fedcba9876543210fedcba98",
	"CI_COMMIT_REF_NAME":                  "feature/x",
	"CI_MERGE_REQUEST_IID":                "12",
	"CI_MERGE_REQUEST_SOURCE_BRANCH_NAME": "feature/x",
	"CI_MERGE_REQUEST_TARGET_BRANCH_NAME": "main",
	"CI_MERGE_REQUEST_EVENT_TYPE":         "detached",
	"CI_PIPELINE_ID":                      "9002",
	"GITLAB_USER_LOGIN":                   "dev2",
}

var gitlabTag = map[string]string{
	"GITLAB_CI":          "true",
	"CI_PIPELINE_SOURCE": "push",
	"CI_PROJECT_PATH":    "acme/widgets",
	"CI_COMMIT_SHA":      "0123456789abcdef0123456789abcdef01234567",
	"CI_COMMIT_REF_NAME": "v1.2.3",
	"CI_COMMIT_TAG":      "v1.2.3",
	"CI_PIPELINE_ID":     "9003",
}

func TestGitLabContext(t *testing.T) {
	cases := []struct {
		name string
		env  map[string]string
		want Context
	}{
		{"push", gitlabPush, Context{
			CI: "gitlab", Repo: "acme/widgets", Ref: "refs/heads/main", SHA: "0123456789abcdef0123456789abcdef01234567",
			Event: "push", Actor: "dev1", TriggeringActor: "dev1", RunID: "9001",
		}},
		{"merge request", gitlabMR, Context{
			CI: "gitlab", Repo: "acme/widgets", Ref: "refs/heads/feature/x", SHA: "fedcba9876543210fedcba9876543210fedcba98",
			Event: "pull_request", Actor: "dev2", TriggeringActor: "dev2", RunID: "9002", PRNumber: "12", BaseRef: "main",
		}},
		{"tag", gitlabTag, Context{
			CI: "gitlab", Repo: "acme/widgets", Ref: "refs/tags/v1.2.3", SHA: "0123456789abcdef0123456789abcdef01234567",
			Event: "push", RunID: "9003",
		}},
		{"scheduled passes through", map[string]string{"CI_PIPELINE_SOURCE": "schedule", "CI_COMMIT_BRANCH": "main"}, Context{
			CI: "gitlab", Ref: "refs/heads/main", Event: "schedule",
		}},
		{"no branch variable falls back to the ref name", map[string]string{"CI_COMMIT_REF_NAME": "dev"}, Context{
			CI: "gitlab", Ref: "refs/heads/dev", Event: "push",
		}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := newGitLab(mapEnv(tc.env), &bytes.Buffer{}, "").Context()
			if !reflect.DeepEqual(got, tc.want) {
				t.Fatalf("context\n got %+v\nwant %+v", got, tc.want)
			}
		})
	}
}

func TestGitLabOutputAndSummaryFiles(t *testing.T) {
	dir := t.TempDir()
	g := newGitLab(mapEnv(gitlabPush), &bytes.Buffer{}, dir)
	if err := g.SetOutput("status", "pass"); err != nil {
		t.Fatal(err)
	}
	if err := g.SetOutput("env_id", "00DXX0000000001"); err != nil {
		t.Fatal(err)
	}
	if err := g.Summary("# gate"); err != nil {
		t.Fatal(err)
	}
	read := func(rel string) string {
		raw, err := os.ReadFile(filepath.Join(dir, filepath.FromSlash(rel)))
		if err != nil {
			t.Fatal(err)
		}
		return lfOnly(raw)
	}
	if got := read(DefaultOutputFile); got != "status=pass\nenv_id=00DXX0000000001\n" {
		t.Fatalf("outputs = %q", got)
	}
	if got := read(DefaultSummaryFile); got != "# gate\n" {
		t.Fatalf("summary = %q", got)
	}
	if err := g.SetOutput("1bad", "x"); err == nil {
		t.Fatal("a name starting with a digit was accepted")
	}
	if err := g.SetOutput("multi", "a\r\nb"); err == nil {
		t.Fatal("a multi-line value was accepted")
	}
}

func TestGitLabOutputFileOverride(t *testing.T) {
	dir := t.TempDir()
	g := newGitLab(mapEnv(map[string]string{OutputFileEnv: "custom/o.env"}), &bytes.Buffer{}, dir)
	if err := g.SetOutput("a", "1"); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(dir, "custom", "o.env")); err != nil {
		t.Fatal(err)
	}
}

func TestGitLabAnnotateIsAPlainLine(t *testing.T) {
	var out bytes.Buffer
	newGitLab(mapEnv(nil), &out, "").Annotate(Warning, "careful")
	if out.String() != "warning: careful\n" {
		t.Fatalf("annotate = %q", out.String())
	}
}
