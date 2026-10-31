package ci

import (
	"fmt"
	"io"
)

// GitLab is the GitLab CI layer. It reads the CI_* predefined variables, appends outputs to a
// dotenv file (which a job publishes with artifacts:reports:dotenv) and the summary to a
// markdown file. Everything it assumes about a real runner comes from GitLab's documentation
// and is [unverified] until a proof pipeline runs.
type GitLab struct {
	env Env
	out io.Writer
	dir string
}

func newGitLab(env Env, out io.Writer, dir string) *GitLab {
	return &GitLab{env: env, out: out, dir: dir}
}

func (g *GitLab) Name() string { return "gitlab" }

// Context maps the predefined variables. A merge request pipeline
// (CI_PIPELINE_SOURCE=merge_request_event) becomes Event "pull_request", and a push stays
// "push". Any other source (web, schedule, api, ...) is passed through unchanged. CI_COMMIT_BRANCH
// is unset in merge request and tag pipelines, so the merge request source branch and the tag
// are used there. GITLAB_USER_LOGIN is whoever started the pipeline. The user's email variable
// is never read. Approval evidence (PRAuthor, Labels, Approvals) is not read from GitLab in v1.
func (g *GitLab) Context() Context {
	c := Context{
		CI:              "gitlab",
		Repo:            g.env("CI_PROJECT_PATH"),
		SHA:             g.env("CI_COMMIT_SHA"),
		Actor:           g.env("GITLAB_USER_LOGIN"),
		TriggeringActor: g.env("GITLAB_USER_LOGIN"),
		RunID:           g.env("CI_PIPELINE_ID"),
	}
	switch source := g.env("CI_PIPELINE_SOURCE"); source {
	case "merge_request_event":
		c.Event = "pull_request"
	case "":
		c.Event = "push"
	default:
		c.Event = source
	}
	switch {
	case g.env("CI_COMMIT_TAG") != "":
		c.Ref = "refs/tags/" + g.env("CI_COMMIT_TAG")
	case g.env("CI_COMMIT_BRANCH") != "":
		c.Ref = "refs/heads/" + g.env("CI_COMMIT_BRANCH")
	case g.env("CI_MERGE_REQUEST_SOURCE_BRANCH_NAME") != "":
		c.Ref = "refs/heads/" + g.env("CI_MERGE_REQUEST_SOURCE_BRANCH_NAME")
	case g.env("CI_COMMIT_REF_NAME") != "":
		c.Ref = "refs/heads/" + g.env("CI_COMMIT_REF_NAME")
	}
	if iid := g.env("CI_MERGE_REQUEST_IID"); iid != "" {
		c.PRNumber = iid
		c.BaseRef = g.env("CI_MERGE_REQUEST_TARGET_BRANCH_NAME")
	}
	return c
}

// SetOutput appends NAME=value to GLASSOPS_OUTPUT_FILE, or .glassops/outputs.env.
func (g *GitLab) SetOutput(name, value string) error {
	return writeOutput(fileTarget(g.env, OutputFileEnv, DefaultOutputFile, g.dir), name, value)
}

// Annotate prints a plain log line.
func (g *GitLab) Annotate(level Level, msg string) {
	fmt.Fprintf(g.out, "%s: %s\n", level, msg)
}

// Summary appends markdown to GLASSOPS_SUMMARY_FILE, or .glassops/summary.md.
func (g *GitLab) Summary(markdown string) error {
	return appendTo(fileTarget(g.env, SummaryFileEnv, DefaultSummaryFile, g.dir), markdown)
}
