package facts

import (
	"github.com/glassops-platform/glassops/internal/core/run"
	"github.com/glassops-platform/glassops/internal/providers"
)

// FromChangeSet returns a document holding the change.* facts of a plan and a copy of the
// provider facts it carries. change.destructive_count is the provider's own
// "<provider>.destructive_count" when the plan has one, else the number of destructive
// manifests, so a plan that names destructive files never reads as zero.
func FromChangeSet(cs providers.ChangeSet) *Document {
	d := New()
	d.Set(ChangePlanSHA256, cs.SHA256)
	d.Set(ChangeEmpty, cs.Empty)
	count := len(cs.Destructive)
	if cs.Provider != "" {
		if n, ok := wholeNumber(cs.Facts[cs.Provider+".destructive_count"]); ok {
			count = n
		}
	}
	d.Set(ChangeDestructiveCount, count)
	d.Merge("", cs.Facts)
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
