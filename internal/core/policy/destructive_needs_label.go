package policy

import "slices"

// DestructiveNeedsLabel requires a PR label before destructive changes reach the listed tiers.
type DestructiveNeedsLabel struct{}

// Name returns the rule name.
func (DestructiveNeedsLabel) Name() string { return "destructive_needs_label" }

// Evaluate checks change.destructive_count against ci.labels.
func (DestructiveNeedsLabel) Evaluate(in Input) []Result {
	const rule = "destructive_needs_label"
	if in.Config == nil || in.Config.Policy.Rules.DestructiveNeedsLabel == nil {
		return one(rule, Skipped, "policy.rules.destructive_needs_label is not set")
	}
	cfg := in.Config.Policy.Rules.DestructiveNeedsLabel
	tier, ok := tierOf(in)
	if !ok {
		return one(rule, Fail, "the facts do not name the target tier")
	}
	if !inTiers(cfg.Tiers, tier) {
		return one(rule, Skipped, "tier %s is not covered by this rule", tier)
	}
	n, ok := in.Facts.Int("change.destructive_count")
	if !ok {
		return one(rule, Fail, "the facts do not say how many destructive changes the plan has")
	}
	if n == 0 {
		return one(rule, Pass, "the plan has no destructive changes")
	}
	if !prPresent(in) {
		return one(rule, Fail, "the plan has %d destructive change(s) and no pull request was found to carry the %q label", n, cfg.Label)
	}
	labels, _ := in.Facts.Strings("ci.labels")
	if slices.Contains(labels, cfg.Label) {
		return one(rule, Pass, "the plan has %d destructive change(s) and the pull request has the %q label", n, cfg.Label)
	}
	return one(rule, Fail, "the plan has %d destructive change(s) and the pull request lacks the %q label", n, cfg.Label)
}
