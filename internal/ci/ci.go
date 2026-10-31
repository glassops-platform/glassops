// Package ci reads the CI context a command runs in and writes outputs back to the CI.
// Each CI system has one file. Detection is by environment variables.
package ci

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	"github.com/glassops-platform/glassops/internal/core/facts"
)

// Context describes where and why a command is running.
type Context struct {
	CI       string // github, local
	Repo     string // owner/name when known
	Ref      string // branch or tag ref
	SHA      string
	Event    string // push, pull_request, ... or "local"
	Actor    string
	PRNumber string
	BaseRef  string // pull request base branch, empty outside a pull request
	RunID    string
	// TriggeringActor is who started this run attempt. On a re-run it can differ from Actor.
	TriggeringActor string
	// PRAuthor, Labels and Approvals are approval evidence. Context only carries them. A
	// layer that reads them from the CI fills them in, and until then they stay empty.
	PRAuthor  string
	Labels    []string
	Approvals []facts.Approval
	// Token is the CI's API token. It must never reach a result, a log or a fact. Do not
	// print it, and JSON encoding leaves it out.
	Token string `json:"-"`
}

// Level is the severity of an annotation.
type Level string

const (
	Notice  Level = "notice"
	Warning Level = "warning"
	Errors  Level = "error"
)

// CI is one CI system's view of context and outputs.
type CI interface {
	Name() string
	Context() Context
	// SetOutput publishes a named output for later steps.
	SetOutput(name, value string) error
	// Annotate reports a message the CI can surface next to the run.
	Annotate(level Level, msg string)
	// Summary appends markdown to the run summary.
	Summary(markdown string) error
}

// Env reads environment variables. It is a function so tests can supply their own.
type Env func(string) string

// Detect picks the CI from the environment. name forces one: "github" or "local".
// An empty name or "auto" detects.
func Detect(name string, env Env, out io.Writer, dir string) (CI, error) {
	if env == nil {
		env = os.Getenv
	}
	switch strings.ToLower(name) {
	case "", "auto":
		if env("GITHUB_ACTIONS") == "true" {
			return newGitHub(env, out), nil
		}
		return newLocal(out, dir), nil
	case "github":
		return newGitHub(env, out), nil
	case "local":
		return newLocal(out, dir), nil
	default:
		return nil, fmt.Errorf("unknown CI %q (want auto, github or local)", name)
	}
}

// git runs a git command in dir and returns trimmed stdout, or "" on any error.
func git(dir string, args ...string) string {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	out, err := cmd.Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}
