package policy

import (
	"fmt"
	"strconv"
	"strings"
)

// CoverageMin requires every changed Apex class to meet policy.rules.coverage_min.
type CoverageMin struct{}

// Name returns the rule name.
func (CoverageMin) Name() string { return "coverage_min" }

// isTestClassName is the heuristic for a class that carries no coverage of its own.
func isTestClassName(name string) bool {
	return strings.HasSuffix(name, "Test") || strings.HasSuffix(name, "_Test")
}

// Evaluate checks the per class coverage facts, and the org wide figure when one is present.
func (CoverageMin) Evaluate(in Input) []Result {
	if in.Config == nil || in.Config.Policy.Rules.CoverageMin == nil {
		return one("coverage_min", Skipped, "policy.rules.coverage_min is not set")
	}
	min := *in.Config.Policy.Rules.CoverageMin
	results := classCoverage(in, min)
	if om := orgMin(in); om != nil {
		if r, ok := orgCoverage(in, *om); ok {
			results = append(results, r)
		}
	}
	return results
}

func classCoverage(in Input, min int) []Result {
	classes, ok := in.Facts.Strings("salesforce.apex_classes")
	if !ok {
		return one("coverage_min", Fail, "the plan facts are missing (salesforce.apex_classes), so coverage cannot be checked against %d%%", min)
	}
	var failed []string
	checked, tests := 0, 0
	for _, c := range classes {
		key := "salesforce.coverage." + c + ".percent"
		pct, has := in.Facts.Int(key)
		if !has {
			if _, present := in.Facts.Get(key); !present && isTestClassName(c) {
				tests++
				continue
			}
			failed = append(failed, c+" has no coverage fact")
			continue
		}
		checked++
		if pct < min {
			failed = append(failed, fmt.Sprintf("%s %d%%", c, pct))
		}
	}
	note := ""
	if tests > 0 {
		note = fmt.Sprintf(" %d class(es) with no coverage entry and a name ending in Test or _Test were treated as test classes.", tests)
	}
	if len(failed) > 0 {
		return one("coverage_min", Fail, "below %d%% or unmeasured: %s.%s", min, strings.Join(failed, ", "), note)
	}
	return one("coverage_min", Pass, "%d class(es) at or above %d%%.%s", checked, min, note)
}

// orgCoverage checks test.org_coverage against min. ok is false when the fact is absent.
// It accepts a number or a string such as "87%".
func orgCoverage(in Input, min int) (Result, bool) {
	v, present := in.Facts.Get("test.org_coverage")
	if !present {
		return Result{}, false
	}
	pct, ok := in.Facts.Int("test.org_coverage")
	if !ok {
		if s, isStr := v.(string); isStr {
			if n, err := strconv.Atoi(strings.TrimSuffix(strings.TrimSpace(s), "%")); err == nil {
				pct, ok = n, true
			}
		}
	}
	if !ok {
		return Result{Rule: "coverage_min.org", Outcome: Fail, Detail: "test.org_coverage is not a whole number"}, true
	}
	if pct < min {
		return Result{Rule: "coverage_min.org", Outcome: Fail, Detail: fmt.Sprintf("org wide coverage %d%% is below %d%%", pct, min)}, true
	}
	return Result{Rule: "coverage_min.org", Outcome: Pass, Detail: fmt.Sprintf("org wide coverage %d%% is at or above %d%%", pct, min)}, true
}

// orgMin returns test.coverage.org_min from the config. The config on this branch has no such
// field yet, so it returns nil and the org wide check does not run. Once foundation B adds the
// field, return it here. orgCoverage is tested directly.
func orgMin(Input) *int { return nil }
