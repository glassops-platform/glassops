package config

import (
	"fmt"
	"regexp"
	"sort"
	"strings"
	"time"
	// Embed the time zone database so window validation does not depend on the host,
	// for example a minimal container image without tzdata.
	_ "time/tzdata"
)

var (
	orgIDPattern  = regexp.MustCompile(`^00D[a-zA-Z0-9]{12}([a-zA-Z0-9]{3})?$`)
	envRefPattern = regexp.MustCompile(`^\$([A-Z_][A-Z0-9_]*)$`)
)

// ProdTier is the tier name the design gives special treatment: at most one target may use it.
const ProdTier = "prod"

// Check applies the rules the schema cannot express and returns one line per problem.
func Check(c *Config) []string {
	var problems []string
	add := func(format string, args ...any) { problems = append(problems, fmt.Sprintf(format, args...)) }

	names := sortedKeys(c.Targets)
	idOwner := map[string]string{}
	var prodTargets []string
	for _, name := range names {
		t := c.Targets[name]
		if t.ID != "" && !envRefPattern.MatchString(t.ID) {
			if !orgIDPattern.MatchString(t.ID) {
				add("targets.%s.id: %q does not look like an org Id", name, t.ID)
			} else {
				key := canonicalOrgID(t.ID)
				if other, ok := idOwner[key]; ok {
					add("targets.%s.id: same org Id as targets.%s", name, other)
				} else {
					idOwner[key] = name
				}
			}
		}
		if t.Tier == ProdTier {
			prodTargets = append(prodTargets, name)
		}
	}
	if len(prodTargets) > 1 {
		add("targets: at most one target may have tier prod, found %s", strings.Join(prodTargets, ", "))
	}

	for _, tier := range sortedKeys(c.Policy.Windows) {
		for i, w := range c.Policy.Windows[tier] {
			if _, err := time.LoadLocation(w.TZ); err != nil {
				add("policy.windows.%s[%d].tz: unknown time zone %q", tier, i, w.TZ)
			}
			if w.From == w.To {
				add("policy.windows.%s[%d]: from and to are both %s", tier, i, w.From)
			}
		}
	}

	for i, b := range c.Policy.Blackouts {
		from, errFrom := time.Parse(time.DateOnly, b.From)
		to, errTo := time.Parse(time.DateOnly, b.To)
		if errFrom != nil {
			add("policy.blackouts[%d].from: %q is not a valid date", i, b.From)
		}
		if errTo != nil {
			add("policy.blackouts[%d].to: %q is not a valid date", i, b.To)
		}
		if errFrom == nil && errTo == nil && to.Before(from) {
			add("policy.blackouts[%d]: to (%s) is before from (%s)", i, b.To, b.From)
		}
	}

	if c.Policy.PermitTTL != "" {
		d, err := time.ParseDuration(c.Policy.PermitTTL)
		switch {
		case err != nil:
			add("policy.permit_ttl: %q is not a duration, such as 2h or 90m", c.Policy.PermitTTL)
		case d < MinPermitTTL || d > MaxPermitTTL:
			add("policy.permit_ttl: %s is outside %s to %s", c.Policy.PermitTTL, MinPermitTTL, MaxPermitTTL)
		}
	}

	if sod := c.Policy.Rules.SeparationOfDuties; sod != nil && sod.Approval != nil {
		a := sod.Approval
		if a.Environment == "" && a.Label == "" {
			add("policy.rules.separation_of_duties.approval: set environment or label")
		}
		if a.Label != "" && len(a.Approvers) == 0 {
			add("policy.rules.separation_of_duties.approval: label %q needs approvers", a.Label)
		}
	}

	for _, name := range sortedKeys(c.Environments) {
		e := c.Environments[name]
		if _, ok := c.Targets[e.DevHub]; !ok {
			add("environments.%s.devhub: %q is not a configured target", name, e.DevHub)
		}
	}

	for _, name := range sortedKeys(c.Pipelines) {
		for i, s := range c.Pipelines[name] {
			if strings.TrimSpace(s.Run) == "" {
				add("pipelines.%s[%d]: empty step", name, i)
			}
		}
	}
	return problems
}

// canonicalOrgID compares 15- and 18-character Ids by their first 15 characters, which are
// the case-sensitive Id itself. The last 3 are a checksum suffix.
func canonicalOrgID(id string) string {
	if len(id) > 15 {
		return id[:15]
	}
	return id
}

// EnvRefs lists every $VAR the config refers to, sorted and without duplicates.
func EnvRefs(c *Config) []string {
	set := map[string]bool{}
	note := func(v string) {
		if m := envRefPattern.FindStringSubmatch(v); m != nil {
			set[m[1]] = true
		}
	}
	for _, t := range c.Targets {
		for _, v := range []string{t.ID, t.Auth.Username, t.Auth.ClientID, t.Auth.Key, t.Auth.URL, t.Auth.Token} {
			note(v)
		}
	}
	out := make([]string, 0, len(set))
	for k := range set {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

func sortedKeys[V any](m map[string]V) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
