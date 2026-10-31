package cli

import (
	"errors"
	"path/filepath"
	"strings"
	"testing"

	"github.com/glassops-platform/glassops/internal/core/exitcode"
	"github.com/glassops-platform/glassops/internal/core/run"
	"github.com/glassops-platform/glassops/internal/providers"
)

// writeDestructivePlan writes a delta plan from base with one destructive manifest.
func writeDestructivePlan(t *testing.T, h *harness, from string) {
	t.Helper()
	cs := providers.ChangeSet{
		Schema: providers.PlanSchema, Provider: "salesforce", From: from, To: "HEAD",
		Dir: ".glassops/plan", Manifest: ".glassops/plan/package/package.xml",
		Destructive: []string{".glassops/plan/destructiveChanges/destructiveChanges.xml"}, Facts: map[string]any{},
	}
	if err := providers.WriteChangeSet(filepath.Join(h.dir, ".glassops", "plan.json"), cs); err != nil {
		t.Fatal(err)
	}
}

// seedOrder records what ran, so a test can check the seed came before the validate.
type seedOrder struct{ calls []string }

func seedHarness(t *testing.T, seedLine string) (*harness, *fakeProvider, *seedOrder) {
	t.Helper()
	h, fp := validateHarness(t)
	if seedLine != "" {
		h.write(t, "glassops.yaml", validateConfig+"  ephemeral: { seed: "+seedLine+" }\n")
	}
	h.env["SF_SCRATCH_AUTH_URL"] = "force://scratch"
	h.env["SF_QA_AUTH_URL"] = "force://qa"
	writeDestructivePlan(t, h, "abc123")
	order := &seedOrder{}
	fp.validateFn = func(providers.Target, providers.ChangeSet, providers.ValidateOptions) (providers.ValidateResult, error) {
		order.calls = append(order.calls, "validate")
		return passingResult(), nil
	}
	return h, fp, order
}

func TestValidateSeedsScratchBeforeValidate(t *testing.T) {
	for _, mode := range []string{"", "minimal", "full"} {
		t.Run("mode "+mode, func(t *testing.T) {
			h, fp, order := seedHarness(t, mode)
			var got providers.SeedRequest
			var onTarget string
			fp.seed = func(s providers.Session, req providers.SeedRequest) (providers.SeedResult, error) {
				order.calls = append(order.calls, "seed")
				got, onTarget = req, s.Target().Name
				return providers.SeedResult{Mode: req.Mode, Members: []string{"CustomField:Account.Tier__c"}, Facts: map[string]any{"salesforce.seed_deploy_id": "0Af000000000008AAA"}}, nil
			}
			if code := h.exec("validate", "--target", "scratch"); code != exitcode.OK {
				t.Fatalf("exit %d, output:\n%s", code, h.out)
			}
			wantMode := mode
			if wantMode == "" {
				wantMode = "minimal"
			}
			if strings.Join(order.calls, ",") != "seed,validate" || onTarget != "scratch" {
				t.Fatalf("order %v on %s", order.calls, onTarget)
			}
			if got.BaseRef != "abc123" || got.Mode != wantMode || got.Dir != h.dir || len(got.Plan.Destructive) != 1 {
				t.Fatalf("seed request %+v", got)
			}
			r := h.result(t, "validate")
			members, _ := r.Facts["validate.seed.members"].([]any)
			if r.Facts["validate.seed.mode"] != wantMode || r.Facts["validate.seed.fallback"] != false || len(members) != 1 ||
				members[0] != "CustomField:Account.Tier__c" || r.Facts["salesforce.seed_deploy_id"] != "0Af000000000008AAA" {
				t.Fatalf("facts %v", r.Facts)
			}
			if checkStatus(r, "seed") != run.StatusPass || !fp.allClosed() {
				t.Fatalf("result %+v", r)
			}
		})
	}
}

func TestValidateSeedFallbackIsReported(t *testing.T) {
	h, fp, _ := seedHarness(t, "")
	fp.seed = func(providers.Session, providers.SeedRequest) (providers.SeedResult, error) {
		return providers.SeedResult{Mode: "full", Fallback: true, Members: []string{"CustomField:Widget__c.Size__c", "CustomObject:Widget__c"}}, nil
	}
	if code := h.exec("validate", "--target", "scratch"); code != exitcode.OK {
		t.Fatalf("exit %d, output:\n%s", code, h.out)
	}
	r := h.result(t, "validate")
	if r.Facts["validate.seed.mode"] != "full" || r.Facts["validate.seed.fallback"] != true {
		t.Fatalf("facts %v", r.Facts)
	}
	if !strings.Contains(h.out.String(), "warning: validate: the minimal seed failed on a missing dependency") {
		t.Fatalf("no warning:\n%s", h.out)
	}
}

