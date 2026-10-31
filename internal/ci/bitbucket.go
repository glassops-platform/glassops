package ci

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// Output and summary files for the CIs that have no output channel of their own. A command
// appends to these files and the pipeline publishes them as artifacts. GitLab reads the
// outputs file as a dotenv report.
const (
	// OutputFileEnv names the variable that moves the outputs file. A relative path is
	// relative to the project directory.
	OutputFileEnv = "GLASSOPS_OUTPUT_FILE"
	// SummaryFileEnv names the variable that moves the summary file.
	SummaryFileEnv = "GLASSOPS_SUMMARY_FILE"
	// DefaultOutputFile is where outputs go when OutputFileEnv is not set.
	DefaultOutputFile = ".glassops/outputs.env"
	// DefaultSummaryFile is where the summary goes when SummaryFileEnv is not set.
	DefaultSummaryFile = ".glassops/summary.md"
)

// outputName is the form a dotenv key can take.
var outputName = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*$`)

// fileTarget resolves a file the layer writes: the variable's value, or the default, made
// absolute against dir when it is relative.
func fileTarget(env Env, variable, def, dir string) string {
	p := env(variable)
	if p == "" {
		p = def
	}
	if !filepath.IsAbs(p) && dir != "" {
		p = filepath.Join(dir, filepath.FromSlash(p))
	}
	return p
}

// appendTo appends text and a newline to path, creating the directory and the file.
func appendTo(path, text string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	if _, err := fmt.Fprintln(f, text); err != nil {
		f.Close()
		return err
	}
	return f.Close()
}

// writeOutput appends NAME=value to the outputs file. A name that is not a valid dotenv key
// and a multi-line value are errors, because the file could not carry them faithfully.
func writeOutput(path, name, value string) error {
	if !outputName.MatchString(name) {
		return fmt.Errorf("output %q: the name must be letters, digits and underscores", name)
	}
	if strings.ContainsAny(value, "\r\n") {
		return fmt.Errorf("output %s: multi-line values are not supported", name)
	}
	return appendTo(path, name+"="+value)
}

// Bitbucket is the Bitbucket Pipelines layer. It reads the BITBUCKET_* predefined variables,
// appends outputs to a dotenv file and the summary to a markdown file. Everything it assumes
// about a real runner comes from Atlassian's documentation and is [unverified] until a proof
// pipeline runs.
type Bitbucket struct {
	env Env
	out io.Writer
	dir string
}

func newBitbucket(env Env, out io.Writer, dir string) *Bitbucket {
	return &Bitbucket{env: env, out: out, dir: dir}
}

func (b *Bitbucket) Name() string { return "bitbucket" }

// Context maps the predefined variables. Bitbucket exposes no user name, only the UUID of
// whoever started the build, so Actor and TriggeringActor hold that UUID. BITBUCKET_PR_ID is
// set in pull request pipelines only, and those become Event "pull_request". Every other
// pipeline is "push", because a scheduled or custom pipeline cannot be told apart by a
// variable. BITBUCKET_BRANCH is the source branch in a pull request pipeline. Approval
// evidence (PRAuthor, Labels, Approvals) is not read from Bitbucket in v1.
func (b *Bitbucket) Context() Context {
	c := Context{
		CI:              "bitbucket",
		Repo:            b.env("BITBUCKET_REPO_FULL_NAME"),
		SHA:             b.env("BITBUCKET_COMMIT"),
		Event:           "push",
		Actor:           b.env("BITBUCKET_STEP_TRIGGERER_UUID"),
		TriggeringActor: b.env("BITBUCKET_STEP_TRIGGERER_UUID"),
		RunID:           b.env("BITBUCKET_BUILD_NUMBER"),
	}
	switch {
	case b.env("BITBUCKET_TAG") != "":
		c.Ref = "refs/tags/" + b.env("BITBUCKET_TAG")
	case b.env("BITBUCKET_BRANCH") != "":
		c.Ref = "refs/heads/" + b.env("BITBUCKET_BRANCH")
	}
	if id := b.env("BITBUCKET_PR_ID"); id != "" {
		c.Event = "pull_request"
		c.PRNumber = id
		c.BaseRef = b.env("BITBUCKET_PR_DESTINATION_BRANCH")
	}
	return c
}

// SetOutput appends NAME=value to GLASSOPS_OUTPUT_FILE, or .glassops/outputs.env.
func (b *Bitbucket) SetOutput(name, value string) error {
	return writeOutput(fileTarget(b.env, OutputFileEnv, DefaultOutputFile, b.dir), name, value)
}

// Annotate prints a plain log line.
func (b *Bitbucket) Annotate(level Level, msg string) {
	fmt.Fprintf(b.out, "%s: %s\n", level, msg)
}

// Summary appends markdown to GLASSOPS_SUMMARY_FILE, or .glassops/summary.md.
func (b *Bitbucket) Summary(markdown string) error {
	return appendTo(fileTarget(b.env, SummaryFileEnv, DefaultSummaryFile, b.dir), markdown)
}
