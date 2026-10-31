package ci

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func mapEnv(m map[string]string) Env {
	return func(k string) string { return m[k] }
}

func TestDetect(t *testing.T) {
	cases := []struct {
		name  string
		force string
		env   map[string]string
		want  string
	}{
		{"github by env", "", map[string]string{"GITHUB_ACTIONS": "true"}, "github"},
		{"local by default", "auto", map[string]string{}, "local"},
		{"forced local inside github", "local", map[string]string{"GITHUB_ACTIONS": "true"}, "local"},
		{"forced github", "github", map[string]string{}, "github"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			c, err := Detect(tc.force, mapEnv(tc.env), &bytes.Buffer{}, t.TempDir())
			if err != nil {
				t.Fatal(err)
			}
			if c.Name() != tc.want {
				t.Fatalf("Name() = %s, want %s", c.Name(), tc.want)
			}
		})
	}
	if _, err := Detect("jenkins", mapEnv(nil), &bytes.Buffer{}, ""); err == nil {
		t.Fatal("expected error for unknown CI")
	}
}

func TestGitHubContext(t *testing.T) {
	g := newGitHub(mapEnv(map[string]string{
		"GITHUB_REPOSITORY": "acme/app",
		"GITHUB_REF":        "refs/pull/42/merge",
		"GITHUB_SHA":        "abc123",
		"GITHUB_EVENT_NAME": "pull_request",
		"GITHUB_ACTOR":      "octocat",
		"GITHUB_RUN_ID":     "99",
	}), &bytes.Buffer{})
	c := g.Context()
	if c.Repo != "acme/app" || c.PRNumber != "42" || c.Event != "pull_request" || c.Actor != "octocat" || c.RunID != "99" {
		t.Fatalf("unexpected context: %+v", c)
	}
	push := newGitHub(mapEnv(map[string]string{"GITHUB_REF": "refs/heads/main"}), &bytes.Buffer{})
	if pr := push.Context().PRNumber; pr != "" {
		t.Fatalf("push PRNumber = %q, want empty", pr)
	}
}

func TestGitHubOutputsAndSummary(t *testing.T) {
	dir := t.TempDir()
	outFile := filepath.Join(dir, "output")
	sumFile := filepath.Join(dir, "summary")
	var stdout bytes.Buffer
	g := newGitHub(mapEnv(map[string]string{"GITHUB_OUTPUT": outFile, "GITHUB_STEP_SUMMARY": sumFile}), &stdout)

	if err := g.SetOutput("status", "pass"); err != nil {
		t.Fatal(err)
	}
	if err := g.SetOutput("bad", "a\nb"); err == nil {
		t.Fatal("expected error for multi-line output")
	}
	if err := g.Summary("## Doctor"); err != nil {
		t.Fatal(err)
	}
	g.Annotate(Errors, "50% broken\nnext")

	if got, _ := os.ReadFile(outFile); string(got) != "status=pass\n" {
		t.Fatalf("output file = %q", got)
	}
	if got, _ := os.ReadFile(sumFile); !strings.Contains(string(got), "## Doctor") {
		t.Fatalf("summary file = %q", got)
	}
	if got := stdout.String(); got != "::error::50%25 broken%0Anext\n" {
		t.Fatalf("annotation = %q", got)
	}
}

func TestLocalWritesToTerminal(t *testing.T) {
	var out bytes.Buffer
	l := newLocal(&out, t.TempDir())
	if err := l.SetOutput("status", "pass"); err != nil {
		t.Fatal(err)
	}
	l.Annotate(Warning, "careful")
	if got := out.String(); !strings.Contains(got, "status=pass") || !strings.Contains(got, "warning: careful") {
		t.Fatalf("local output = %q", got)
	}
	if c := l.Context(); c.CI != "local" || c.Event != "local" {
		t.Fatalf("unexpected context: %+v", c)
	}
}

func TestGitHubBaseRef(t *testing.T) {
	pr := newGitHub(mapEnv(map[string]string{"GITHUB_BASE_REF": "main", "GITHUB_EVENT_NAME": "pull_request"}), &bytes.Buffer{})
	if got := pr.Context().BaseRef; got != "main" {
		t.Fatalf("pull request BaseRef = %q, want main", got)
	}
	push := newGitHub(mapEnv(map[string]string{"GITHUB_REF": "refs/heads/main"}), &bytes.Buffer{})
	if got := push.Context().BaseRef; got != "" {
		t.Fatalf("push BaseRef = %q, want empty", got)
	}
	if got := newLocal(&bytes.Buffer{}, t.TempDir()).Context().BaseRef; got != "" {
		t.Fatalf("local BaseRef = %q, want empty", got)
	}
}

func TestGitHubTriggeringActorAndToken(t *testing.T) {
	g := newGitHub(mapEnv(map[string]string{
		"GITHUB_ACTOR":            "octocat",
		"GITHUB_TRIGGERING_ACTOR": "hubot",
		"GITHUB_TOKEN":            "ghs_example",
	}), &bytes.Buffer{})
	c := g.Context()
	if c.Actor != "octocat" || c.TriggeringActor != "hubot" || c.Token != "ghs_example" {
		t.Fatalf("actor %q, triggering actor %q, token set %v", c.Actor, c.TriggeringActor, c.Token != "")
	}
	if c.PRAuthor != "" || c.Labels != nil || c.Approvals != nil {
		t.Fatal("the GitHub layer filled approval evidence")
	}
	data, err := json.Marshal(c)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(data), "ghs_example") {
		t.Fatal("the token was encoded")
	}
	l := newLocal(&bytes.Buffer{}, t.TempDir()).Context()
	if l.Token != "" || l.TriggeringActor != "" || l.PRAuthor != "" {
		t.Fatal("local context filled a GitHub-only field")
	}
}
