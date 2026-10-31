package cli

import (
	"path/filepath"
	"testing"

	"github.com/glassops-platform/glassops/internal/core/exitcode"
	"github.com/glassops-platform/glassops/internal/providers"
)

// TestValidateRecordsPlanHashAndTestLevel checks the two facts gate needs to bind a validation
// into a PERMIT: the hash of the plan validated and the test level actually passed.
func TestValidateRecordsPlanHashAndTestLevel(t *testing.T) {
	cases := []struct {
		target, want string
	}{
		{"qa", "RunLocalTests"},
		{"scratch", "NoTestRun"},
	}
	for _, tc := range cases {
		t.Run(tc.target, func(t *testing.T) {
			h, fp := validateHarness(t)
			const sha = "9b1d3f5a7c9e1b3d5f7a9c1e3b5d7f9a1c3e5b7d9f1a3c5e7b9d1f3a5c7e9b1d" // synthetic
			cs := providers.ChangeSet{
				Schema: providers.PlanSchema, Provider: "salesforce",
				Dir: ".glassops/plan", Manifest: ".glassops/plan/package/package.xml",
				Destructive: []string{}, SHA256: sha, Facts: map[string]any{},
			}
			if err := providers.WriteChangeSet(filepath.Join(h.dir, ".glassops", "plan.json"), cs); err != nil {
				t.Fatal(err)
			}
			var passed string
			fp.validateFn = func(_ providers.Target, _ providers.ChangeSet, o providers.ValidateOptions) (providers.ValidateResult, error) {
				passed = o.TestLevel
				r := passingResult()
				// A provider fact under the same key must not replace what validate records.
				r.Facts["validate.plan_sha256"] = "provider-value"
				return r, nil
			}
			if code := h.exec("validate", "--target", tc.target); code != exitcode.OK {
				t.Fatalf("exit %d, output:\n%s", code, h.out)
			}
			r := h.result(t, "validate")
			if r.Facts["validate.plan_sha256"] != sha {
				t.Fatalf("plan hash %v", r.Facts["validate.plan_sha256"])
			}
			if r.Facts["validate.test_level"] != tc.want || passed != tc.want {
				t.Fatalf("test level recorded %v, passed %q, want %q", r.Facts["validate.test_level"], passed, tc.want)
			}
		})
	}
}
