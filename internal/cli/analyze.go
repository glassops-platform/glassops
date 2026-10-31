package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/glassops-platform/glassops/internal/core/exitcode"
	"github.com/glassops-platform/glassops/internal/core/run"
)

type analyzeOptions struct {
	ruleSelector string
	sarifPath    string
}

func newAnalyzeCommand(app *App, g *globals) *cobra.Command {
	o := &analyzeOptions{}
	cmd := &cobra.Command{
		Use:   "analyze",
		Short: "Run static analysis over the source",
		Long: "analyze runs the Salesforce Code Analyzer over the project and records the\n" +
			"violations as facts analyze.total and analyze.sev1 to analyze.sev5, and keeps the\n" +
			"SARIF file as an artifact. It exits 0 when the analyzer ran, whatever it found,\n" +
			"and 2 when it could not run. Deciding what is too severe is the job of gate\n" +
			"(analyzer_max_severity), not of this command.",
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return app.analyze(g, o)
		},
	}
	f := cmd.Flags()
	f.StringVar(&o.ruleSelector, "rule-selector", "", "rules to run (default analyze.rule_selector, else Recommended)")
	f.StringVar(&o.sarifPath, "sarif", "", "SARIF file to write (default <results-dir>/analyze.sarif)")
	return cmd
}

// violationCounts is the part of the Code Analyzer JSON output analyze reads.
type violationCounts struct {
	Total int `json:"total"`
	Sev1  int `json:"sev1"`
	Sev2  int `json:"sev2"`
	Sev3  int `json:"sev3"`
	Sev4  int `json:"sev4"`
	Sev5  int `json:"sev5"`
}

// parseViolationCounts reads violationCounts from the analyzer's JSON output. A file without
// the object is an error, because an empty file must never read as a clean run.
func parseViolationCounts(data []byte) (violationCounts, error) {
	var doc struct {
		Counts *violationCounts `json:"violationCounts"`
	}
	if err := json.Unmarshal(data, &doc); err != nil {
		return violationCounts{}, fmt.Errorf("parse analyzer output: %w", err)
	}
	if doc.Counts == nil {
		return violationCounts{}, errors.New("analyzer output has no violationCounts")
	}
	return *doc.Counts, nil
}

func (app *App) analyze(g *globals, o *analyzeOptions) error {
	c, err := app.detectCI(g)
	if err != nil {
		return &exitcode.ExitError{Code: exitcode.Error, Err: err}
	}
	proj, err := app.loadProject(g)
	if err != nil {
		return err
	}
	res := run.New("analyze", app.Now)
	res.Provider = proj.provider.Name()

	if e := proj.cfg.Analyze.Enabled; e != nil && !*e {
		res.Status = run.StatusEmpty
		res.AddCheck("analyze", run.StatusSkipped, "analyze.enabled is false")
		return app.finish(res, c, g)
	}
	if proj.provider.Name() != "salesforce" {
		res.AddCheck("analyze", run.StatusError, fmt.Sprintf("analyze supports the salesforce provider only, not %s", proj.provider.Name()))
		return app.finish(res, c, g)
	}

	selector := o.ruleSelector
	if selector == "" {
		selector = proj.cfg.Analyze.Selector()
	}
	results, err := filepath.Abs(g.results())
	if err != nil {
		res.AddCheck("analyze", run.StatusError, err.Error())
		return app.finish(res, c, g)
	}
	// The analyzer's own JSON cannot sit in the results dir as analyze.json, because the
	// step result is written there under that name. It is kept beside the results instead.
	rawPath := filepath.Join(filepath.Dir(results), "analyze", "code-analyzer.json")
	sarifPath := o.sarifPath
	if sarifPath == "" {
		sarifPath = filepath.Join(results, "analyze.sarif")
	}
	if sarifPath, err = filepath.Abs(sarifPath); err != nil {
		res.AddCheck("analyze", run.StatusError, err.Error())
		return app.finish(res, c, g)
	}
	for _, p := range []string{rawPath, sarifPath} {
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			res.AddCheck("analyze", run.StatusError, fmt.Sprintf("create output dir: %v", err))
			return app.finish(res, c, g)
		}
		// A stale file from an earlier run must not be read as this run's output.
		if err := os.Remove(p); err != nil && !errors.Is(err, os.ErrNotExist) {
			res.AddCheck("analyze", run.StatusError, fmt.Sprintf("remove stale output: %v", err))
			return app.finish(res, c, g)
		}
	}

	// No --severity-threshold: that only changes sf's exit code. gate decides.
	_, err = app.Run(g.dir, "sf", "code-analyzer", "run",
		"--workspace", ".",
		"--rule-selector", selector,
		"--output-file", rawPath,
		"--output-file", sarifPath)
	if err != nil {
		res.AddCheck("analyze", run.StatusError, fmt.Sprintf("sf code-analyzer run failed: %v", err))
		return app.finish(res, c, g)
	}
	data, err := os.ReadFile(rawPath)
	if err != nil {
		res.AddCheck("analyze", run.StatusError, fmt.Sprintf("analyzer wrote no JSON output: %v", err))
		return app.finish(res, c, g)
	}
	counts, err := parseViolationCounts(data)
	if err != nil {
		res.AddCheck("analyze", run.StatusError, err.Error())
		return app.finish(res, c, g)
	}
	if _, err := os.Stat(sarifPath); err != nil {
		res.AddCheck("analyze", run.StatusError, fmt.Sprintf("analyzer wrote no SARIF output: %v", err))
		return app.finish(res, c, g)
	}

	res.Facts["analyze.rule_selector"] = selector
	res.Facts["analyze.total"] = counts.Total
	res.Facts["analyze.sev1"] = counts.Sev1
	res.Facts["analyze.sev2"] = counts.Sev2
	res.Facts["analyze.sev3"] = counts.Sev3
	res.Facts["analyze.sev4"] = counts.Sev4
	res.Facts["analyze.sev5"] = counts.Sev5
	res.Artifacts = append(res.Artifacts, relativeTo(g.dir, sarifPath), relativeTo(g.dir, rawPath))
	res.AddCheck("analyze", run.StatusPass, fmt.Sprintf("%d violations (sev1 %d, sev2 %d, sev3 %d, sev4 %d, sev5 %d)",
		counts.Total, counts.Sev1, counts.Sev2, counts.Sev3, counts.Sev4, counts.Sev5))
	return app.finish(res, c, g)
}

// relativeTo returns path relative to dir when it lies inside it, else path unchanged, so an
// artifact path in a result means the same thing outside the container that wrote it.
func relativeTo(dir, path string) string {
	abs, err := filepath.Abs(dir)
	if err != nil {
		return path
	}
	rel, err := filepath.Rel(abs, path)
	if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return path
	}
	return filepath.ToSlash(rel)
}
