package cli

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/glassops-platform/glassops/internal/core/exitcode"
	"github.com/glassops-platform/glassops/internal/core/permit"
	"github.com/glassops-platform/glassops/internal/core/policy"
	"github.com/glassops-platform/glassops/internal/core/run"
	"github.com/glassops-platform/glassops/internal/providers"
)

// validate records the change set hash gate binds a PERMIT by, from the same function. The
// plan names a destructive file, so the hash differs from plan_sha256.
func TestValidateRecordsTheChangeSetHash(t *testing.T) {
	h, fp := validateHarness(t)
	cs := providers.ChangeSet{
		Schema: providers.PlanSchema, Provider: "salesforce",
		Dir: ".glassops/plan", Manifest: ".glassops/plan/package/package.xml",
		Destructive: []string{".glassops/plan/destructive/destructiveChanges.xml"},
		SHA256:      strings.Repeat("ab", 32), Facts: map[string]any{},
	}
	if err := providers.WriteChangeSet(filepath.Join(h.dir, ".glassops", "plan.json"), cs); err != nil {
		t.Fatal(err)
	}
	fp.validateFn = func(_ providers.Target, _ providers.ChangeSet, _ providers.ValidateOptions) (providers.ValidateResult, error) {
		r := passingResult()
		// A provider fact under the same key must not replace what validate records.
		r.Facts["validate.change_set_sha256"] = "provider-value"
		return r, nil
	}
	if code := h.exec("validate", "--target", "qa"); code != exitcode.OK {
		t.Fatalf("exit %d, output:\n%s", code, h.out)
	}
	want := permit.ChangeSetDigest(cs)
	if got := h.result(t, "validate").Facts["validate.change_set_sha256"]; got != want || want == cs.SHA256 {
		t.Fatalf("recorded %v, want %s", got, want)
	}
}

// gate binds a validation only when the validate result names this change set. A result with
// the right plan hash but another change set hash, or none, is not bound.
func TestGateBindsValidationOnlyForTheSameChangeSet(t *testing.T) {
	const same = "same"
	cases := []struct {
		name   string
		hash   string
		bound  bool
		reason string
	}{
		{"same change set", same, true, "is bound"},
		{"another change set", strings.Repeat("0", 64), false, "different change set"},
		{"none recorded", "", false, "does not record which change set"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			installGateFakes(t, fixedRule("plan_not_empty", policy.Pass))
			h, _ := gateHarness(t, gateConfig)
			r := run.New("validate", nil)
			r.Target = "prod"
			r.Status = run.StatusPass
			r.Facts["salesforce.deploy_id"] = "0Af000000000001AAA"
			r.Facts["salesforce.check_only"] = true
			r.Facts["validate.test_level"] = "RunLocalTests"
			r.Facts["validate.plan_sha256"] = planHash(t, h)
			switch tc.hash {
			case same:
				r.Facts["validate.change_set_sha256"] = changeSetHash(t, h)
			case "":
			default:
				r.Facts["validate.change_set_sha256"] = tc.hash
			}
			writeResult(t, h, r)
			if code := h.gate("--target", "prod"); code != exitcode.OK {
				t.Fatalf("exit %d, output:\n%s", code, h.out)
			}
			p := readPermit(t, h)
			c, _ := findCheck(p.Checks, "validation.bound")
			if (p.Validation != nil) != tc.bound || !strings.Contains(c.Detail, tc.reason) {
				t.Fatalf("validation %+v, check %+v, want bound=%v containing %q", p.Validation, c, tc.bound, tc.reason)
			}
		})
	}
}
