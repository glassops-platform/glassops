package ci

import (
	"fmt"
	"io"
	"os"
	"strings"
)

// GitHub reads GITHUB_* variables and writes to GITHUB_OUTPUT and GITHUB_STEP_SUMMARY.
type GitHub struct {
	env Env
	out io.Writer
}

func newGitHub(env Env, out io.Writer) *GitHub {
	return &GitHub{env: env, out: out}
}

func (g *GitHub) Name() string { return "github" }

func (g *GitHub) Context() Context {
	c := Context{
		CI:    "github",
		Repo:  g.env("GITHUB_REPOSITORY"),
		Ref:   g.env("GITHUB_REF"),
		SHA:   g.env("GITHUB_SHA"),
		Event: g.env("GITHUB_EVENT_NAME"),
		Actor: g.env("GITHUB_ACTOR"),
		RunID: g.env("GITHUB_RUN_ID"),
		// GITHUB_BASE_REF is set only on pull request events.
		BaseRef:         g.env("GITHUB_BASE_REF"),
		TriggeringActor: g.env("GITHUB_TRIGGERING_ACTOR"),
		Token:           firstSet(g.env("GITHUB_TOKEN"), g.env("GH_TOKEN")),
	}
	// Pull request refs look like refs/pull/<number>/merge.
	if parts := strings.Split(c.Ref, "/"); len(parts) == 4 && parts[1] == "pull" {
		c.PRNumber = parts[2]
	}
	return c
}

func (g *GitHub) SetOutput(name, value string) error {
	if strings.ContainsAny(value, "\r\n") {
		return fmt.Errorf("output %s: multi-line values are not supported", name)
	}
	return appendLine(g.env("GITHUB_OUTPUT"), name+"="+value)
}

func (g *GitHub) Annotate(level Level, msg string) {
	fmt.Fprintf(g.out, "::%s::%s\n", level, escapeData(msg))
}

func (g *GitHub) Summary(markdown string) error {
	return appendLine(g.env("GITHUB_STEP_SUMMARY"), markdown)
}

// escapeData applies GitHub's workflow command escaping for message data.
func escapeData(s string) string {
	s = strings.ReplaceAll(s, "%", "%25")
	s = strings.ReplaceAll(s, "\r", "%0D")
	return strings.ReplaceAll(s, "\n", "%0A")
}

func appendLine(path, line string) error {
	if path == "" {
		return nil
	}
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = fmt.Fprintln(f, line)
	return err
}

// firstSet returns the first non-empty value.
func firstSet(values ...string) string {
	for _, v := range values {
		if v != "" {
			return v
		}
	}
	return ""
}
