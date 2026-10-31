package policy

import (
	"reflect"
	"testing"
)

// fixed is a rule that returns the same results every time.
type fixed struct {
	name    string
	results []Result
}

func (f fixed) Name() string            { return f.name }
func (f fixed) Evaluate(Input) []Result { return f.results }

func TestDenies(t *testing.T) {
	for _, o := range []Outcome{Pass, Skipped, WaivedByConfig, SatisfiedByApproval} {
		if o.Denies() {
			t.Fatalf("%s denies", o)
		}
	}
	if !Fail.Denies() {
		t.Fatal("fail does not deny")
	}
}

func TestEvaluate(t *testing.T) {
	pass := fixed{"a", []Result{{Rule: "a", Outcome: Pass}}}
	two := fixed{"b", []Result{{Rule: "b.qa", Outcome: Skipped}, {Rule: "b.prod", Outcome: SatisfiedByApproval, Mechanism: "label", Approver: "octocat"}}}
	fail := fixed{"c", []Result{{Rule: "c", Outcome: Fail, Detail: "no"}}}

	cases := []struct {
		name      string
		rules     []Rule
		wantAllow bool
		wantRules []string
	}{
		{"no rules", nil, true, []string{}},
		{"all pass", []Rule{pass, two}, true, []string{"a", "b.qa", "b.prod"}},
		{"one fail", []Rule{pass, fail, two}, false, []string{"a", "c", "b.qa", "b.prod"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			allow, results := Evaluate(tc.rules, Input{})
			if allow != tc.wantAllow {
				t.Fatalf("allow = %v", allow)
			}
			got := []string{}
			for _, r := range results {
				got = append(got, r.Rule)
			}
			if !reflect.DeepEqual(got, tc.wantRules) {
				t.Fatalf("rules = %v, want %v", got, tc.wantRules)
			}
		})
	}
}
