package ci

import (
	"errors"
	"fmt"
	"io"
)

// errBitbucketNotImplemented is what the Bitbucket layer returns for outputs and the summary
// until it lands.
var errBitbucketNotImplemented = errors.New("the Bitbucket Pipelines layer is not implemented yet")

// Bitbucket is the Bitbucket Pipelines layer. It is not implemented yet: Context names the
// CI only, and outputs and the summary return an error.
type Bitbucket struct {
	env Env
	out io.Writer
	dir string
}

func newBitbucket(env Env, out io.Writer, dir string) *Bitbucket {
	return &Bitbucket{env: env, out: out, dir: dir}
}

func (b *Bitbucket) Name() string { return "bitbucket" }

func (b *Bitbucket) Context() Context { return Context{CI: "bitbucket"} }

func (b *Bitbucket) SetOutput(name, value string) error { return errBitbucketNotImplemented }

// Annotate prints a plain log line.
func (b *Bitbucket) Annotate(level Level, msg string) {
	fmt.Fprintf(b.out, "%s: %s\n", level, msg)
}

func (b *Bitbucket) Summary(markdown string) error { return errBitbucketNotImplemented }
