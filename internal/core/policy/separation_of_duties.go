package policy

import (
	"slices"
	"strings"
)

// SeparationOfDuties requires that the person who deploys is not the author of the change,
// unless a configured approval stands in for it.
type SeparationOfDuties struct{}

// Name returns the rule name.
func (SeparationOfDuties) Name() string { return "separation_of_duties" }

// Evaluate applies the rule for the target's tier.
func (SeparationOfDuties) Evaluate(in Input) []Result {
	const rule = "separation_of_duties"
	if in.Config == nil || in.Config.Policy.Rules.SeparationOfDuties == nil {
		return one(rule, Skipped, "policy.rules.separation_of_duties is not set")
	}
	cfg := in.Config.Policy.Rules.SeparationOfDuties
	tier, ok := tierOf(in)
	if !ok {
		return one(rule, Fail, "the facts do not name the target tier")
	}
	if !slices.Contains(cfg.Tiers, tier) {
		return one(rule, WaivedByConfig, "tier %s is not in policy.rules.separation_of_duties.tiers", tier)
	}
	author, hasAuthor := in.Facts.String("ci.pr_author")
	actor, hasActor := in.Facts.String("ci.actor")
	hasAuthor = hasAuthor && author != ""
	hasActor = hasActor && actor != ""
	differ := hasAuthor && hasActor && !strings.EqualFold(author, actor)

	if cfg.Approval == nil {
		switch {
		case !hasAuthor || !hasActor:
			return one(rule, Fail, "enforced for tier %s but the pull request author or the deploying actor is unknown", tier)
		case !differ:
			return one(rule, Fail, "enforced for tier %s: %s authored the change and is also deploying it", tier, actor)
		}
		return one(rule, Pass, "enforced for tier %s: author %s and deployer %s differ", tier, author, actor)
	}

	if differ {
		return one(rule, Pass, "enforced for tier %s: author %s and deployer %s differ", tier, author, actor)
	}
	ap := cfg.Approval
	for _, a := range in.Facts.Approvals() {
		switch {
		case a.Mechanism == "environment" && ap.Environment != "" &&
			a.Environment == ap.Environment && a.State == "approved":
			return approved(rule, tier, a.Mechanism, a.Approver)
		case a.Mechanism == "label" && ap.Label != "" && a.Label == ap.Label &&
			inFold(ap.Approvers, a.Approver):
			return approved(rule, tier, a.Mechanism, a.Approver)
		}
	}
	return one(rule, Fail, "tier %s: author and deployer are the same or unknown, and no configured approval was found", tier)
}

func approved(rule, tier, mechanism, approver string) []Result {
	return []Result{{
		Rule: rule, Outcome: SatisfiedByApproval, Mechanism: mechanism, Approver: approver,
		Detail: "tier " + tier + ": satisfied by " + mechanism + " approval from " + approver,
	}}
}

func inFold(list []string, s string) bool {
	return slices.ContainsFunc(list, func(x string) bool { return strings.EqualFold(x, s) })
}
