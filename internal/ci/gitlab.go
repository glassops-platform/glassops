package ci

import (
	"errors"
	"fmt"
	"io"
)

// errGitLabNotImplemented is what the GitLab layer returns for outputs and the summary until
// it lands.
var errGitLabNotImplemented = errors.New("the GitLab CI layer is not implemented yet")

// GitLab is the GitLab CI layer. It is not implemented yet: Context names the CI only, and
// outputs and the summary return an error.
type GitLab struct {
	env Env
	out io.Writer
	dir string
}

func newGitLab(env Env, out io.Writer, dir string) *GitLab {
	return &GitLab{env: env, out: out, dir: dir}
}

func (g *GitLab) Name() string { return "gitlab" }

func (g *GitLab) Context() Context { return Context{CI: "gitlab"} }

func (g *GitLab) SetOutput(name, value string) error { return errGitLabNotImplemented }

// Annotate prints a plain log line.
func (g *GitLab) Annotate(level Level, msg string) {
	fmt.Fprintf(g.out, "%s: %s\n", level, msg)
}

func (g *GitLab) Summary(markdown string) error { return errGitLabNotImplemented }
