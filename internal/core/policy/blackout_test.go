package policy

import (
	"strings"
	"testing"
	"time"

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

func TestBlackoutStart(t *testing.T) {
	cfg := cfgWith(func(c *config.Config) {
		c.Policy.Blackouts = []config.Blackout{
			{From: "2027-12-20", To: "2028-01-02", Tiers: []string{"prod"}},
			{From: "2027-12-10", To: "2027-12-11", Tiers: []string{"qa"}},
			{From: "2027-11-01", To: "2027-11-03"},
			{From: "not a date", To: "2999-01-01"},
		}
	})
	tests := []struct {
		name string
		cfg  *config.Config
		tier string
		now  string
		want string // "" for none
	}{
		{"the next blackout for the tier", cfg, "prod", "2027-12-01T12:00:00Z", "2027-12-20T00:00:00Z"},
		{"the earliest of several", cfg, "qa", "2027-12-01T12:00:00Z", "2027-12-10T00:00:00Z"},
		{"a blackout for no tier covers every tier", cfg, "uat", "2027-10-30T08:00:00Z", "2027-11-01T00:00:00Z"},
		{"one that has begun gives now", cfg, "prod", "2027-12-25T08:30:00Z", "2027-12-25T08:30:00Z"},
		{"the last day counts", cfg, "prod", "2028-01-02T23:59:59Z", "2028-01-02T23:59:59Z"},
		{"the day after the end", cfg, "prod", "2028-01-03T00:00:00Z", ""},
		{"other tier", cfg, "uat", "2027-12-01T12:00:00Z", ""},
		{"the UTC date is used", cfg, "prod", "2027-12-19T20:00:00-05:00", "2027-12-20T01:00:00Z"},
		{"an unknown tier is covered by every blackout", cfg, "", "2027-12-01T12:00:00Z", "2027-12-10T00:00:00Z"},
		{"the start is not before now", cfg, "prod", "2027-12-20T00:00:00Z", "2027-12-20T00:00:00Z"},
		{"no config", nil, "prod", "2027-12-01T12:00:00Z", ""},
		{"no blackouts", &config.Config{}, "prod", "2027-12-01T12:00:00Z", ""},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, ok := BlackoutStart(tc.cfg, tc.tier, at(t, tc.now))
			if tc.want == "" {
				if ok || !got.IsZero() {
					t.Fatalf("got %s, %v, want none", got, ok)
				}
				return
			}
			if !ok || !got.Equal(at(t, tc.want)) || got.Location() != time.UTC {
				t.Fatalf("got %s (%v), want %s", got, ok, tc.want)
			}
		})
	}
}
