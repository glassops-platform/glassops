package facts

import (
	"strings"

	"github.com/glassops-platform/glassops/internal/core/run"
	"github.com/glassops-platform/glassops/internal/providers"
)

// GateOwnedPrefixes are the fact prefixes only gate derives: the change set, the target and
// the CI context. A value under them from a plan file or an earlier result is never copied,
// because those files are not signed and gate would otherwise read an attacker's value where
// it has none of its own.
var GateOwnedPrefixes = []string{"change.", "ci.", "target."}

// resultPrefixes are the step prefixes gate copies from an earlier step's result, the ones the
// rules read. ResultPrefixes adds the provider's own prefix.
var resultPrefixes = []string{"validate.", "plan.", "analyze.", "test.", "check."}

// ResultPrefixes is the one allowlist of fact prefixes gate copies from earlier step results:
// validate., plan., analyze., test., check. and the provider's own prefix, such as
// "salesforce.". Nothing else, and never a gate-owned prefix, is copied.
func ResultPrefixes(provider string) []string {
	out := append([]string(nil), resultPrefixes...)
	if provider != "" {
		out = append(out, provider+".")
	}
	return out
}

// gateOwned reports whether key is under a gate-owned prefix.
func gateOwned(key string) bool {
	for _, p := range GateOwnedPrefixes {
		if strings.HasPrefix(key, p) {
			return true
		}
	}
	return false
}

// MergeOnly copies every key in m that starts with one of prefixes and is not under a
// gate-owned prefix. An empty prefix list copies nothing, and an empty prefix matches nothing.
func (d *Document) MergeOnly(prefixes []string, m map[string]any) {
	for k, v := range m {
		if gateOwned(k) {
			continue
		}
		for _, p := range prefixes {
			if p != "" && strings.HasPrefix(k, p) {
				d.Set(k, v)
				break
			}
		}
	}
}

// FromChangeSet returns a document holding a copy of the facts a plan carries and the
// change.* facts derived from the plan. The plan's own change., ci. and target. facts are
// dropped, and change.* is set after the copy, so the plan file cannot speak for gate.
// change.destructive_count is the larger of the provider's own
// "<provider>.destructive_count" and the number of destructive manifests, so a plan that
// names destructive files never reads as zero.
func FromChangeSet(cs providers.ChangeSet) *Document {
	d := New()
	for k, v := range cs.Facts {
		if !gateOwned(k) {
			d.Set(k, v)
		}
	}
	count := len(cs.Destructive)
	if cs.Provider != "" {
		if n, ok := wholeNumber(cs.Facts[cs.Provider+".destructive_count"]); ok && n > count {
			count = n
		}
	}
	d.Set(ChangePlanSHA256, cs.SHA256)
	d.Set(ChangeEmpty, cs.Empty)
	d.Set(ChangeDestructiveCount, count)
	return d
}

// FromResult returns a document holding a copy of the facts of a step result.
func FromResult(r run.Result) *Document {
	d := New()
	d.Merge("", r.Facts)
	return d
}

func wholeNumber(v any) (int, bool) {
	d := New()
	d.Set("k", v)
	return d.Int("k")
}
