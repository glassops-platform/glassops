package policy

import (
	"testing"

	"github.com/glassops-platform/glassops/internal/core/config"
	"github.com/glassops-platform/glassops/internal/core/facts"
)

// These tests exist to kill mutants that the table tests let live. Each one names the mutant.

// window.go: `mins >= from` in the midnight crossing branch. The window opens on its start
// minute, so a boundary mutant (`>`) would refuse the first minute.
func TestMutWindowOvernightStartMinute(t *testing.T) {
	overnight := win([]string{"fri"}, "22:00", "06:00", "UTC")
	// 2027-01-08 is a Friday.
	in := Input{Facts: doc(map[string]any{"target.tier": "prod"}), Config: windowCfg(overnight), Now: at(t, "2027-01-08T22:00:00Z")}
	r := one1(t, Window{}.Evaluate(in))
	if r.Outcome != Pass {
		t.Fatalf("the start minute of an overnight window is inside it: %+v", r)
	}
	in.Now = at(t, "2027-01-08T21:59:59Z")
	if r := one1(t, Window{}.Evaluate(in)); r.Outcome != Fail {
		t.Fatalf("one second before the start is outside: %+v", r)
	}
}

// window.go: `from < to` against `from <= to`. A window whose from equals its to is empty.
// The mutant `<=` is equivalent, because both branches then report false, and it is left alone.
func TestMutWindowEmptyWhenFromEqualsTo(t *testing.T) {
	w := win(allDays, "09:00", "09:00", "UTC")
	for _, now := range []string{"2027-01-04T08:59:00Z", "2027-01-04T09:00:00Z", "2027-01-04T09:01:00Z"} {
		in := Input{Facts: doc(map[string]any{"target.tier": "prod"}), Config: windowCfg(w), Now: at(t, now)}
		if r := one1(t, Window{}.Evaluate(in)); r.Outcome != Fail {
			t.Fatalf("an empty window allows nothing at %s: %+v", now, r)
		}
	}
}

// coverage_min.go: `checked++` and `tests > 0` are visible only in the exact detail text.
func TestMutCoverageDetailExact(t *testing.T) {
	cfg := covCfg(intp(75))
	tests := []struct {
		name   string
		facts  map[string]any
		detail string
	}{
		{"one class, no test classes", map[string]any{
			"salesforce.apex_classes":       []string{"A"},
			"salesforce.coverage.A.percent": 75,
		}, "1 class(es) at or above 75%."},
		{"two classes, no test classes", map[string]any{
			"salesforce.apex_classes":       []string{"A", "B"},
			"salesforce.coverage.A.percent": 80,
			"salesforce.coverage.B.percent": 75,
		}, "2 class(es) at or above 75%."},
		{"one test class", map[string]any{
			"salesforce.apex_classes":       []string{"A", "ATest"},
			"salesforce.coverage.A.percent": 80,
		}, "1 class(es) at or above 75%. 1 class(es) with no coverage entry and a name ending in Test or _Test were treated as test classes."},
		{"two test classes", map[string]any{
			"salesforce.apex_classes":       []string{"A", "ATest", "B_Test"},
			"salesforce.coverage.A.percent": 80,
		}, "1 class(es) at or above 75%. 2 class(es) with no coverage entry and a name ending in Test or _Test were treated as test classes."},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			r := one1(t, CoverageMin{}.Evaluate(Input{Facts: doc(tc.facts), Config: cfg}))
			if r.Outcome != Pass || r.Detail != tc.detail {
				t.Fatalf("got %+v, want pass %q", r, tc.detail)
			}
		})
	}
}

