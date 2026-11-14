package cli

import (
	"strings"
	"testing"
	"time"

	"github.com/glassops-platform/glassops/internal/core/exitcode"
	"github.com/glassops-platform/glassops/internal/core/policy"
)

const rtBlackoutFromTomorrow = `version: 1
targets:
  qa:
    id: 00D000000000002AAA
    tier: qa
    auth: { method: sfdx-url, url: $SF_QA_AUTH_URL }
policy:
  blackouts:
    - { from: "2026-10-27", to: "2026-10-28", tiers: [qa], reason: freeze }
`

// A PERMIT issued before a blackout starts must expire when the blackout starts, so it is
// never valid inside the blackout. Gate at 23:30 UTC with the default 2h TTL would otherwise
// expire at 01:30, an hour and a half into the blackout.
func TestRTPermitExpiresWhenABlackoutStarts(t *testing.T) {
	at := func(day, h, m int) func() time.Time {
		return func() time.Time { return time.Date(2026, 10, day, h, m, 0, 0, time.UTC) }
	}
	e := rtNew(t, rtBlackoutFromTomorrow)
	rtUseRules(t, policy.Default)
	e.app.Now = at(26, 23, 30)
	e.gate(t, exitcode.OK, "--target", "qa")
	if got := e.readPermit(t).ExpiresAt; !got.Equal(at(27, 0, 0)()) {
		t.Fatalf("expires_at %s, want 2026-10-27T00:00:00Z, the start of the blackout", got.UTC().Format(time.RFC3339))
	}
	r := e.result(t, "gate")
	if r.Facts["gate.expires_at_blackout_start"] != true || r.Facts["gate.expires_at_window_end"] != false {
		t.Fatalf("facts = %v", r.Facts)
	}
	e.app.Now = at(27, 0, 10)
	e.refuse(t, true, "--target", "qa")
	e.verifyRefused(t, "--target", "qa")
	e.app.Now = at(26, 23, 59)
	if code := e.cmd("deploy", "--target", "qa"); code != exitcode.OK {
		t.Fatalf("deploy before the blackout exit %d. output:\n%s", code, e.out)
	}
}

// A blackout for another tier, or one that starts after the TTL, leaves the TTL alone.
func TestRTPermitKeepsTheTTLWhenNoBlackoutIsInReach(t *testing.T) {
	for name, cfg := range map[string]string{
		"another tier":  strings.Replace(rtBlackoutFromTomorrow, "tiers: [qa]", "tiers: [prod]", 1),
		"after the TTL": strings.Replace(rtBlackoutFromTomorrow, `from: "2026-10-27"`, `from: "2026-10-28"`, 1),
	} {
		t.Run(name, func(t *testing.T) {
			e := rtNew(t, cfg)
			rtUseRules(t, policy.Default)
			e.app.Now = func() time.Time { return time.Date(2026, 10, 26, 23, 30, 0, 0, time.UTC) }
			e.gate(t, exitcode.OK, "--target", "qa")
			p := e.readPermit(t)
			if got := p.ExpiresAt.Sub(p.IssuedAt); got != 2*time.Hour {
				t.Fatalf("lifetime %s, want the 2h TTL", got)
			}
			if r := e.result(t, "gate"); r.Facts["gate.expires_at_blackout_start"] != false {
				t.Fatalf("facts = %v", r.Facts)
			}
		})
	}
}
