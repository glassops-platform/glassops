package policy

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/glassops-platform/glassops/internal/core/config"
	"github.com/glassops-platform/glassops/internal/core/facts"
)

// mutFullConfig turns on every rule, so each one reads the fuzzed facts.
func mutFullConfig() *config.Config {
	return &config.Config{
		Policy: config.Policy{
			Windows: map[string][]config.Window{
				"prod": {{Days: []string{"mon", "tue", "wed"}, From: "09:00", To: "17:00", TZ: "America/New_York"}},
				"qa":   {{Days: []string{"fri"}, From: "22:00", To: "06:00", TZ: "UTC"}},
			},
			Blackouts: []config.Blackout{{From: "2027-12-20", To: "2028-01-02", Tiers: []string{"prod"}, Reason: "freeze"}},
			Rules: config.Rules{
				CoverageMin:           intp(75),
				DestructiveNeedsLabel: &config.DestructiveLabelRule{Tiers: []string{"prod"}, Label: "destructive-ok"},
				SeparationOfDuties: &config.SoDRule{
					Tiers:    []string{"prod"},
					Approval: &config.SoDApproval{Environment: "prod-deploy", Label: "second-eyes", Approvers: []string{"carol"}},
				},
			},
		},
		Test: config.Test{Coverage: config.Coverage{OrgMin: intp(75)}},
	}
}

// FuzzMutPolicyFacts decodes arbitrary JSON as a facts document and runs every default rule
// over it. No rule may panic, and the verdict must follow the results: it allows exactly when
// no result is a Fail. A document that does not say whether the plan is empty, or says it is,
// must never be allowed.
func FuzzMutPolicyFacts(f *testing.F) {
	f.Add([]byte(`{"facts":{"target.tier":"prod","change.empty":false,"salesforce.apex_classes":["A"],"salesforce.coverage.A.percent":80,"change.destructive_count":0,"ci.pr_author":"alice","ci.actor":"bob"}}`), int64(1799000000))
	f.Add([]byte(`{"facts":{"target.tier":"prod","change.empty":true}}`), int64(0))
	f.Add([]byte(`{"facts":{"target.tier":"qa","change.empty":false,"ci.approvals":[{"mechanism":"label","label":"second-eyes","approver":"carol"}],"ci.pr_author":"a","ci.actor":"a"}}`), int64(1830000000))
	f.Add([]byte(`{"facts":{"target.tier":["prod"],"change.empty":"no","salesforce.apex_classes":"A","test.org_coverage":"87%"}}`), int64(-1))
	f.Add([]byte(`{"facts":{"target.tier":"prod","change.empty":false,"ci.approvals":[1,2,{"mechanism":5}],"ci.labels":[1],"change.destructive_count":1.5}}`), int64(1))
	f.Add([]byte(`{}`), int64(0))
	f.Add([]byte(`{"facts":null}`), int64(0))
	f.Add([]byte(`[]`), int64(0))
	cfg := mutFullConfig()
	f.Fuzz(func(t *testing.T, data []byte, unix int64) {
		var d facts.Document
		if err := json.Unmarshal(data, &d); err != nil {
			return
		}
		if d.Facts == nil {
			d.Facts = map[string]any{}
		}
		now := time.Unix(unix%10_000_000_000, 0).UTC()
		allow, results := Evaluate(Default(), Input{Facts: &d, Config: cfg, Now: now})
		denied := false
		for _, r := range results {
			if r.Outcome.Denies() {
				denied = true
			}
			if r.Rule == "" || r.Outcome == "" {
				t.Fatalf("a result without a rule or outcome: %+v", r)
			}
		}
		if allow == denied {
			t.Fatalf("allow=%v with denied=%v: %+v", allow, denied, results)
		}
		if empty, ok := d.Bool(facts.ChangeEmpty); (!ok || empty) && allow {
			t.Fatalf("allowed a plan that is empty or of unknown size: %+v", results)
		}
		// With no config at all every rule must still decide without panicking.
		_, _ = Evaluate(Default(), Input{Facts: &d, Config: nil, Now: now})
	})
}
