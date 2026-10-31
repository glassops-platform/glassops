package policy

import (
	"fmt"
	"strings"
)

// AnalyzerMaxSeverity refuses a deploy when Code Analyzer found violations at the configured
// severity or worse. Severity 1 is the worst.
type AnalyzerMaxSeverity struct{}

// Name returns the rule name.
func (AnalyzerMaxSeverity) Name() string { return "analyzer_max_severity" }

// analyzeFallback returns analyze.max_severity from the config, the default the rule falls back
// to when policy.rules.analyzer_max_severity is not set.
func analyzeFallback(in Input) *int {
	if in.Config == nil {
		return nil
	}
	return in.Config.Analyze.MaxSeverity
}

// Evaluate checks analyze.sev1 to analyze.sev5.
func (AnalyzerMaxSeverity) Evaluate(in Input) []Result {
	const rule = "analyzer_max_severity"
	var max *int
	if in.Config != nil {
		max = in.Config.Policy.Rules.AnalyzerMaxSeverity
	}
	if max == nil {
		max = analyzeFallback(in)
	}
	if max == nil {
		return one(rule, Skipped, "no analyzer severity limit is set")
	}
	counts := make([]int, 5)
	found := false
	for i := range counts {
		if n, ok := in.Facts.Int(fmt.Sprintf("analyze.sev%d", i+1)); ok {
			counts[i] = n
			found = true
		}
	}
	if !found {
		return one(rule, Skipped, "no analyze facts are present, so the analyzer did not run")
	}
	var bad []string
	for sev := 1; sev <= *max && sev <= 5; sev++ {
		if counts[sev-1] > 0 {
			bad = append(bad, fmt.Sprintf("%d at severity %d", counts[sev-1], sev))
		}
	}
	if len(bad) > 0 {
		return one(rule, Fail, "violations at severity %d or worse: %s", *max, strings.Join(bad, ", "))
	}
	return one(rule, Pass, "no violations at severity %d or worse", *max)
}
