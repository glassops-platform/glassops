package policy

import (
	"strings"
	"testing"

	"github.com/glassops-platform/glassops/internal/core/config"
)

func TestBlackout(t *testing.T) {
	freeze := config.Blackout{From: "2027-12-20", To: "2028-01-02", Tiers: []string{"prod"}, Reason: "holiday freeze"}
	all := config.Blackout{From: "2027-06-01", To: "2027-06-01", Reason: "audit day"}
	cfg := cfgWith(func(c *config.Config) { c.Policy.Blackouts = []config.Blackout{freeze, all} })
	tests := []struct {
		name string
		cfg  *config.Config
		tier string
		now  string
		want Outcome
		in   string
	}{
		{"first day is inside", cfg, "prod", "2027-12-20T00:00:00Z", Fail, "holiday freeze"},
		{"last day is inside", cfg, "prod", "2028-01-02T23:59:59Z", Fail, "holiday freeze"},
		{"day before", cfg, "prod", "2027-12-19T23:59:59Z", Pass, ""},
		{"day after", cfg, "prod", "2028-01-03T00:00:00Z", Pass, ""},
		{"other tier is not covered", cfg, "qa", "2027-12-25T12:00:00Z", Pass, ""},
		{"no tiers means all tiers", cfg, "qa", "2027-06-01T12:00:00Z", Fail, "audit day"},
		{"single day range", cfg, "prod", "2027-06-01T23:00:00Z", Fail, "audit day"},
		{"no blackouts", &config.Config{}, "prod", "2027-12-25T12:00:00Z", Skipped, ""},
		// 2027-12-19T20:00-05:00 is 2027-12-20T01:00Z, so the UTC date is inside the freeze.
		{"UTC date is used", cfg, "prod", "2027-12-19T20:00:00-05:00", Fail, "holiday freeze"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			in := Input{Facts: doc(map[string]any{"target.tier": tc.tier}), Config: tc.cfg, Now: at(t, tc.now)}
			r := one1(t, Blackout{}.Evaluate(in))
			if r.Outcome != tc.want || r.Rule != "blackout" || !strings.Contains(r.Detail, tc.in) {
				t.Fatalf("got %+v, want %s containing %q", r, tc.want, tc.in)
			}
		})
	}
}

func TestBlackoutUnknownTierIsTreatedAsCovered(t *testing.T) {
	cfg := cfgWith(func(c *config.Config) {
		c.Policy.Blackouts = []config.Blackout{{From: "2027-12-20", To: "2028-01-02", Tiers: []string{"prod"}}}
	})
	r := one1(t, Blackout{}.Evaluate(Input{Facts: doc(nil), Config: cfg, Now: at(t, "2027-12-25T00:00:00Z")}))
	if r.Outcome != Fail {
		t.Fatalf("got %+v", r)
	}
	r = one1(t, Blackout{}.Evaluate(Input{Facts: doc(nil), Config: cfg, Now: at(t, "2027-11-25T00:00:00Z")}))
	if r.Outcome != Pass {
		t.Fatalf("outside every range: %+v", r)
	}
}
