package policy

import (
	"strings"
	"testing"

	"github.com/glassops-platform/glassops/internal/core/config"
)

func destructiveCfg(tiers ...string) *config.Config {
	return cfgWith(func(c *config.Config) {
		c.Policy.Rules.DestructiveNeedsLabel = &config.DestructiveLabelRule{Label: "destructive-ok", Tiers: tiers}
	})
}

func TestDestructiveNeedsLabel(t *testing.T) {
	tests := []struct {
		name   string
		cfg    *config.Config
		facts  map[string]any
		want   Outcome
		detail string
	}{
		{"rule absent", &config.Config{}, map[string]any{"target.tier": "prod"}, Skipped, "not set"},
		{"tier not listed", destructiveCfg("prod"), map[string]any{"target.tier": "qa", "change.destructive_count": 2}, Skipped, "tier qa"},
		{"no destructive changes", destructiveCfg("prod"), map[string]any{"target.tier": "prod", "change.destructive_count": 0}, Pass, "no destructive"},
		{"destructive with label", destructiveCfg("prod"), map[string]any{
			"target.tier": "prod", "change.destructive_count": 2, "ci.pr_number": "17",
			"ci.labels": []string{"bug", "destructive-ok"}}, Pass, "has the"},
		{"destructive without label", destructiveCfg("prod"), map[string]any{
			"target.tier": "prod", "change.destructive_count": 2, "ci.pr_number": "17",
			"ci.labels": []string{"bug"}}, Fail, "lacks"},
		{"destructive, no labels fact", destructiveCfg("prod"), map[string]any{
			"target.tier": "prod", "change.destructive_count": 1, "ci.pr_number": "17"}, Fail, "lacks"},
		{"no PR found", destructiveCfg("prod"), map[string]any{
			"target.tier": "prod", "change.destructive_count": 2, "ci.labels": []string{"destructive-ok"}}, Fail, "no pull request"},
		{"empty PR number", destructiveCfg("prod"), map[string]any{
			"target.tier": "prod", "change.destructive_count": 2, "ci.pr_number": ""}, Fail, "no pull request"},
		{"numeric PR number", destructiveCfg("prod"), map[string]any{
			"target.tier": "prod", "change.destructive_count": 2, "ci.pr_number": 17,
			"ci.labels": []any{"destructive-ok"}}, Pass, "has the"},
		{"zero PR number is no PR", destructiveCfg("prod"), map[string]any{
			"target.tier": "prod", "change.destructive_count": 2, "ci.pr_number": 0}, Fail, "no pull request"},
		{"empty tiers means every tier", destructiveCfg(), map[string]any{
			"target.tier": "qa", "change.destructive_count": 1, "ci.pr_number": "3"}, Fail, "lacks"},
		{"count missing fails", destructiveCfg("prod"), map[string]any{"target.tier": "prod"}, Fail, "how many"},
		{"tier missing fails", destructiveCfg("prod"), map[string]any{"change.destructive_count": 0}, Fail, "tier"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			r := one1(t, DestructiveNeedsLabel{}.Evaluate(Input{Facts: doc(tc.facts), Config: tc.cfg}))
			if r.Outcome != tc.want || r.Rule != "destructive_needs_label" || !strings.Contains(r.Detail, tc.detail) {
				t.Fatalf("got %+v, want %s containing %q", r, tc.want, tc.detail)
			}
		})
	}
}
