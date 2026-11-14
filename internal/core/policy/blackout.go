package policy

import (
	"fmt"
	"strings"
	"time"

	"github.com/glassops-platform/glassops/internal/core/config"
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

// BlackoutStart returns when the first blackout that covers tier begins, for a blackout that
// has not ended by the UTC date of now. A blackout that has already begun gives now, because
// nothing issued inside it may stand. An empty tier is treated as covered by every blackout,
// as the blackout rule treats a tier the facts do not name. ok is false when no such blackout
// exists. gate caps a PERMIT's expiry at this time, so a PERMIT issued before a blackout
// cannot deploy inside it.
func BlackoutStart(cfg *config.Config, tier string, now time.Time) (start time.Time, ok bool) {
	if cfg == nil {
		return time.Time{}, false
	}
	today := now.UTC().Format(time.DateOnly)
	for _, b := range cfg.Policy.Blackouts {
		if today > b.To || (tier != "" && !inTiers(b.Tiers, tier)) {
			continue
		}
		// The config checks dates when it loads, and a date that does not parse is not a
		// blackout this function can place.
		from, err := time.Parse(time.DateOnly, b.From)
		if err != nil {
			continue
		}
		if from.Before(now) {
			from = now.UTC()
		}
		if !ok || from.Before(start) {
			start, ok = from, true
		}
	}
	return start, ok
}