// coverage_min.go: the boundary between pass and fail is exactly the minimum.
func TestMutCoverageBoundary(t *testing.T) {
	for _, tc := range []struct {
		pct  int
		want Outcome
	}{{74, Fail}, {75, Pass}, {76, Pass}} {
		in := Input{Facts: doc(map[string]any{
			"salesforce.apex_classes":       []string{"A"},
			"salesforce.coverage.A.percent": tc.pct,
		}), Config: covCfg(intp(75))}
		if r := one1(t, CoverageMin{}.Evaluate(in)); r.Outcome != tc.want {
			t.Fatalf("class at %d: got %+v, want %s", tc.pct, r, tc.want)
		}
	}
	// The org wide figure has the same boundary, as a number and as a string.
	cfg := cfgWith(func(c *config.Config) {
		c.Policy.Rules.CoverageMin = intp(0)
		c.Test.Coverage.OrgMin = intp(75)
	})
	for _, tc := range []struct {
		val  any
		want Outcome
	}{{74, Fail}, {75, Pass}, {"74%", Fail}, {"75%", Pass}, {" 75 % ", Fail}, {"x", Fail}, {true, Fail}} {
		in := Input{Facts: doc(map[string]any{
			"salesforce.apex_classes": []string{},
			"test.org_coverage":       tc.val,
		}), Config: cfg}
		rs := CoverageMin{}.Evaluate(in)
		var org *Result
		for i := range rs {
			if rs[i].Rule == "coverage_min.org" {
				org = &rs[i]
			}
		}
		if org == nil || org.Outcome != tc.want {
			t.Fatalf("org coverage %#v: got %+v, want %s", tc.val, rs, tc.want)
		}
	}
}

// separation_of_duties.go: every part of each approval condition must matter on its own.
func TestMutSoDApprovalConditions(t *testing.T) {
	same := func(approvals ...facts.Approval) map[string]any {
		return map[string]any{"target.tier": "prod", "ci.pr_author": "alice", "ci.actor": "alice", "ci.approvals": approvals}
	}
	env := &config.SoDApproval{Environment: "prod-deploy"}
	label := &config.SoDApproval{Label: "ok", Approvers: []string{"carol"}}
	both := &config.SoDApproval{Environment: "prod-deploy", Label: "ok", Approvers: []string{"carol"}}
	tests := []struct {
		name  string
		appr  *config.SoDApproval
		facts map[string]any
		want  Outcome
	}{
		{"env ok", env, same(facts.Approval{Mechanism: "environment", Environment: "prod-deploy", State: "approved", Approver: "c"}), SatisfiedByApproval},
		{"env: label mechanism with matching environment", env, same(facts.Approval{Mechanism: "label", Environment: "prod-deploy", State: "approved", Approver: "c"}), Fail},
		{"env: pending state", env, same(facts.Approval{Mechanism: "environment", Environment: "prod-deploy", State: "pending", Approver: "c"}), Fail},
		{"env: empty state", env, same(facts.Approval{Mechanism: "environment", Environment: "prod-deploy", Approver: "c"}), Fail},
		{"env: other environment", env, same(facts.Approval{Mechanism: "environment", Environment: "dev", State: "approved", Approver: "c"}), Fail},
		{"env not configured, evidence has no environment either", &config.SoDApproval{Label: "ok", Approvers: []string{"carol"}},
			same(facts.Approval{Mechanism: "environment", State: "approved", Approver: "carol"}), Fail},
		{"env configured empty, both empty", &config.SoDApproval{Environment: "", Label: "x"},
			same(facts.Approval{Mechanism: "environment", Environment: "", State: "approved", Approver: "c"}), Fail},
		{"label ok", label, same(facts.Approval{Mechanism: "label", Label: "ok", Approver: "CAROL"}), SatisfiedByApproval},
		{"label: environment mechanism with matching label", label, same(facts.Approval{Mechanism: "environment", Label: "ok", Approver: "carol"}), Fail},
		{"label: other label", label, same(facts.Approval{Mechanism: "label", Label: "no", Approver: "carol"}), Fail},
		{"label: approver not listed", label, same(facts.Approval{Mechanism: "label", Label: "ok", Approver: "mallory"}), Fail},
		{"label: no approver recorded", label, same(facts.Approval{Mechanism: "label", Label: "ok"}), Fail},
		{"label not configured, evidence label empty, approver listed", &config.SoDApproval{Environment: "prod-deploy", Approvers: []string{"carol"}},
			same(facts.Approval{Mechanism: "label", Approver: "carol"}), Fail},
		{"both configured, label satisfies", both, same(facts.Approval{Mechanism: "label", Label: "ok", Approver: "carol"}), SatisfiedByApproval},
		{"both configured, env satisfies", both, same(facts.Approval{Mechanism: "environment", Environment: "prod-deploy", State: "approved", Approver: "x"}), SatisfiedByApproval},
		{"both configured, first approval is junk, second satisfies", both, same(
			facts.Approval{Mechanism: "label", Label: "no", Approver: "carol"},
			facts.Approval{Mechanism: "environment", Environment: "prod-deploy", State: "approved", Approver: "x"}), SatisfiedByApproval},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			cfg := sodCfg(tc.appr, "prod")
			r := one1(t, SeparationOfDuties{}.Evaluate(Input{Facts: doc(tc.facts), Config: cfg}))
			if r.Outcome != tc.want {
				t.Fatalf("got %+v, want %s", r, tc.want)
			}
		})
	}
}

