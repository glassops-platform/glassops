package cli

import (
	"errors"
	"path/filepath"
	"strings"
	"testing"

	"github.com/glassops-platform/glassops/internal/core/exitcode"
	"github.com/glassops-platform/glassops/internal/core/run"
	"github.com/glassops-platform/glassops/internal/providers"
	"github.com/glassops-platform/glassops/internal/providers/salesforce"
)

const validateConfig = `version: 1
targets:
  qa:
    id: 00D000000000002AAA
    tier: qa
    auth: { method: sfdx-url, url: $SF_QA_AUTH_URL }
  scratch:
    id: 00D000000000003AAA
    tier: ephemeral
    auth: { method: sfdx-url, url: $SF_SCRATCH_AUTH_URL }
salesforce:
  test_level: { qa: RunLocalTests, ephemeral: NoTestRun }
`

func writePlan(t *testing.T, h *harness, empty bool) string {
	t.Helper()
	path := filepath.Join(h.dir, ".glassops", "plan.json")
	cs := providers.ChangeSet{
		Schema: providers.PlanSchema, Provider: "salesforce",
		Dir: ".glassops/plan", Manifest: ".glassops/plan/package/package.xml",
		Destructive: []string{}, Empty: empty, Facts: map[string]any{},
	}
	if err := providers.WriteChangeSet(path, cs); err != nil {
		t.Fatal(err)
	}
	return path
}

func passingResult() providers.ValidateResult {
	return providers.ValidateResult{
		Passed: true, ID: "0Af000000000001AAA",
		Facts: map[string]any{"salesforce.deploy_id": "0Af000000000001AAA", "salesforce.tests_total": 3},
	}
}

// validateHarness returns a harness with the config, a plan and a fake provider whose logins
// all work and whose describe reports each target as the guard expects.
func validateHarness(t *testing.T) (*harness, *fakeProvider) {
	t.Helper()
	h := newHarness(t)
	h.write(t, "glassops.yaml", validateConfig)
	writePlan(t, h, false)
	fp := &fakeProvider{
		detect: true,
		auth:   func(providers.Target) error { return nil },
		describe: func(tg providers.Target) (providers.TargetInfo, error) {
			k := providers.KindSandbox
			if tg.Tier == "ephemeral" {
				k = providers.KindScratch
			}
			return providers.TargetInfo{ID: tg.ID, Kind: k, Facts: map[string]any{}}, nil
		},
		validateFn: func(providers.Target, providers.ChangeSet, providers.ValidateOptions) (providers.ValidateResult, error) {
			return passingResult(), nil
		},
	}
	h.app.Providers = providers.Registry{fp}
	return h, fp
}

func TestValidateRefusesCheckOnlyFalse(t *testing.T) {
	h, fp := validateHarness(t)
	if code := h.exec("validate", "--target", "qa", "--check-only=false"); code != exitcode.Error {
		t.Fatalf("exit %d, want 2. output:\n%s", code, h.out)
	}
	if !strings.Contains(h.out.String(), "never deploys") || len(fp.sessions) != 0 {
		t.Fatalf("output %q, sessions %d", h.out, len(fp.sessions))
	}
}

func TestValidatePass(t *testing.T) {
	h, fp := validateHarness(t)
	var got providers.ValidateOptions
	var onTarget string
	fp.validateFn = func(tg providers.Target, _ providers.ChangeSet, o providers.ValidateOptions) (providers.ValidateResult, error) {
		got, onTarget = o, tg.Name
		return passingResult(), nil
	}
	if code := h.exec("validate", "--target", "qa"); code != exitcode.OK {
		t.Fatalf("exit %d, output:\n%s", code, h.out)
	}
	r := h.result(t, "validate")
	if r.Status != run.StatusPass || r.Target != "qa" || r.Provider != "salesforce" {
		t.Fatalf("result: %+v", r)
	}
	if r.Facts["validate.fallback_used"] != false || r.Facts["salesforce.deploy_id"] != "0Af000000000001AAA" {
		t.Fatalf("facts: %v", r.Facts)
	}
	if got.TestLevel != "RunLocalTests" || got.Dir != h.dir || onTarget != "qa" {
		t.Fatalf("options %+v on %s", got, onTarget)
	}
	if !fp.allClosed() || len(fp.sessions) != 1 {
		t.Fatalf("sessions: %d, closed: %v", len(fp.sessions), fp.allClosed())
	}
}

