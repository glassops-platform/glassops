package cli

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/glassops-platform/glassops/internal/ci"
	"github.com/glassops-platform/glassops/internal/core/exitcode"
	"github.com/glassops-platform/glassops/internal/core/report"
	"github.com/glassops-platform/glassops/internal/core/sarif"
)

type reportOptions struct {
	sarifPath   string
	summaryPath string
	prComment   bool
}

func newReportCommand(app *App, g *globals) *cobra.Command {
	o := &reportOptions{}
	cmd := &cobra.Command{
		Use:   "report",
		Short: "Summarise every result for the CI and write SARIF",
		Long: "report reads every result in .glassops/results/, writes a markdown summary to\n" +
			"the CI's run summary, and writes SARIF for policy denials and validate failures.\n" +
			"With --summary the markdown also goes to a file. SARIF files already in the results\n" +
			"directory, such as Code Analyzer's, are merged into the SARIF output.\n" +
			"report exits with the worst status among the results, so a CI can run it after a\n" +
			"failed step and still fail the job. It exits 2 when there are no results or one\n" +
			"cannot be read.\n" +
			"With --pr-comment, or report.pr_comment in the config, a pull request run also posts the\n" +
			"summary as one comment, created the first time and updated after. A failed post warns\n" +
			"and never changes the exit code.",
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return app.report(g, o)
		},
	}
	f := cmd.Flags()
	f.StringVar(&o.sarifPath, "sarif", "", "SARIF file to write (default <dir>/.glassops/report.sarif)")
	f.StringVar(&o.summaryPath, "summary", "", "also write the markdown summary to this file")
	f.BoolVar(&o.prComment, "pr-comment", false, "post the summary as a pull request comment (GitHub)")
	return cmd
}

// report builds the summary and the SARIF log from the results directory.
func (app *App) report(g *globals, o *reportOptions) error {
	fail := func(err error) error { return &exitcode.ExitError{Code: exitcode.Error, Err: err} }

	c, err := app.detectCI(g)
	if err != nil {
		return fail(err)
	}
	sarifPath := o.sarifPath
	if sarifPath == "" {
		sarifPath = filepath.Join(g.dir, ".glassops", "report.sarif")
	}
	b, err := report.LoadSkipping(g.results(), []string{sarifPath},
		filepath.Join(g.dir, ".glassops", "permit.json"))
	if err != nil {
		if errors.Is(err, report.ErrNoResults) {
			return fail(fmt.Errorf("no results in %s, so there is nothing to report", g.results()))
		}
		return fail(err)
	}

	md := report.Summary(b)
	if err := c.Summary(md); err != nil {
		return fail(err)
	}
	if o.summaryPath != "" {
		if err := writeReportFile(o.summaryPath, []byte(md)); err != nil {
			return fail(fmt.Errorf("write summary: %w", err))
		}
	}
	if o.prComment || app.peekReportPRComment(g) {
		app.postPRComment(c, b)
	}
	data, err := sarif.Marshal(report.SARIF(b, app.Version))
	if err != nil {
		return fail(err)
	}
	if err := writeReportFile(sarifPath, data); err != nil {
		return fail(fmt.Errorf("write sarif: %w", err))
	}
	fmt.Fprintf(app.Stdout, "report: %s (sarif: %s)\n", b.Worst(), sarifPath)

	if code := b.ExitCode(); code != exitcode.OK {
		return &exitcode.ExitError{Code: code}
	}
	return nil
}

// writeReportFile writes data to path, creating its directory.
func writeReportFile(path string, data []byte) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}

// peekReportPRComment reports whether report.pr_comment is set. A config that does not load
// means false, and the commands that need the config report that error themselves.
func (app *App) peekReportPRComment(g *globals) bool {
	cfg := app.peekConfig(g)
	return cfg != nil && cfg.Report.PRComment
}

// postPRComment creates or updates the report comment on the pull request this run belongs to.
// It does nothing outside a GitHub pull_request run, and a failure is a warning only, because
// the comment is a convenience and the job's result must not depend on it.
func (app *App) postPRComment(c ci.CI, b *report.Bundle) {
	cx := c.Context()
	if c.Name() != "github" || cx.Event != "pull_request" || cx.PRNumber == "" {
		c.Annotate(ci.Notice, "report: no pull request comment, because this is not a GitHub pull_request run")
		return
	}
	n, err := strconv.Atoi(cx.PRNumber)
	if err != nil || cx.Repo == "" || cx.Token == "" {
		c.Annotate(ci.Warning, "report: no pull request comment, because the repository, pull request number or token is missing")
		return
	}
	api := ci.NewGitHubAPI(app.Getenv("GITHUB_API_URL"), cx.Token, nil)
	if _, err := api.UpsertComment(context.Background(), cx.Repo, n, report.CommentMarker, report.Comment(b)); err != nil {
		msg := "report: the pull request comment was not posted: " + err.Error()
		var he *ci.HTTPError
		if errors.As(err, &he) && he.Status == 403 {
			msg += " (the job needs pull-requests: write, and a pull request from a fork gets a read-only token)"
		}
		c.Annotate(ci.Warning, msg)
	}
}
