package policy

import (
	"fmt"
	"strings"
	"time"
)

// Blackout refuses changes to a tier during a configured date range. Dates are compared in
// UTC, because the config gives blackouts no time zone.
type Blackout struct{}

// Name returns the rule name.
func (Blackout) Name() string { return "blackout" }

// Evaluate checks every blackout against today's UTC date.
func (Blackout) Evaluate(in Input) []Result {
	if in.Config == nil || len(in.Config.Policy.Blackouts) == 0 {
		return one("blackout", Skipped, "no blackouts are configured")
	}
	tier, tierKnown := tierOf(in)
	today := in.Now.UTC().Format(time.DateOnly)
	var hits []string
	for _, b := range in.Config.Policy.Blackouts {
		// Dates are YYYY-MM-DD, so string order is date order.
		if today < b.From || today > b.To {
			continue
		}
		if tierKnown && !inTiers(b.Tiers, tier) {
			continue
		}
		hit := fmt.Sprintf("%s to %s", b.From, b.To)
		if b.Reason != "" {
			hit += " (" + b.Reason + ")"
		}
		hits = append(hits, hit)
	}
	if len(hits) == 0 {
		return one("blackout", Pass, "no blackout covers %s for this target", today)
	}
	detail := fmt.Sprintf("blackout in force on %s: %s", today, strings.Join(hits, ", "))
	if !tierKnown {
		detail += " (the facts do not name the target tier, so it is treated as covered)"
	}
	return one("blackout", Fail, "%s", detail)
}