func TestValidateTestLevelFollowsTierOfTargetUsed(t *testing.T) {
	h, fp := validateHarness(t)
	var levels []string
	fp.validateFn = func(_ providers.Target, _ providers.ChangeSet, o providers.ValidateOptions) (providers.ValidateResult, error) {
		levels = append(levels, o.TestLevel)
		return passingResult(), nil
	}
	if code := h.exec("validate", "--target", "scratch"); code != exitcode.OK {
		t.Fatalf("exit %d, output:\n%s", code, h.out)
	}
	// The fallback's tier counts, not the requested target's.
	fp.auth = func(tg providers.Target) error {
		if tg.Name == "qa" {
			return errors.New("no")
		}
		return nil
	}
	if code := h.exec("validate", "--target", "qa", "--fallback-target", "scratch"); code != exitcode.OK {
		t.Fatalf("exit %d, output:\n%s", code, h.out)
	}
	if len(levels) != 2 || levels[0] != "NoTestRun" || levels[1] != "NoTestRun" {
		t.Fatalf("levels: %v", levels)
	}
}

func TestValidateUnsetTestLevelIsEmpty(t *testing.T) {
	h, fp := validateHarness(t)
	h.write(t, "glassops.yaml", strings.Split(validateConfig, "salesforce:")[0])
	got := "unset"
	fp.validateFn = func(_ providers.Target, _ providers.ChangeSet, o providers.ValidateOptions) (providers.ValidateResult, error) {
		got = o.TestLevel
		return passingResult(), nil
	}
	if code := h.exec("validate", "--target", "qa"); code != exitcode.OK || got != "" {
		t.Fatalf("exit %d, level %q, output:\n%s", code, got, h.out)
	}
}

func TestValidateFailureIsExit1(t *testing.T) {
	h, fp := validateHarness(t)
	fp.validateFn = func(providers.Target, providers.ChangeSet, providers.ValidateOptions) (providers.ValidateResult, error) {
		return providers.ValidateResult{
			Passed: false, ID: "0Af000000000002AAA",
			Problems: []string{"ApexClass Greeter: bad (line 6)", "GreeterTest.m: nope"},
			Facts:    map[string]any{"salesforce.components_failed": 1},
		}, nil
	}
	if code := h.exec("validate", "--target", "qa"); code != exitcode.Fail {
		t.Fatalf("exit %d, output:\n%s", code, h.out)
	}
	r := h.result(t, "validate")
	if r.Status != run.StatusFail || len(r.Checks) != 2 {
		t.Fatalf("result: %+v", r)
	}
	for i, want := range []string{"ApexClass Greeter: bad (line 6)", "GreeterTest.m: nope"} {
		if r.Checks[i].Status != run.StatusFail || r.Checks[i].Detail != want {
			t.Errorf("check %d: %+v", i, r.Checks[i])
		}
	}
	if r.Facts["salesforce.components_failed"] != float64(1) || !fp.allClosed() {
		t.Fatalf("facts %v closed %v", r.Facts, fp.allClosed())
	}
}

func TestValidateProviderErrors(t *testing.T) {
	cases := []struct {
		name string
		err  error
	}{
		{"not check-only", salesforce.ErrNotCheckOnly},
		{"tool error", errors.New("sf exploded")},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			h, fp := validateHarness(t)
			fp.validateFn = func(providers.Target, providers.ChangeSet, providers.ValidateOptions) (providers.ValidateResult, error) {
				return providers.ValidateResult{}, tc.err
			}
			if code := h.exec("validate", "--target", "qa"); code != exitcode.Error {
				t.Fatalf("exit %d, output:\n%s", code, h.out)
			}
			r := h.result(t, "validate")
			if r.Status != run.StatusError || !fp.allClosed() || len(fp.sessions) != 1 {
				t.Fatalf("result %+v closed %v", r, fp.allClosed())
			}
		})
	}
}

