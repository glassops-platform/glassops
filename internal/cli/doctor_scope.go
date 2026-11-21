package cli

import (
	"fmt"
	"sort"
	"strings"

	"github.com/glassops-platform/glassops/internal/core/config"
	"github.com/glassops-platform/glassops/internal/core/run"
)

func init() {
	doctorExtras = append(doctorExtras, doctorPlanScope)
}

// doctorPlanScope prints the plan scope each target gets and where it comes from, so a target
// that plans the whole source, or one that does not, is visible before a run. It covers the
// names under targets and under environments, and is silent when the config names neither.
func doctorPlanScope(_ *App, cfg *config.Config) []run.Check {
	seen := map[string]bool{}
	var names []string
	for name := range cfg.Targets {
		seen[name] = true
		names = append(names, name)
	}
	for name := range cfg.Environments {
		if !seen[name] {
			names = append(names, name)
		}
	}
	if len(names) == 0 {
		return nil
	}
	sort.Strings(names)
	parts := make([]string, 0, len(names))
	for _, name := range names {
		scope, source := cfg.PlanScope(name)
		parts = append(parts, fmt.Sprintf("%s %s (%s)", name, scope, source))
	}
	return []run.Check{{Name: "plan.scope", Status: run.StatusPass, Detail: strings.Join(parts, ", ")}}
}