// blackout.go: a blackout that does not apply must not hide a later one that does, the reason
// is part of the hit, and an unknown tier is called out. The extended mutators (loop control,
// self assignment) reach these.
func TestMutBlackoutScan(t *testing.T) {
	past := config.Blackout{From: "2020-01-01", To: "2020-01-02", Reason: "past"}
	otherTier := config.Blackout{From: "2027-06-01", To: "2027-06-30", Tiers: []string{"qa"}, Reason: "qa only"}
	live := config.Blackout{From: "2027-06-10", To: "2027-06-20", Tiers: []string{"prod"}, Reason: "audit"}
	noReason := config.Blackout{From: "2027-06-15", To: "2027-06-15"}
	cfg := cfgWith(func(c *config.Config) {
		c.Policy.Blackouts = []config.Blackout{past, otherTier, live, noReason}
	})
	now := at(t, "2027-06-15T12:00:00Z")

	r := one1(t, Blackout{}.Evaluate(Input{Facts: doc(map[string]any{"target.tier": "prod"}), Config: cfg, Now: now}))
	want := "blackout in force on 2027-06-15: 2027-06-10 to 2027-06-20 (audit), 2027-06-15 to 2027-06-15"
	if r.Outcome != Fail || r.Detail != want {
		t.Fatalf("got %+v, want fail %q", r, want)
	}

	// A tier that is not named in the facts is treated as covered by everything.
	r = one1(t, Blackout{}.Evaluate(Input{Facts: doc(nil), Config: cfg, Now: now}))
	want = "blackout in force on 2027-06-15: 2027-06-01 to 2027-06-30 (qa only), 2027-06-10 to 2027-06-20 (audit), 2027-06-15 to 2027-06-15 (the facts do not name the target tier, so it is treated as covered)"
	if r.Outcome != Fail || r.Detail != want {
		t.Fatalf("got %+v, want fail %q", r, want)
	}

	r = one1(t, Blackout{}.Evaluate(Input{Facts: doc(map[string]any{"target.tier": "dev"}), Config: cfg, Now: now}))
	if r.Outcome != Fail || r.Detail != "blackout in force on 2027-06-15: 2027-06-15 to 2027-06-15" {
		t.Fatalf("got %+v", r)
	}
	cfg.Policy.Blackouts = []config.Blackout{past, otherTier}
	r = one1(t, Blackout{}.Evaluate(Input{Facts: doc(map[string]any{"target.tier": "prod"}), Config: cfg, Now: now}))
	if r.Outcome != Pass || r.Detail != "no blackout covers 2027-06-15 for this target" {
		t.Fatalf("got %+v", r)
	}
}

