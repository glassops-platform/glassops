// Package analyzer wraps the Salesforce Code Analyzer.
package analyzer

import (
	"encoding/json"
	"os/exec"
	"strings"

	"github.com/glassops-platform/glassops/packages/runtime/internal/gha"
)

// Result contains analysis results.
type Result struct {
	Violations []Violation
	ExitCode   int
}

// Violation represents a single code analysis finding.
type Violation struct {
	Rule        string `json:"rule"`
	Description string `json:"description"`
	Severity    int    `json:"severity"`
	File        string `json:"file"`
	Line        int    `json:"line"`
}

// Analyzer wraps the sf code-analyzer command.
type Analyzer struct{}

// New creates a new Analyzer instance.
func New() *Analyzer {
	return &Analyzer{}
}

// Scan runs the Salesforce Code Analyzer on the specified paths.
func (a *Analyzer) Scan(paths []string, ruleset string) (*Result, error) {
	if err := a.EnsureCompatibility(); err != nil {
		return nil, err
	}

	args := []string{
		"code-analyzer",
		"run",
		"--normalize-severity",
		"--output-format", "json",
		"--target", strings.Join(paths, ","),
	}

	if ruleset != "" {
		args = append(args, "--ruleset", ruleset)
	}

	cmd := exec.Command("sf", args...)
	output, err := cmd.Output()
	exitCode := 0
	if exitErr, ok := err.(*exec.ExitError); ok {
		exitCode = exitErr.ExitCode()
		// Analyzer returns non-zero on violations, so we continue
	} else if err != nil {
		gha.Error("Analyzer execution failed: " + err.Error())
		return nil, err
	}

	return a.parseOutput(string(output), exitCode), nil
}

// EnsureCompatibility verifies the environment is correctly configured.
// We explicitly reject legacy "sf scanner" usage in favor of code-analyzer.
func (a *Analyzer) EnsureCompatibility() error {
	// Placeholder for opinionated policy enforcement.
	// In production, we might check for legacy scanner installation and warn/fail.
	return nil
}

// parseOutput extracts violations from the analyzer JSON output.
func (a *Analyzer) parseOutput(jsonOutput string, exitCode int) *Result {
	result := &Result{
		Violations: []Violation{},
		ExitCode:   exitCode,
	}

	// Find JSON array in output (it might have some clutter)
	jsonStart := strings.Index(jsonOutput, "[")
	jsonEnd := strings.LastIndex(jsonOutput, "]")

	if jsonStart == -1 || jsonEnd == -1 || jsonEnd <= jsonStart {
		return result
	}

	cleanJSON := jsonOutput[jsonStart : jsonEnd+1]

	var rawResults []struct {
		FileName   string `json:"fileName"`
		Violations []struct {
			RuleName string `json:"ruleName"`
			Message  string `json:"message"`
			Severity int    `json:"severity"`
			Line     int    `json:"line"`
		} `json:"violations"`
	}

	if err := json.Unmarshal([]byte(cleanJSON), &rawResults); err != nil {
		gha.Warning("Failed to parse analyzer output: " + err.Error())
		return result
	}

	for _, fileResult := range rawResults {
		for _, v := range fileResult.Violations {
			result.Violations = append(result.Violations, Violation{
				Rule:        v.RuleName,
				Description: v.Message,
				Severity:    v.Severity,
				File:        fileResult.FileName,
				Line:        v.Line,
			})
		}
	}

	return result
}
