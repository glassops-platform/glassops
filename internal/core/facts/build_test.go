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