// helpers.go tierOf: a tier fact that is present but empty names no tier.
func TestMutEmptyTierFactNamesNoTier(t *testing.T) {
	cfg := windowCfg(win(allDays, "00:00", "23:59", "UTC"))
	in := Input{Facts: doc(map[string]any{"target.tier": ""}), Config: cfg, Now: at(t, "2027-01-04T12:00:00Z")}
	if r := one1(t, Window{}.Evaluate(in)); r.Outcome != Fail || r.Rule != "window" {
		t.Fatalf("an empty tier must not be treated as a tier: %+v", r)
	}
	in.Config = sodCfg(nil, "prod")
	if r := one1(t, SeparationOfDuties{}.Evaluate(in)); r.Outcome != Fail || r.Detail != "the facts do not name the target tier" {
		t.Fatalf("separation of duties with an empty tier: %+v", r)
	}
}

// coverage_min.go: the scan keeps going after a class with no coverage fact, so every
// offender is named.
func TestMutCoverageNamesEveryOffender(t *testing.T) {
	in := Input{Facts: doc(map[string]any{
		"salesforce.apex_classes":       []string{"A", "B", "C"},
		"salesforce.coverage.B.percent": 10,
		"salesforce.coverage.C.percent": 90,
	}), Config: covCfg(intp(75))}
	r := one1(t, CoverageMin{}.Evaluate(in))
	want := "below 75% or unmeasured: A has no coverage fact, B 10%."
	if r.Outcome != Fail || r.Detail != want {
		t.Fatalf("got %+v, want fail %q", r, want)
	}
}

// separation_of_duties.go: an empty deploying actor is as unknown as a missing one, and an
// empty author likewise, in every mode.
func TestMutSoDEmptyIdentityIsUnknown(t *testing.T) {
	approval := &config.SoDApproval{Environment: "prod-deploy"}
	for _, appr := range []*config.SoDApproval{nil, approval} {
		for name, kv := range map[string]map[string]any{
			"empty actor":  {"target.tier": "prod", "ci.pr_author": "alice", "ci.actor": ""},
			"empty author": {"target.tier": "prod", "ci.pr_author": "", "ci.actor": "bob"},
			"both empty":   {"target.tier": "prod", "ci.pr_author": "", "ci.actor": ""},
		} {
			r := one1(t, SeparationOfDuties{}.Evaluate(Input{Facts: doc(kv), Config: sodCfg(appr, "prod")}))
			if r.Outcome != Fail {
				t.Fatalf("%s (approval %v): got %+v, want fail", name, appr != nil, r)
			}
		}
	}
}

// separation_of_duties.go: with no approval configured, an unknown identity and a shared
// identity are different failures and say so. A mutant that merges the first two cases still
// fails the rule, so only the detail tells them apart.
func TestMutSoDEnforceDetailTellsUnknownFromSame(t *testing.T) {
	run := func(kv map[string]any) Result {
		kv["target.tier"] = "prod"
		return one1(t, SeparationOfDuties{}.Evaluate(Input{Facts: doc(kv), Config: sodCfg(nil, "prod")}))
	}
	for name, kv := range map[string]map[string]any{
		"author unknown": {"ci.actor": "bob"},
		"actor unknown":  {"ci.pr_author": "alice"},
		"both unknown":   {},
	} {
		r := run(kv)
		if r.Outcome != Fail || r.Detail != "enforced for tier prod but the pull request author or the deploying actor is unknown" {
			t.Fatalf("%s: %+v", name, r)
		}
	}
	r := run(map[string]any{"ci.pr_author": "alice", "ci.actor": "ALICE"})
	if r.Outcome != Fail || r.Detail != "enforced for tier prod: ALICE authored the change and is also deploying it" {
		t.Fatalf("same person: %+v", r)
	}
}
