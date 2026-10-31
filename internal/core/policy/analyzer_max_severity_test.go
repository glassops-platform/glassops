package policy

import (
	"strings"
	"testing"

	"github.com/glassops-platform/glassops/internal/core/config"
)

func analyzerCfg(max *int) *config.Config {
	return cfgWith(func(c *config.Config) { c.Policy.Rules.AnalyzerMaxSeverity = max })
}

func TestAnalyzerMaxSeverity(t *testing.T) {
	counts := func(s1, s2, s3, s4, s5 int) map[string]any {
		return map[string]any{"analyze.sev1": s1, "analyze.sev2": s2, "analyze.sev3": s3, "analyze.sev4": s4, "analyze.sev5": s5}
	}
	tests := []struct {
		name   string
		cfg    *config.Config
		facts  map[string]any
		want   Outcome
		detail string
	}{
		{"no limit set", analyzerCfg(nil), counts(1, 1, 1, 1, 1), Skipped, "no analyzer severity limit"},
		{"facts absent", analyzerCfg(intp(2)), map[string]any{}, Skipped, "did not run"},
		{"clean", analyzerCfg(intp(2)), counts(0, 0, 0, 0, 0), Pass, "no violations"},
		{"only milder than the limit", analyzerCfg(intp(2)), counts(0, 0, 5, 9, 3), Pass, "severity 2"},
		{"at the limit fails", analyzerCfg(intp(2)), counts(0, 1, 0, 0, 0), Fail, "1 at severity 2"},
		{"worse than the limit fails", analyzerCfg(intp(2)), counts(2, 0, 0, 0, 0), Fail, "2 at severity 1"},
		{"both reported", analyzerCfg(intp(3)), counts(1, 0, 4, 0, 0), Fail, "1 at severity 1, 4 at severity 3"},
		{"limit 5 refuses anything", analyzerCfg(intp(5)), counts(0, 0, 0, 0, 1), Fail, "severity 5"},
		{"limit 1 only refuses severity 1", analyzerCfg(intp(1)), counts(0, 7, 7, 7, 7), Pass, "severity 1"},
		{"counts through JSON types", analyzerCfg(intp(2)), map[string]any{"analyze.sev1": float64(0), "analyze.sev2": float64(3)}, Fail, "3 at severity 2"},
		{"partial facts count as present", analyzerCfg(intp(2)), map[string]any{"analyze.sev3": 1}, Pass, "no violations"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			r := one1(t, AnalyzerMaxSeverity{}.Evaluate(Input{Facts: doc(tc.facts), Config: tc.cfg}))
			if r.Outcome != tc.want || r.Rule != "analyzer_max_severity" || !strings.Contains(r.Detail, tc.detail) {
				t.Fatalf("got %+v, want %s containing %q", r, tc.want, tc.detail)
			}
		})
	}
}
