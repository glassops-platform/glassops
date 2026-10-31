package policy

import (
	"testing"

	"github.com/glassops-platform/glassops/internal/core/config"
	"github.com/glassops-platform/glassops/internal/core/facts"
)

func sodCfg(approval *config.SoDApproval, tiers ...string) *config.Config {
	return cfgWith(func(c *config.Config) {
		c.Policy.Rules.SeparationOfDuties = &config.SoDRule{Tiers: tiers, Approval: approval}
	})
}

func TestSeparationOfDuties(t *testing.T) {
	envApproval := &config.SoDApproval{Environment: "prod-deploy"}
	labelApproval := &config.SoDApproval{Label: "second-pair-of-eyes", Approvers: []string{"Reviewer1", "reviewer2"}}
	prod := func(kv map[string]any) map[string]any {
		kv["target.tier"] = "prod"
		return kv
	}
	tests := []struct {
		name      string
		cfg       *config.Config
		facts     map[string]any
		want      Outcome
		mechanism string
		approver  string
	}{
		{"rule absent", &config.Config{}, prod(map[string]any{}), Skipped, "", ""},
		{"tier not listed is waived", sodCfg(nil, "prod"), map[string]any{"target.tier": "qa", "ci.pr_author": "a", "ci.actor": "a"}, WaivedByConfig, "", ""},
		{"empty tiers waives everything", sodCfg(nil), prod(map[string]any{"ci.pr_author": "a", "ci.actor": "a"}), WaivedByConfig, "", ""},
		{"enforce: different people pass", sodCfg(nil, "prod"), prod(map[string]any{"ci.pr_author": "alice", "ci.actor": "bob"}), Pass, "", ""},
		{"enforce: same person fails", sodCfg(nil, "prod"), prod(map[string]any{"ci.pr_author": "alice", "ci.actor": "alice"}), Fail, "", ""},
		{"enforce: logins compare without case", sodCfg(nil, "prod"), prod(map[string]any{"ci.pr_author": "Alice", "ci.actor": "alice"}), Fail, "", ""},
		{"enforce: author missing fails", sodCfg(nil, "prod"), prod(map[string]any{"ci.actor": "bob"}), Fail, "", ""},
		{"enforce: actor missing fails", sodCfg(nil, "prod"), prod(map[string]any{"ci.pr_author": "alice"}), Fail, "", ""},
		{"enforce: empty author fails", sodCfg(nil, "prod"), prod(map[string]any{"ci.pr_author": "", "ci.actor": "bob"}), Fail, "", ""},
		{"approval configured, different people pass", sodCfg(envApproval, "prod"), prod(map[string]any{"ci.pr_author": "alice", "ci.actor": "bob"}), Pass, "", ""},
		{"environment approval satisfies", sodCfg(envApproval, "prod"), prod(map[string]any{
			"ci.pr_author": "alice", "ci.actor": "alice",
			"ci.approvals": []facts.Approval{{Mechanism: "environment", Environment: "prod-deploy", State: "approved", Approver: "carol"}}},
		), SatisfiedByApproval, "environment", "carol"},
		{"environment approval through JSON form", sodCfg(envApproval, "prod"), prod(map[string]any{
			"ci.pr_author": "alice", "ci.actor": "alice",
			"ci.approvals": []any{map[string]any{"mechanism": "environment", "environment": "prod-deploy", "state": "approved", "approver": "carol"}}},
		), SatisfiedByApproval, "environment", "carol"},
		{"environment approval rejected", sodCfg(envApproval, "prod"), prod(map[string]any{
			"ci.pr_author": "alice", "ci.actor": "alice",
			"ci.approvals": []facts.Approval{{Mechanism: "environment", Environment: "prod-deploy", State: "rejected", Approver: "carol"}}},
		), Fail, "", ""},
		{"environment approval for another environment", sodCfg(envApproval, "prod"), prod(map[string]any{
			"ci.pr_author": "alice", "ci.actor": "alice",
			"ci.approvals": []facts.Approval{{Mechanism: "environment", Environment: "qa-deploy", State: "approved", Approver: "carol"}}},
		), Fail, "", ""},
		{"label approval satisfies", sodCfg(labelApproval, "prod"), prod(map[string]any{
			"ci.pr_author": "alice", "ci.actor": "alice",
			"ci.approvals": []facts.Approval{{Mechanism: "label", Label: "second-pair-of-eyes", Approver: "reviewer1"}}},
		), SatisfiedByApproval, "label", "reviewer1"},
		{"label by someone not in approvers", sodCfg(labelApproval, "prod"), prod(map[string]any{
			"ci.pr_author": "alice", "ci.actor": "alice",
			"ci.approvals": []facts.Approval{{Mechanism: "label", Label: "second-pair-of-eyes", Approver: "mallory"}}},
		), Fail, "", ""},
		{"wrong label", sodCfg(labelApproval, "prod"), prod(map[string]any{
			"ci.pr_author": "alice", "ci.actor": "alice",
			"ci.approvals": []facts.Approval{{Mechanism: "label", Label: "other", Approver: "reviewer1"}}},
		), Fail, "", ""},
		{"environment evidence does not satisfy a label config", sodCfg(labelApproval, "prod"), prod(map[string]any{
			"ci.pr_author": "alice", "ci.actor": "alice",
			"ci.approvals": []facts.Approval{{Mechanism: "environment", Environment: "prod-deploy", State: "approved", Approver: "reviewer1"}}},
		), Fail, "", ""},
		{"same person and no approvals", sodCfg(envApproval, "prod"), prod(map[string]any{"ci.pr_author": "alice", "ci.actor": "alice"}), Fail, "", ""},
		{"approval configured, author unknown, approval found", sodCfg(envApproval, "prod"), prod(map[string]any{
			"ci.actor":     "bob",
			"ci.approvals": []facts.Approval{{Mechanism: "environment", Environment: "prod-deploy", State: "approved", Approver: "carol"}}},
		), SatisfiedByApproval, "environment", "carol"},
		{"tier fact missing fails", sodCfg(nil, "prod"), map[string]any{"ci.pr_author": "a", "ci.actor": "b"}, Fail, "", ""},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			r := one1(t, SeparationOfDuties{}.Evaluate(Input{Facts: doc(tc.facts), Config: tc.cfg}))
			if r.Outcome != tc.want || r.Rule != "separation_of_duties" || r.Mechanism != tc.mechanism || r.Approver != tc.approver || r.Detail == "" {
				t.Fatalf("got %+v, want %s %s %s", r, tc.want, tc.mechanism, tc.approver)
			}
		})
	}
}
