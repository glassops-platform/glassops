// Package run holds the result envelope every command writes and the .glassops store.
package run

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/glassops-platform/glassops/internal/core/exitcode"
)

// Schema is the identifier written into every result envelope.
const Schema = "glassops.result/v1"

// Status is the outcome of one step.
type Status string

const (
	StatusPass    Status = "pass"
	StatusFail    Status = "fail"
	StatusError   Status = "error"
	StatusSkipped Status = "skipped"
	StatusEmpty   Status = "empty"
)

// ExitCode maps a status to the process exit code.
func (s Status) ExitCode() int {
	switch s {
	case StatusFail:
		return exitcode.Fail
	case StatusError:
		return exitcode.Error
	default:
		return exitcode.OK
	}
}

// Check is one named check inside a step, such as a tool version or a config rule.
type Check struct {
	Name   string `json:"name"`
	Status Status `json:"status"`
	Detail string `json:"detail,omitempty"`
}

// Result is the envelope written to .glassops/results/<step>.json.
type Result struct {
	Schema    string         `json:"schema"`
	Step      string         `json:"step"`
	Status    Status         `json:"status"`
	Started   time.Time      `json:"started"`
	Ended     time.Time      `json:"ended"`
	Target    string         `json:"target,omitempty"`
	Provider  string         `json:"provider,omitempty"`
	Checks    []Check        `json:"checks,omitempty"`
	Facts     map[string]any `json:"facts"`
	Artifacts []string       `json:"artifacts"`
	Errors    []string       `json:"errors"`
}

// New starts a result for a step. The clock is injectable for tests.
func New(step string, now func() time.Time) *Result {
	if now == nil {
		now = time.Now
	}
	return &Result{
		Schema:    Schema,
		Step:      step,
		Status:    StatusPass,
		Started:   now().UTC(),
		Facts:     map[string]any{},
		Artifacts: []string{},
		Errors:    []string{},
	}
}

// AddCheck records a check and lowers the step status when the check is worse.
func (r *Result) AddCheck(name string, status Status, detail string) {
	r.Checks = append(r.Checks, Check{Name: name, Status: status, Detail: detail})
	r.Degrade(status)
}

// Degrade lowers the step status to s when s is worse. Error beats fail beats pass.
func (r *Result) Degrade(s Status) {
	if rank(s) > rank(r.Status) {
		r.Status = s
	}
}

// AddError records an error message and marks the step as errored.
func (r *Result) AddError(err error) {
	r.Errors = append(r.Errors, err.Error())
	r.Degrade(StatusError)
}

func rank(s Status) int {
	switch s {
	case StatusError:
		return 3
	case StatusFail:
		return 2
	case StatusPass:
		return 1
	default:
		return 0
	}
}

// Write finishes the result and writes it to <dir>/<step>.json. It returns the file path.
func (r *Result) Write(dir string, now func() time.Time) (string, error) {
	if now == nil {
		now = time.Now
	}
	r.Ended = now().UTC()
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", fmt.Errorf("create results dir: %w", err)
	}
	data, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		return "", fmt.Errorf("encode result: %w", err)
	}
	path := filepath.Join(dir, r.Step+".json")
	if err := os.WriteFile(path, append(data, '\n'), 0o644); err != nil {
		return "", fmt.Errorf("write result: %w", err)
	}
	return path, nil
}
