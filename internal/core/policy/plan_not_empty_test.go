package policy

import "testing"

func TestPlanNotEmpty(t *testing.T) {
	tests := []struct {
		name  string
		facts map[string]any
		want  Outcome
	}{
		{"has changes", map[string]any{"change.empty": false}, Pass},
		{"empty", map[string]any{"change.empty": true}, Fail},
		{"missing key", map[string]any{}, Fail},
		{"wrong type", map[string]any{"change.empty": "false"}, Fail},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			r := one1(t, PlanNotEmpty{}.Evaluate(Input{Facts: doc(tc.facts)}))
			if r.Outcome != tc.want || r.Rule != "plan_not_empty" || r.Detail == "" {
				t.Fatalf("got %+v, want %s", r, tc.want)
			}
		})
	}
}