func TestValidateEmptyPlanDoesNotLogIn(t *testing.T) {
	h, fp := validateHarness(t)
	writePlan(t, h, true)
	fp.auth = func(providers.Target) error { t.Fatal("logged in for an empty plan"); return nil }
	if code := h.exec("validate", "--target", "qa"); code != exitcode.OK {
		t.Fatalf("exit %d, output:\n%s", code, h.out)
	}
	if r := h.result(t, "validate"); r.Status != run.StatusEmpty || len(fp.sessions) != 0 {
		t.Fatalf("result: %+v", r)
	}
}

func TestValidateMissingPlanIsExit2(t *testing.T) {
	h, fp := validateHarness(t)
	if code := h.exec("validate", "--target", "qa", "--plan", filepath.Join(h.dir, "nope.json")); code != exitcode.Error {
		t.Fatalf("exit %d, output:\n%s", code, h.out)
	}
	if len(fp.sessions) != 0 {
		t.Fatal("logged in without a plan")
	}
}

func TestValidateReadsPlanFlag(t *testing.T) {
	h, fp := validateHarness(t)
	writePlan(t, h, true) // the default plan is empty
	other := filepath.Join(h.dir, "other.json")
	cs := providers.ChangeSet{Schema: providers.PlanSchema, Provider: "salesforce", Manifest: "m.xml", Destructive: []string{}}
	if err := providers.WriteChangeSet(other, cs); err != nil {
		t.Fatal(err)
	}
	var manifest string
	fp.validateFn = func(_ providers.Target, c providers.ChangeSet, _ providers.ValidateOptions) (providers.ValidateResult, error) {
		manifest = c.Manifest
		return passingResult(), nil
	}
	if code := h.exec("validate", "--target", "qa", "--plan", other); code != exitcode.OK || manifest != "m.xml" {
		t.Fatalf("exit %d, manifest %q, output:\n%s", code, manifest, h.out)
	}
}

func TestValidateFallbacks(t *testing.T) {
	loginFails := func(names ...string) func(providers.Target) error {
		return func(tg providers.Target) error {
			for _, n := range names {
				if tg.Name == n {
					return errors.New("invalid_grant")
				}
			}
			return nil
		}
	}
	cases := []struct {
		name      string
		args      []string
		auth      func(providers.Target) error
		wantExit  int
		wantUsed  string
		wantFall  bool
		wantAuths []string
	}{
		{"not configured, fallback set", []string{"--target", "ghost", "--fallback-target", "scratch"}, loginFails(),
			exitcode.OK, "scratch", true, []string{"scratch"}},
		{"not configured, no fallback", []string{"--target", "ghost"}, loginFails(),
			exitcode.Error, "", false, nil},
		{"login fails, fallback set", []string{"--target", "qa", "--fallback-target", "scratch"}, loginFails("qa"),
			exitcode.OK, "scratch", true, []string{"qa", "scratch"}},
		{"login fails, no fallback", []string{"--target", "qa"}, loginFails("qa"),
			exitcode.Error, "", false, []string{"qa"}},
		{"fallback login fails too", []string{"--target", "qa", "--fallback-target", "scratch"}, loginFails("qa", "scratch"),
			exitcode.Error, "", false, []string{"qa", "scratch"}},
		{"fallback not configured", []string{"--target", "ghost", "--fallback-target", "ghost2"}, loginFails(),
			exitcode.Error, "", false, nil},
		{"primary works, fallback unused", []string{"--target", "qa", "--fallback-target", "scratch"}, loginFails(),
			exitcode.OK, "qa", false, []string{"qa"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			h, fp := validateHarness(t)
			var authed []string
			fp.auth = func(tg providers.Target) error {
				authed = append(authed, tg.Name)
				return tc.auth(tg)
			}
			if code := h.exec(append([]string{"validate"}, tc.args...)...); code != tc.wantExit {
				t.Fatalf("exit %d, want %d. output:\n%s", code, tc.wantExit, h.out)
			}
			if strings.Join(authed, ",") != strings.Join(tc.wantAuths, ",") {
				t.Errorf("logins %v, want %v", authed, tc.wantAuths)
			}
			r := h.result(t, "validate")
			if tc.wantExit == exitcode.OK {
				if r.Target != tc.wantUsed || r.Facts["validate.fallback_used"] != tc.wantFall {
					t.Errorf("target %q fallback %v", r.Target, r.Facts["validate.fallback_used"])
				}
			} else if r.Status != run.StatusError {
				t.Errorf("status %s", r.Status)
			}
			if !fp.allClosed() {
				t.Error("a session was left open")
			}
		})
	}
}

