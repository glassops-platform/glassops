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

// IsTestClassName is the name heuristic for a class that carries no coverage of its own. It is
// only the fallback for a plan written before the fact salesforce.apex_test_classes existed.
// The plan now records which classes are annotated @IsTest, and that fact wins when present.
func IsTestClassName(name string) bool {
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
	// salesforce.apex_test_classes lists the changed classes whose source is annotated @IsTest.
	// An older plan has no such fact, and then the name heuristic decides.
	declared, haveDeclared := in.Facts.Strings("salesforce.apex_test_classes")
	isDeclared := map[string]bool{}
	for _, d := range declared {
		isDeclared[d] = true
	}
	// salesforce.apex_nonexecutable_types lists the changed classes whose top level type is an
	// interface or an enum. They have no executable lines, so Salesforce lists no coverage for them.
	nonExec := map[string]bool{}
	if names, ok := in.Facts.Strings("salesforce.apex_nonexecutable_types"); ok {
		for _, n := range names {
			nonExec[n] = true
		}
	}
	var failed []string
	checked, tests, unmeasured := 0, 0, 0
	for _, c := range classes {
		key := "salesforce.coverage." + c + ".percent"
		pct, has := in.Facts.Int(key)
		if !has {
			if _, present := in.Facts.Get(key); !present {
				if nonExec[c] {
					unmeasured++
					continue
				}
				if isDeclared[c] {
					tests++
					continue
				}
				if !haveDeclared && IsTestClassName(c) {
					tests++
					continue
				}
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
		if haveDeclared {
			note = fmt.Sprintf(" %d class(es) with no coverage entry were treated as test classes because their source is annotated @IsTest.", tests)
		} else {
			note = fmt.Sprintf(" %d class(es) with no coverage entry and a name ending in Test or _Test were treated as test classes.", tests)
		}
	}
	if unmeasured > 0 {
		note += fmt.Sprintf(" %d interface or enum type(s) have no executable lines and were not measured.", unmeasured)
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

// orgMin returns test.coverage.org_min from the config, or nil when it is not set, and then the
// org wide check does not run.
func orgMin(in Input) *int {
	if in.Config == nil {
		return nil
	}
	return in.Config.Test.Coverage.OrgMin
}
