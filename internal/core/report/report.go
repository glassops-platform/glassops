// Package report reads every result a run wrote and turns them into one markdown summary and
// one SARIF log. It only reads. It never decides whether a deploy may happen.
package report

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/glassops-platform/glassops/internal/core/permit"
	"github.com/glassops-platform/glassops/internal/core/run"
	"github.com/glassops-platform/glassops/internal/core/sarif"
)

// ErrNoResults means the results directory held no result. A report with no evidence is an
// error, never a pass.
var ErrNoResults = errors.New("report: no results found")

// Bundle is everything a report is built from.
type Bundle struct {
	// Results are the step results, oldest first.
	Results []run.Result
	// Permit is the PERMIT when one was found.
	Permit *permit.Permit
	// Foreign holds SARIF logs already in the results directory, such as Code Analyzer's.
	Foreign []*sarif.Log
	// Problems lists files that could not be read. Any problem makes the report an error.
	Problems []string
}

// Load reads every *.json and *.sarif file in dir. permitPaths are extra places to look for
// a PERMIT. The first one that exists is used, and a PERMIT found in dir is used when none of
// them exists.
func Load(dir string, permitPaths ...string) (*Bundle, error) {
	return LoadSkipping(dir, nil, permitPaths...)
}

// LoadSkipping is Load that ignores the SARIF files named in skip, so a report that writes
// its own SARIF into the results directory does not merge its last output into the next.
func LoadSkipping(dir string, skip []string, permitPaths ...string) (*Bundle, error) {
	b := &Bundle{}
	jsons, err := filepath.Glob(filepath.Join(dir, "*.json"))
	if err != nil {
		return nil, fmt.Errorf("list results: %w", err)
	}
	sort.Strings(jsons)
	for _, p := range jsons {
		b.loadJSON(p)
	}
	sarifs, err := filepath.Glob(filepath.Join(dir, "*.sarif"))
	if err != nil {
		return nil, fmt.Errorf("list sarif: %w", err)
	}
	sort.Strings(sarifs)
	for _, p := range sarifs {
		if skipped(p, skip) {
			continue
		}
		l, err := sarif.Read(p)
		if err != nil {
			b.Problems = append(b.Problems, fmt.Sprintf("%s: %v", filepath.Base(p), err))
			continue
		}
		b.Foreign = append(b.Foreign, l)
	}
	for _, p := range permitPaths {
		if p == "" {
			continue
		}
		if _, err := os.Stat(p); err != nil {
			continue
		}
		b.Permit = nil
		b.loadJSON(p)
		break
	}
	if len(b.Results) == 0 && len(b.Problems) == 0 {
		return nil, ErrNoResults
	}
	sort.SliceStable(b.Results, func(i, j int) bool {
		a, c := b.Results[i], b.Results[j]
		if !a.Started.Equal(c.Started) {
			return a.Started.Before(c.Started)
		}
		return a.Step < c.Step
	})
	return b, nil
}

// loadJSON reads one JSON file and files it by its schema.
func (b *Bundle) loadJSON(path string) {
	name := filepath.Base(path)
	data, err := os.ReadFile(path)
	if err != nil {
		b.Problems = append(b.Problems, fmt.Sprintf("%s: %v", name, err))
		return
	}
	var head struct {
		Schema string `json:"schema"`
	}
	if err := json.Unmarshal(data, &head); err != nil {
		b.Problems = append(b.Problems, fmt.Sprintf("%s: not valid JSON", name))
		return
	}
	switch head.Schema {
	case run.Schema:
		var r run.Result
		if err := json.Unmarshal(data, &r); err != nil {
			b.Problems = append(b.Problems, fmt.Sprintf("%s: not a valid result", name))
			return
		}
		b.Results = append(b.Results, r)
	case permit.Schema:
		var p permit.Permit
		if err := json.Unmarshal(data, &p); err != nil {
			b.Problems = append(b.Problems, fmt.Sprintf("%s: not a valid permit", name))
			return
		}
		b.Permit = &p
	default:
		b.Problems = append(b.Problems, fmt.Sprintf("%s: unknown schema %q", name, head.Schema))
	}
}

// Worst is the status the report as a whole exits with: error beats fail beats the rest.
// Unreadable evidence counts as an error.
func (b *Bundle) Worst() run.Status {
	worst := run.StatusPass
	if len(b.Problems) > 0 {
		return run.StatusError
	}
	for _, r := range b.Results {
		switch r.Status {
		case run.StatusError:
			return run.StatusError
		case run.StatusFail:
			worst = run.StatusFail
		}
	}
	return worst
}

// ExitCode is the process exit code for the report.
func (b *Bundle) ExitCode() int { return b.Worst().ExitCode() }

// deployIDFacts are the fact keys a deploy result may carry its Id under, in order.
var deployIDFacts = []string{"salesforce.deploy_id", "deploy.id", "deploy_id"}

// DeployID returns the Id of the deploy when a deploy result holds one.
func (b *Bundle) DeployID() string {
	for _, r := range b.Results {
		if r.Step != "deploy" {
			continue
		}
		for _, k := range deployIDFacts {
			if v, ok := r.Facts[k].(string); ok && v != "" {
				return v
			}
		}
	}
	return ""
}

// stepOf returns the result for a step name.
func (b *Bundle) stepOf(name string) *run.Result {
	for i := range b.Results {
		if b.Results[i].Step == name {
			return &b.Results[i]
		}
	}
	return nil
}

// failingChecks returns the checks that failed or errored.
func failingChecks(r run.Result) []run.Check {
	var out []run.Check
	for _, c := range r.Checks {
		if c.Status == run.StatusFail || c.Status == run.StatusError {
			out = append(out, c)
		}
	}
	return out
}

// clean makes free text safe for one table cell or one list line.
func clean(s string) string {
	s = strings.Join(strings.Fields(s), " ")
	return strings.ReplaceAll(s, "|", "\\|")
}

// skipped reports whether path is one of the files to ignore.
func skipped(path string, skip []string) bool {
	abs, err := filepath.Abs(path)
	if err != nil {
		return false
	}
	for _, s := range skip {
		if sa, err := filepath.Abs(s); err == nil && sa == abs {
			return true
		}
	}
	return false
}
