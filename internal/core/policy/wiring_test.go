package policy

import (
	"testing"

	"github.com/glassops-platform/glassops/internal/core/config"
)

// The two config fallbacks the rules read once the Phase 4 config sections exist.

func TestOrgMinReadsTestCoverage(t *testing.T) {
	if got := orgMin(Input{}); got != nil {
		t.Fatalf("no config: got %v", *got)
	}
	c := cfgWith(func(c *config.Config) { c.Test.Coverage.OrgMin = intp(80) })
	got := orgMin(Input{Config: c})
	if got == nil || *got != 80 {
		t.Fatalf("got %v, want 80", got)
	}
}

func TestAnalyzeFallbackReadsAnalyzeMaxSeverity(t *testing.T) {
	if got := analyzeFallback(Input{}); got != nil {
		t.Fatalf("no config: got %v", *got)
	}
	c := cfgWith(func(c *config.Config) { c.Analyze.MaxSeverity = intp(2) })
	got := analyzeFallback(Input{Config: c})
	if got == nil || *got != 2 {
		t.Fatalf("got %v, want 2", got)
	}
}
