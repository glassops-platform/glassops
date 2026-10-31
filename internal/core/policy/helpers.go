package policy

import (
	"fmt"
	"strings"
)

// tierOf returns the target tier from the facts.
func tierOf(in Input) (string, bool) {
	if in.Facts == nil {
		return "", false
	}
	t, ok := in.Facts.String("target.tier")
	return t, ok && t != ""
}

// inTiers reports whether tier is listed. An empty list means every tier.
func inTiers(list []string, tier string) bool {
	if len(list) == 0 {
		return true
	}
	for _, t := range list {
		if t == tier {
			return true
		}
	}
	return false
}

// prPresent reports whether the facts name a pull request. The number may be a string or a
// number, depending on how the CI layer filled it.
func prPresent(in Input) bool {
	if in.Facts == nil {
		return false
	}
	if s, ok := in.Facts.String("ci.pr_number"); ok {
		return strings.TrimSpace(s) != ""
	}
	if n, ok := in.Facts.Int("ci.pr_number"); ok {
		return n > 0
	}
	return false
}

func one(rule string, o Outcome, format string, args ...any) []Result {
	return []Result{{Rule: rule, Outcome: o, Detail: fmt.Sprintf(format, args...)}}
}
