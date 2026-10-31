package policy

import (
	"testing"

	"github.com/glassops-platform/glassops/internal/core/config"
)

func TestDefaultRules(t *testing.T) {
	var names []string
	for _, r := range Default() {
		names = append(names, r.Name())
	}
	want := []string{"plan_not_empty", "window", "blackout", "coverage_min", "destructive_needs_label", "separation_of_duties", "analyzer_max_severity"}
	if len(names) != len(want) {
		t.Fatalf("rules = %v", names)
	}
	for i := range want {
		if names[i] != want[i] {
			t.Fatalf("rules = %v, want %v", names, want)
		}
	}
}

func TestDefaultEvaluateEmptyConfig(t *testing.T) {
	d := doc(map[string]any{"target.tier": "prod", "change.empty": false})
	allow, results := Evaluate(Default(), Input{Facts: d, Config: &config.Config{}, Now: at(t, "2027-01-04T12:00:00Z")})
	if !allow {
		t.Fatalf("an empty config should allow a non empty plan: %+v", results)
	}
	if len(results) != len(Default()) {
		t.Fatalf("got %d results: %+v", len(results), results)
	}
}

func TestDefaultEvaluateReportsEveryFailure(t *testing.T) {
	cfg := cfgWith(func(c *config.Config) {
		c.Policy.Rules.CoverageMin = intp(75)
		c.Policy.Rules.SeparationOfDuties = &config.SoDRule{Tiers: []string{"prod"}}
	})
	d := doc(map[string]any{"target.tier": "prod", "change.empty": true, "ci.pr_author": "a", "ci.actor": "a"})
	allow, results := Evaluate(Default(), Input{Facts: d, Config: cfg, Now: at(t, "2027-01-04T12:00:00Z")})
	if allow {
		t.Fatal("expected deny")
	}
	failed := map[string]bool{}
	for _, r := range results {
		if r.Outcome == Fail {
			failed[r.Rule] = true
		}
	}
	for _, name := range []string{"plan_not_empty", "coverage_min", "separation_of_duties"} {
		if !failed[name] {
			t.Errorf("%s did not fail: %+v", name, results)
		}
	}
}
