package ci

import (
	"fmt"
	"io"
	"os/user"
)

// Local reads git for context and prints outputs to the terminal.
type Local struct {
	out io.Writer
	dir string
}

func newLocal(out io.Writer, dir string) *Local {
	return &Local{out: out, dir: dir}
}

func (l *Local) Name() string { return "local" }

func (l *Local) Context() Context {
	c := Context{
		CI:    "local",
		Ref:   git(l.dir, "rev-parse", "--abbrev-ref", "HEAD"),
		SHA:   git(l.dir, "rev-parse", "HEAD"),
		Event: "local",
	}
	if u, err := user.Current(); err == nil {
		c.Actor = u.Username
	}
	return c
}

func (l *Local) SetOutput(name, value string) error {
	_, err := fmt.Fprintf(l.out, "output %s=%s\n", name, value)
	return err
}

func (l *Local) Annotate(level Level, msg string) {
	fmt.Fprintf(l.out, "%s: %s\n", level, msg)
}

func (l *Local) Summary(markdown string) error {
	_, err := fmt.Fprintln(l.out, markdown)
	return err
}
