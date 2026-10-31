package policy

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/glassops-platform/glassops/internal/core/config"
)

// stubEvaluator returns fixed results and an optional error.
type stubEvaluator struct {
	name    string
	results []Result
	err     error
}

func (s stubEvaluator) Name() string { return s.name }
func (s stubEvaluator) Evaluate(context.Context, Input) ([]Result, error) {
	return s.results, s.err
}

func TestEvaluators(t *testing.T) {
	rules := Default()
	evs := Evaluators(rules, &config.Config{})
	if len(evs) != 1 || evs[0].Name() != "go" {
		t.Fatalf("no rego: %v", evs)
	}
	if !reflect.DeepEqual(evs[0].(GoRules).Rules, rules) {
		t.Fatal("the Go evaluator does not hold the rules it was given")
	}
	evs = Evaluators(rules, &config.Config{Policy: config.Policy{Rego: config.Rego{Files: []string{"p.rego"}}}})
	if len(evs) != 2 || evs[0].Name() != "go" || evs[1].Name() != "rego" {
		t.Fatalf("with rego: %v", evs)
	}
	if evs := Evaluators(rules, nil); len(evs) != 1 {
		t.Fatalf("nil config: %v", evs)
	}
}

func TestGoRulesMatchesEvaluate(t *testing.T) {
	rules := []Rule{
		fixed{"a", []Result{{Rule: "a", Outcome: Pass}}},
		fixed{"b", []Result{{Rule: "b", Outcome: Fail}}},
	}
	_, want := Evaluate(rules, Input{})
	got, err := GoRules{Rules: rules}.Evaluate(context.Background(), Input{})
	if err != nil || !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v, %v", got, err)
	}
}

func TestEvaluateAll(t *testing.T) {
	pass := stubEvaluator{name: "go", results: []Result{{Rule: "a", Outcome: Pass}}}
	deny := stubEvaluator{name: "go", results: []Result{{Rule: "a", Outcome: Fail}}}
	broken := stubEvaluator{name: "rego", err: errors.New("opa missing")}
	cases := []struct {
		name      string
		evs       []Evaluator
		wantAllow bool
		wantRules []string
		wantErr   bool
	}{
		{"none", nil, true, []string{}, false},
		{"pass", []Evaluator{pass}, true, []string{"a"}, false},
		{"deny", []Evaluator{deny}, false, []string{"a"}, false},
		{"broken fails closed", []Evaluator{pass, broken}, false, []string{"a", "rego"}, true},
		{"broken first, the rest still run", []Evaluator{broken, pass}, false, []string{"rego", "a"}, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			allow, results, err := EvaluateAll(context.Background(), tc.evs, Input{})
			if allow != tc.wantAllow || (err != nil) != tc.wantErr {
				t.Fatalf("allow %v, err %v", allow, err)
			}
			got := []string{}
			for _, r := range results {
				got = append(got, r.Rule)
			}
			if !reflect.DeepEqual(got, tc.wantRules) {
				t.Fatalf("rules = %v", got)
			}
		})
	}
}

func TestRegoStubFailsClosed(t *testing.T) {
	_, err := NewRego(config.Rego{Files: []string{"p.rego"}}).Evaluate(context.Background(), Input{})
	if err == nil {
		t.Fatal("the rego stub returned no error")
	}
}
