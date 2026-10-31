package policy

import (
	"strings"
	"testing"

	"github.com/glassops-platform/glassops/internal/core/config"
)

func covCfg(min *int) *config.Config {
	return cfgWith(func(c *config.Config) { c.Policy.Rules.CoverageMin = min })
}

func TestCoverageMin(t *testing.T) {
	tests := []struct {
		name   string
		cfg    *config.Config
		facts  map[string]any
		want   Outcome
		detail string
	}{
		{"rule absent", covCfg(nil), map[string]any{}, Skipped, "not set"},
		{"all at or above", covCfg(intp(75)), map[string]any{
			"salesforce.apex_classes":       []string{"A", "B"},
			"salesforce.coverage.A.percent": 75,
			"salesforce.coverage.B.percent": 100,
		}, Pass, "2 class"},
		{"one below", covCfg(intp(75)), map[string]any{
			"salesforce.apex_classes":       []string{"A", "B"},
			"salesforce.coverage.A.percent": 74,
			"salesforce.coverage.B.percent": 100,
		}, Fail, "A 74%"},
		{"missing coverage fails, never passes", covCfg(intp(75)), map[string]any{
			"salesforce.apex_classes":       []string{"A", "B"},
			"salesforce.coverage.A.percent": 90,
		}, Fail, "B has no coverage fact"},
		{"test class heuristic Test", covCfg(intp(75)), map[string]any{
			"salesforce.apex_classes":       []string{"A", "ATest"},
			"salesforce.coverage.A.percent": 90,
		}, Pass, "treated as test classes"},
		{"test class heuristic _Test", covCfg(intp(75)), map[string]any{
			"salesforce.apex_classes":       []string{"A", "A_Test"},
			"salesforce.coverage.A.percent": 90,
		}, Pass, "Test or _Test"},
		{"a Test named class with coverage is held to the minimum", covCfg(intp(75)), map[string]any{
			"salesforce.apex_classes":           []string{"ATest"},
			"salesforce.coverage.ATest.percent": 10,
		}, Fail, "ATest 10%"},
		{"no apex classes fact", covCfg(intp(75)), map[string]any{}, Fail, "plan facts are missing"},
		{"empty class list passes", covCfg(intp(75)), map[string]any{"salesforce.apex_classes": []string{}}, Pass, "0 class"},
		{"coverage through JSON types", covCfg(intp(75)), map[string]any{
			"salesforce.apex_classes":       []any{"A"},
			"salesforce.coverage.A.percent": float64(80),
		}, Pass, "1 class"},
		{"non numeric coverage fails", covCfg(intp(75)), map[string]any{
			"salesforce.apex_classes":       []string{"A"},
			"salesforce.coverage.A.percent": "80",
		}, Fail, "A has no coverage fact"},
		{"minimum zero", covCfg(intp(0)), map[string]any{
			"salesforce.apex_classes":       []string{"A"},
			"salesforce.coverage.A.percent": 0,
		}, Pass, "0%"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			r := one1(t, CoverageMin{}.Evaluate(Input{Facts: doc(tc.facts), Config: tc.cfg}))
			if r.Outcome != tc.want || r.Rule != "coverage_min" || !strings.Contains(r.Detail, tc.detail) {
				t.Fatalf("got %+v, want %s containing %q", r, tc.want, tc.detail)
			}
		})
	}
}

func TestOrgCoverage(t *testing.T) {
	tests := []struct {
		name    string
		v       any
		present bool
		want    Outcome
	}{
		{"absent is not checked", nil, false, ""},
		{"number above", 80, true, Pass},
		{"number below", 74, true, Fail},
		{"string with percent", "100%", true, Pass},
		{"string with percent below", " 60% ", true, Fail},
		{"float from JSON", float64(75), true, Pass},
		{"not a number fails", "high", true, Fail},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			kv := map[string]any{}
			if tc.present {
				kv["test.org_coverage"] = tc.v
			}
			r, ok := orgCoverage(Input{Facts: doc(kv)}, 75)
			if ok != tc.present {
				t.Fatalf("ok = %v", ok)
			}
			if ok && (r.Outcome != tc.want || r.Rule != "coverage_min.org") {
				t.Fatalf("got %+v, want %s", r, tc.want)
			}
		})
	}
}