func TestValidateGuardMismatchNeverFallsBack(t *testing.T) {
	h, fp := validateHarness(t)
	var authed []string
	fp.auth = func(tg providers.Target) error { authed = append(authed, tg.Name); return nil }
	fp.describe = func(tg providers.Target) (providers.TargetInfo, error) {
		// The qa login reaches some other org.
		id := tg.ID
		if tg.Name == "qa" {
			id = "00D999999999999AAA"
		}
		return providers.TargetInfo{ID: id, Kind: providers.KindScratch}, nil
	}
	fp.validateFn = func(providers.Target, providers.ChangeSet, providers.ValidateOptions) (providers.ValidateResult, error) {
		t.Fatal("validate ran after a guard mismatch")
		return providers.ValidateResult{}, nil
	}
	if code := h.exec("validate", "--target", "qa", "--fallback-target", "scratch"); code != exitcode.Error {
		t.Fatalf("exit %d, output:\n%s", code, h.out)
	}
	if strings.Join(authed, ",") != "qa" {
		t.Fatalf("logins %v, the fallback must not be tried", authed)
	}
	if r := h.result(t, "validate"); r.Status != run.StatusError || !fp.allClosed() {
		t.Fatalf("result %+v closed %v", r, fp.allClosed())
	}
}

func TestValidateTestLevelPrecedence(t *testing.T) {
	cfg := func(levels string) string {
		return strings.Split(validateConfig, "salesforce:")[0] + "salesforce:\n  test_level: " + levels + "\n"
	}
	cases := []struct {
		name     string
		levels   string
		args     []string
		failAuth string
		want     string
	}{
		{"tier wins over default", "{ default: RunLocalTests, qa: RunAllTestsInOrg }", []string{"--target", "qa"}, "", "RunAllTestsInOrg"},
		{"default when tier has no entry", "{ default: RunLocalTests, ephemeral: NoTestRun }", []string{"--target", "qa"}, "", "RunLocalTests"},
		{"empty when neither is set", "{ ephemeral: NoTestRun }", []string{"--target", "qa"}, "", ""},
		{"after fallback, tier of fallback wins", "{ default: RunLocalTests, ephemeral: NoTestRun }",
			[]string{"--target", "qa", "--fallback-target", "scratch"}, "qa", "NoTestRun"},
		{"after fallback, default applies", "{ default: RunLocalTests, qa: RunAllTestsInOrg }",
			[]string{"--target", "qa", "--fallback-target", "scratch"}, "qa", "RunLocalTests"},
		{"after fallback, empty when neither", "{ qa: RunAllTestsInOrg }",
			[]string{"--target", "qa", "--fallback-target", "scratch"}, "qa", ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			h, fp := validateHarness(t)
			h.write(t, "glassops.yaml", cfg(tc.levels))
			fp.auth = func(tg providers.Target) error {
				if tg.Name == tc.failAuth {
					return errors.New("invalid_grant")
				}
				return nil
			}
			got := "unset"
			fp.validateFn = func(_ providers.Target, _ providers.ChangeSet, o providers.ValidateOptions) (providers.ValidateResult, error) {
				got = o.TestLevel
				return passingResult(), nil
			}
			if code := h.exec(append([]string{"validate"}, tc.args...)...); code != exitcode.OK || got != tc.want {
				t.Fatalf("exit %d, level %q, want %q. output:\n%s", code, got, tc.want, h.out)
			}
		})
	}
}
