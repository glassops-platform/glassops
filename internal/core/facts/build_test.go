package facts

import (
	"encoding/json"
	"testing"

	"github.com/glassops-platform/glassops/internal/core/run"
	"github.com/glassops-platform/glassops/internal/providers"
)

// Synthetic change sets, built here and not read from a fixture.
func TestFromChangeSet(t *testing.T) {
	tests := []struct {
		name      string
		cs        providers.ChangeSet
		wantEmpty bool
		wantDest  int
	}{
		{"provider count wins", providers.ChangeSet{Provider: "salesforce", SHA256: "abc", Destructive: []string{"a.xml"},
			Facts: map[string]any{"salesforce.destructive_count": 3}}, false, 3},
		{"count from json float", providers.ChangeSet{Provider: "salesforce",
			Facts: map[string]any{"salesforce.destructive_count": float64(2)}}, false, 2},
		{"manifests when no fact", providers.ChangeSet{Provider: "salesforce", Destructive: []string{"a.xml", "b.xml"}}, false, 2},
		{"empty plan", providers.ChangeSet{Provider: "salesforce", Empty: true}, true, 0},
		{"fraction ignored", providers.ChangeSet{Provider: "salesforce", Destructive: []string{"a.xml"},
			Facts: map[string]any{"salesforce.destructive_count": 1.5}}, false, 1},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			d := FromChangeSet(tc.cs)
			if d.Schema != Schema {
				t.Fatalf("schema %q", d.Schema)
			}
			if e, ok := d.Bool(ChangeEmpty); !ok || e != tc.wantEmpty {
				t.Fatalf("empty = %v, %v", e, ok)
			}
			if n, ok := d.Int(ChangeDestructiveCount); !ok || n != tc.wantDest {
				t.Fatalf("destructive_count = %d, %v, want %d", n, ok, tc.wantDest)
			}
			if s, _ := d.String(ChangePlanSHA256); s != tc.cs.SHA256 {
				t.Fatalf("sha = %q", s)
			}
		})
	}
}

func TestFromChangeSetCopiesProviderFacts(t *testing.T) {
	cs := providers.ChangeSet{Provider: "salesforce", Facts: map[string]any{
		"salesforce.apex_classes": []string{"A"}, "salesforce.counts": map[string]int{"ApexClass": 1}}}
	d := FromChangeSet(cs)
	if got, ok := d.Strings("salesforce.apex_classes"); !ok || len(got) != 1 || got[0] != "A" {
		t.Fatalf("apex_classes = %v, %v", got, ok)
	}
	// The copy is a separate map.
	d.Set("salesforce.extra", 1)
	if _, ok := cs.Facts["salesforce.extra"]; ok {
		t.Fatal("document shares the change set map")
	}
}

func TestFromChangeSetThroughJSON(t *testing.T) {
	cs := providers.ChangeSet{Provider: "salesforce", Facts: map[string]any{
		"salesforce.destructive_count": 4, "salesforce.apex_classes": []string{"A", "B"}}}
	data, _ := json.Marshal(cs)
	var back providers.ChangeSet
	if err := json.Unmarshal(data, &back); err != nil {
		t.Fatal(err)
	}
	d := FromChangeSet(back)
	if n, _ := d.Int(ChangeDestructiveCount); n != 4 {
		t.Fatalf("count = %d", n)
	}
	if got, ok := d.Strings("salesforce.apex_classes"); !ok || len(got) != 2 {
		t.Fatalf("classes = %v, %v", got, ok)
	}
}

func TestFromResult(t *testing.T) {
	r := run.Result{Facts: map[string]any{"salesforce.deploy_id": "0Af000", "salesforce.tests_failed": 0}}
	d := FromResult(r)
	if s, ok := d.String("salesforce.deploy_id"); !ok || s != "0Af000" {
		t.Fatalf("deploy_id = %q, %v", s, ok)
	}
	if n, ok := d.Int("salesforce.tests_failed"); !ok || n != 0 {
		t.Fatalf("tests_failed = %d, %v", n, ok)
	}
	if got := FromResult(run.Result{}); len(got.Facts) != 0 || got.Schema != Schema {
		t.Fatalf("empty result gave %+v", got)
	}
}

// Synthetic: a plan file that names gate-owned facts and a provider count lower than its
// destructive manifests.
func TestFromChangeSetDropsGateOwnedFacts(t *testing.T) {
	cs := providers.ChangeSet{Provider: "salesforce", SHA256: "abc", Destructive: []string{"a.xml", "b.xml"},
		Facts: map[string]any{
			"salesforce.destructive_count": 0, "change.destructive_count": 0, "change.plan_sha256": "forged",
			"ci.pr_author": "someone-else", "target.tier": "dev", "salesforce.apex_classes": []string{"A"},
		}}
	d := FromChangeSet(cs)
	if n, _ := d.Int(ChangeDestructiveCount); n != 2 {
		t.Fatalf("destructive_count = %d, want the 2 manifests", n)
	}
	if s, _ := d.String(ChangePlanSHA256); s != "abc" {
		t.Fatalf("plan_sha256 = %q", s)
	}
	for _, k := range []string{"ci.pr_author", "target.tier"} {
		if _, ok := d.Get(k); ok {
			t.Fatalf("%s was copied from the plan", k)
		}
	}
	if _, ok := d.Strings("salesforce.apex_classes"); !ok {
		t.Fatal("a provider fact was dropped")
	}
}

func TestMergeOnly(t *testing.T) {
	m := map[string]any{
		"validate.plan_sha256": "a", "test.org_coverage": 80, "salesforce.deploy_id": "0Af",
		"ci.pr_author": "x", "ci.approvals": []any{}, "target.id": "y", "change.empty": true,
		"other.key": 1, "exec.thing": 2,
	}
	d := New()
	d.MergeOnly(ResultPrefixes("salesforce"), m)
	for _, k := range []string{"validate.plan_sha256", "test.org_coverage", "salesforce.deploy_id"} {
		if _, ok := d.Get(k); !ok {
			t.Errorf("%s was not copied", k)
		}
	}
	for _, k := range []string{"ci.pr_author", "ci.approvals", "target.id", "change.empty", "other.key", "exec.thing"} {
		if _, ok := d.Get(k); ok {
			t.Errorf("%s was copied", k)
		}
	}
	// A gate-owned prefix in the allowlist is still refused.
	d = New()
	d.MergeOnly([]string{"ci.", ""}, m)
	if len(d.Facts) != 0 {
		t.Fatalf("copied %v", d.Facts)
	}
	if got := ResultPrefixes(""); len(got) != 5 {
		t.Fatalf("no provider gives %v", got)
	}
}
