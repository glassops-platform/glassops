// Package guard checks that a logged-in target is the one the config names, before any step
// does anything to it.
package guard

import (
	"fmt"

	"github.com/glassops-platform/glassops/internal/providers"
)

// Tiers with a rule about the kind of target. Every other tier checks the Id only.
const (
	TierProd      = "prod"
	TierDevHub    = "devhub"
	TierEphemeral = "ephemeral"
)

// MismatchError means the target is not the one configured. Commands exit 2 on it.
type MismatchError struct {
	Target string
	// Want and Got are the configured and reported Ids.
	Want string
	Got  string
	// Reason says what did not match.
	Reason string
}

func (e *MismatchError) Error() string {
	return fmt.Sprintf("guard: target %s: %s (configured %s, logged in to %s)", e.Target, e.Reason, e.Want, e.Got)
}

// Check compares what the target reports with its config. canon gives the form two Ids are
// compared by, such as the provider's CanonicalID.
func Check(want providers.Target, got providers.TargetInfo, canon func(string) string) error {
	mismatch := func(reason string) error {
		return &MismatchError{Target: want.Name, Want: want.ID, Got: got.ID, Reason: reason}
	}
	if want.ID == "" {
		return mismatch("no Id is configured, so the target cannot be checked")
	}
	if got.ID == "" || canon(want.ID) != canon(got.ID) {
		return mismatch("Id does not match")
	}
	if need, ok := kindFor(want.Tier); ok && got.Kind != need {
		return mismatch(fmt.Sprintf("tier %s needs a %s target, this one is %s", want.Tier, need, got.Kind))
	}
	return nil
}

// kindFor returns the kind a tier requires, if it has a rule.
func kindFor(tier string) (providers.Kind, bool) {
	switch tier {
	case TierProd, TierDevHub:
		return providers.KindProduction, true
	case TierEphemeral:
		return providers.KindScratch, true
	}
	return "", false
}