func TestValidateSeedFailureStopsValidate(t *testing.T) {
	h, fp, order := seedHarness(t, "")
	fp.seed = func(providers.Session, providers.SeedRequest) (providers.SeedResult, error) {
		return providers.SeedResult{}, errors.New("salesforce: seed deploy Failed: ApexClass Foo: bad")
	}
	if code := h.exec("validate", "--target", "scratch"); code != exitcode.Error {
		t.Fatalf("exit %d, output:\n%s", code, h.out)
	}
	if len(order.calls) != 0 {
		t.Fatalf("validate ran after a failed seed: %v", order.calls)
	}
	if r := h.result(t, "validate"); checkStatus(r, "seed") != run.StatusError || !fp.allClosed() {
		t.Fatalf("result %+v", r)
	}
}

func TestValidateSeedOff(t *testing.T) {
	h, fp, order := seedHarness(t, "off")
	fp.seed = func(providers.Session, providers.SeedRequest) (providers.SeedResult, error) {
		t.Fatal("seed called with seed off")
		return providers.SeedResult{}, nil
	}
	// The validate fails on the destructive warning, as it does without a seed.
	fp.validateFn = func(providers.Target, providers.ChangeSet, providers.ValidateOptions) (providers.ValidateResult, error) {
		order.calls = append(order.calls, "validate")
		return providers.ValidateResult{Problems: []string{"destructiveChanges.xml: No CustomField named: Account.Tier__c found"}}, nil
	}
	if code := h.exec("validate", "--target", "scratch"); code != exitcode.Fail {
		t.Fatalf("exit %d, output:\n%s", code, h.out)
	}
	r := h.result(t, "validate")
	if r.Facts["validate.seed.mode"] != "off" || checkStatus(r, "seed") != run.StatusSkipped || len(order.calls) != 1 {
		t.Fatalf("result %+v, calls %v", r, order.calls)
	}
}

func TestValidateDoesNotSeed(t *testing.T) {
	cases := []struct {
		name  string
		setup func(t *testing.T, h *harness)
		args  []string
		facts bool
	}{
		{"target is not scratch", nil, []string{"--target", "qa"}, false},
		{"plan has no destructive members", func(t *testing.T, h *harness) { writePlan(t, h, false) }, []string{"--target", "scratch"}, false},
		{"full plan has no base ref", func(t *testing.T, h *harness) { writeDestructivePlan(t, h, "") }, []string{"--target", "scratch"}, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			h, fp, order := seedHarness(t, "")
			if tc.setup != nil {
				tc.setup(t, h)
			}
			fp.seed = func(providers.Session, providers.SeedRequest) (providers.SeedResult, error) {
				t.Fatal("seed called")
				return providers.SeedResult{}, nil
			}
			if code := h.exec(append([]string{"validate"}, tc.args...)...); code != exitcode.OK {
				t.Fatalf("exit %d, output:\n%s", code, h.out)
			}
			r := h.result(t, "validate")
			_, has := r.Facts["validate.seed.mode"]
			if has != tc.facts || len(order.calls) != 1 {
				t.Fatalf("facts %v, calls %v", r.Facts, order.calls)
			}
			if tc.facts && r.Facts["validate.seed.mode"] != "none" {
				t.Fatalf("mode %v", r.Facts["validate.seed.mode"])
			}
		})
	}
}

// A provider without the Seeder capability fails closed rather than validating a
// destructive change it was asked to seed for.
type noSeedProvider struct{ providers.Provider }

func TestValidateSeedNeedsSeeder(t *testing.T) {
	h, fp, order := seedHarness(t, "")
	h.app.Providers = providers.Registry{noSeedProvider{fp}}
	if code := h.exec("validate", "--target", "scratch"); code != exitcode.Error {
		t.Fatalf("exit %d, output:\n%s", code, h.out)
	}
	if !strings.Contains(h.out.String(), "cannot seed a target") || len(order.calls) != 0 {
		t.Fatalf("calls %v, output:\n%s", order.calls, h.out)
	}
}
