package cli

import (
	"strings"
	"testing"

	"github.com/glassops-platform/glassops/internal/core/exitcode"
	"github.com/glassops-platform/glassops/internal/core/policy"
	"github.com/glassops-platform/glassops/internal/core/run"
)

// validateFor is a validate result of this harness's plan for target, with status.
func validateFor(t *testing.T, h *harness, target string, status run.Status, plan string) *run.Result {
	t.Helper()
	r := run.New("validate", nil)
	r.Target = target
	r.Status = status
	r.Facts["salesforce.check_only"] = true
	if plan == "" {
		plan = planHash(t, h)
	}
	r.Facts["validate.plan_sha256"] = plan
	return r
}

func TestGateValidatePassed(t *testing.T) {
	withRule := func(tiers string) string {
		return strings.Replace(gateConfig, "  permit_ttl: 90m\n", "  permit_ttl: 90m\n  rules:\n    validate_passed: { tiers: "+tiers+" }\n", 1)
	}
	cases := []struct {
		name    string
		cfg     string
		target  string
		result  func(t *testing.T, h *harness) *run.Result
		code    int
		outcome policy.Outcome
		detail  string
	}{
		{"failed validate on prod denies by default", gateConfig, "prod",
			func(t *testing.T, h *harness) *run.Result { return validateFor(t, h, "prod", run.StatusFail, "") },
			exitcode.Fail, policy.Fail, "is fail, not pass"},
		{"errored validate on prod denies", gateConfig, "prod",
			func(t *testing.T, h *harness) *run.Result { return validateFor(t, h, "prod", run.StatusError, "") },
			exitcode.Fail, policy.Fail, "is error"},
		{"passed validate on prod passes", gateConfig, "prod",
			func(t *testing.T, h *harness) *run.Result { return validateFor(t, h, "prod", run.StatusPass, "") },
			exitcode.OK, policy.Pass, "passed"},
		{"no validate result is skipped", gateConfig, "prod", nil,
			exitcode.OK, policy.Skipped, "no validate result"},
		{"a result for another target is skipped", gateConfig, "prod",
			func(t *testing.T, h *harness) *run.Result { return validateFor(t, h, "qa", run.StatusFail, "") },
			exitcode.OK, policy.Skipped, `for target "qa"`},
		{"a result for another plan is skipped", gateConfig, "prod",
			func(t *testing.T, h *harness) *run.Result { return validateFor(t, h, "prod", run.StatusFail, "0000") },
			exitcode.OK, policy.Skipped, "different plan"},
		{"qa is not covered by default", gateConfig, "qa",
			func(t *testing.T, h *harness) *run.Result { return validateFor(t, h, "qa", run.StatusFail, "") },
			exitcode.OK, policy.Skipped, "does not cover tier qa"},
		{"qa covered when listed", withRule("[prod, qa]"), "qa",
			func(t *testing.T, h *harness) *run.Result { return validateFor(t, h, "qa", run.StatusFail, "") },
			exitcode.Fail, policy.Fail, "is fail"},
		{"an empty list switches it off", withRule("[]"), "prod",
			func(t *testing.T, h *harness) *run.Result { return validateFor(t, h, "prod", run.StatusFail, "") },
			exitcode.OK, policy.Skipped, "does not cover tier prod"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			installGateFakes(t, fixedRule("plan_not_empty", policy.Pass))
			h, _ := gateHarness(t, tc.cfg)
			if tc.result != nil {
				writeResult(t, h, tc.result(t, h))
			}
			if code := h.gate("--target", tc.target); code != tc.code {
				t.Fatalf("exit %d, want %d, output:\n%s", code, tc.code, h.out)
			}
			p := readPermit(t, h)
			c, ok := findCheck(p.Checks, "validate_passed")
			if !ok || c.Outcome != tc.outcome || !strings.Contains(c.Detail, tc.detail) {
				t.Fatalf("validate_passed = %+v, want %s containing %q", c, tc.outcome, tc.detail)
			}
			if p.Allow != (tc.code == exitcode.OK) {
				t.Fatalf("allow = %v with exit %d", p.Allow, tc.code)
			}
		})
	}
}

// An ephemeral target gets the automatic PERMIT, and validate_passed does not run for it.
func TestGateValidatePassedNotForEphemeral(t *testing.T) {
	installGateFakes(t, fixedRule("plan_not_empty", policy.Pass))
	h, _ := gateHarness(t, gateConfig)
	writeResult(t, h, validateFor(t, h, "scratch", run.StatusFail, ""))
	if code := h.gate("--target", "scratch"); code != exitcode.OK {
		t.Fatalf("exit %d, output:\n%s", code, h.out)
	}
	if _, ok := findCheck(readPermit(t, h).Checks, "validate_passed"); ok {
		t.Fatal("validate_passed ran for an ephemeral target")
	}
}
