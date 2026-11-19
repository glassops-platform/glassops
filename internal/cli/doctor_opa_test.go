package cli

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/glassops-platform/glassops/internal/core/config"
	"github.com/glassops-platform/glassops/internal/core/exitcode"
	"github.com/glassops-platform/glassops/internal/core/policy"
	"github.com/glassops-platform/glassops/internal/core/run"
	"github.com/glassops-platform/glassops/internal/providers"
)

// The opa outputs below are synthetic, in the shape "opa eval --format json" printed for OPA
// 1.22.0 over examples/rego.
const (
	syntheticOPADenyOut      = `{"result":[{"expressions":[{"value":[{"detail":"the plan changes the protected profile Admin","rule":"protected_profile"}],"text":"data.glassops.deny"}]}]}`
	syntheticOPAEmptyOut     = `{"result":[{"expressions":[{"value":[],"text":"data.glassops.deny"}]}]}`
	syntheticOPAUndefinedOut = `{}`
	syntheticOPAVersion      = "Version: 1.22.0\nBuild Commit: synthetic\nRego Version: v1\n"
)

// regoGateConfig is gateConfig with one Rego file.
var regoGateConfig = strings.Replace(gateConfig, "  permit_ttl: 90m\n", "  permit_ttl: 90m\n  rego:\n    files: [policy/deploy.rego]\n", 1)

// gateOPACall is what gate handed the stand-in opa.
type gateOPACall struct {
	dir   string
	args  []string
	input map[string]any
}

// fakeGateOPA replaces policy.RunOPA with a stand-in that records each call and returns out
// and err.
func fakeGateOPA(t *testing.T, out string, err error) *[]gateOPACall {
	t.Helper()
	calls := &[]gateOPACall{}
	old := policy.RunOPA
	t.Cleanup(func() { policy.RunOPA = old })
	policy.RunOPA = func(_ context.Context, dir string, args ...string) ([]byte, error) {
		c := gateOPACall{dir: dir, args: args}
		for i, a := range args {
			if a == "--input" && i+1 < len(args) {
				data, rerr := os.ReadFile(args[i+1])
				if rerr != nil {
					t.Errorf("read the opa input: %v", rerr)
				}
				if jerr := json.Unmarshal(data, &c.input); jerr != nil {
					t.Errorf("opa input is not JSON: %v", jerr)
				}
			}
		}
		*calls = append(*calls, c)
		return []byte(out), err
	}
	return calls
}

func regoGateHarness(t *testing.T) *harness {
	t.Helper()
	h, _ := gateHarness(t, regoGateConfig)
	h.write(t, "policy/deploy.rego", "package glassops\n")
	return h
}

func TestGateRegoDenies(t *testing.T) {
	installGateFakes(t, fixedRule("plan_not_empty", policy.Pass))
	h := regoGateHarness(t)
	calls := fakeGateOPA(t, syntheticOPADenyOut, nil)
	if code := h.gate("--target", "prod"); code != exitcode.Fail {
		t.Fatalf("exit %d, want 1. output:\n%s", code, h.out)
	}
	p := readPermit(t, h)
	verifySigned(t, p)
	if p.Allow {
		t.Fatal("a Rego denial produced an allowing PERMIT")
	}
	if c, ok := findCheck(p.Checks, "plan_not_empty"); !ok || c.Outcome != policy.Pass {
		t.Fatalf("Go rule result changed: %+v", p.Checks)
	}
	if c, ok := findCheck(p.Checks, "rego.protected_profile"); !ok || c.Outcome != policy.Fail || c.Detail != "the plan changes the protected profile Admin" {
		t.Fatalf("no failed rego.protected_profile check: %+v", p.Checks)
	}
	if _, ok := findCheck(p.Checks, "rego"); ok {
		t.Fatalf("a denial also added the rego summary check: %+v", p.Checks)
	}
	if checkStatus(h.result(t, "gate"), "rego.protected_profile") != run.StatusFail {
		t.Fatalf("result: %+v", h.result(t, "gate").Checks)
	}

	// opa ran once, in the project dir, over the facts gate wrote to facts.json.
	if len(*calls) != 1 {
		t.Fatalf("opa ran %d times", len(*calls))
	}
	c := (*calls)[0]
	if c.dir != h.dir {
		t.Fatalf("opa ran in %q, want %q", c.dir, h.dir)
	}
	if c.args[len(c.args)-1] != policy.RegoQuery || !strings.Contains(strings.Join(c.args, " "), "--data policy/deploy.rego") {
		t.Fatalf("args = %q", c.args)
	}
	data, err := os.ReadFile(filepath.Join(h.dir, ".glassops", "facts.json"))
	if err != nil {
		t.Fatal(err)
	}
	var written map[string]any
	if err := json.Unmarshal(data, &written); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(c.input, written) {
		t.Fatalf("opa input differs from facts.json:\n%v\n%v", c.input, written)
	}
}

func TestGateRegoEmptyPasses(t *testing.T) {
	installGateFakes(t, fixedRule("plan_not_empty", policy.Pass))
	h := regoGateHarness(t)
	fakeGateOPA(t, syntheticOPAEmptyOut, nil)
	if code := h.gate("--target", "prod"); code != exitcode.OK {
		t.Fatalf("exit %d, output:\n%s", code, h.out)
	}
	p := readPermit(t, h)
	if !p.Allow {
		t.Fatalf("an empty deny set denied: %+v", p.Checks)
	}
	if c, ok := findCheck(p.Checks, "rego"); !ok || c.Outcome != policy.Pass || !strings.Contains(c.Detail, "policy/deploy.rego") {
		t.Fatalf("no passing rego check: %+v", p.Checks)
	}
}

// TestGateRegoInputIsGatesOwnFacts pins that opa sees only the facts document gate built. ci.
// and target. keys planted in an unsigned result file or in the plan's facts never reach the
// Rego input, and change. keys from the plan's facts do not override what gate derived.
func TestGateRegoInputIsGatesOwnFacts(t *testing.T) {
	installGateFakes(t, fixedRule("plan_not_empty", policy.Pass))
	h := regoGateHarness(t)
	calls := fakeGateOPA(t, syntheticOPAEmptyOut, nil)
	planted := run.New("plan", nil)
	planted.Facts["ci.actor"] = "someone-else"
	planted.Facts["ci.pr_author"] = "someone-else"
	planted.Facts["target.id"] = "00D0000000000EVILAA"
	planted.Facts["target.tier"] = "sandbox"
	writeResult(t, h, planted)
	writeGatePlan(t, h, providers.ChangeSet{
		Destructive: []string{},
		Facts: map[string]any{
			"ci.actor":                 "someone-else",
			"target.kind":              "scratch",
			"change.destructive_count": 7,
			"change.empty":             true,
		},
	})
	if code := h.gate("--target", "prod"); code != exitcode.OK {
		t.Fatalf("exit %d, output:\n%s", code, h.out)
	}
	if len(*calls) != 1 {
		t.Fatalf("opa ran %d times", len(*calls))
	}
	in, _ := (*calls)[0].input["facts"].(map[string]any)
	if in == nil {
		t.Fatalf("opa input has no facts: %v", (*calls)[0].input)
	}
	for k, want := range map[string]any{
		"ci.actor":                 "deployer",
		"target.id":                "00D000000000001AAA",
		"target.tier":              "prod",
		"target.kind":              "production",
		"change.destructive_count": float64(0),
		"change.empty":             false,
	} {
		if in[k] != want {
			t.Errorf("opa input %s = %v, want %v", k, in[k], want)
		}
	}
	if v, ok := in["ci.pr_author"]; ok && v == "someone-else" {
		t.Errorf("a planted ci.pr_author reached the opa input")
	}
	data, err := os.ReadFile(filepath.Join(h.dir, ".glassops", "facts.json"))
	if err != nil {
		t.Fatal(err)
	}
	var written map[string]any
	if err := json.Unmarshal(data, &written); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual((*calls)[0].input, written) {
		t.Fatalf("opa input differs from facts.json:\n%v\n%v", (*calls)[0].input, written)
	}
}

// TestGateRegoCannotPassAGoFailure pins T19: an empty deny set never turns a Go failure into a
// pass.
func TestGateRegoCannotPassAGoFailure(t *testing.T) {
	installGateFakes(t, fixedRule("plan_not_empty", policy.Fail))
	h := regoGateHarness(t)
	fakeGateOPA(t, syntheticOPAEmptyOut, nil)
	if code := h.gate("--target", "prod"); code != exitcode.Fail {
		t.Fatalf("exit %d, output:\n%s", code, h.out)
	}
	p := readPermit(t, h)
	if p.Allow {
		t.Fatal("Rego passed a Go failure")
	}
	if c, ok := findCheck(p.Checks, "plan_not_empty"); !ok || c.Outcome != policy.Fail {
		t.Fatalf("Go rule result changed: %+v", p.Checks)
	}
}

// TestGateRegoFailsClosedOnOPA covers a missing opa, an opa error and an undefined deny: each
// exits 2 and writes a signed PERMIT that denies.
func TestGateRegoFailsClosedOnOPA(t *testing.T) {
	cases := []struct {
		name string
		out  string
		err  error
		want string
	}{
		{"missing opa", "", policy.ErrOPANotFound, "opa is not installed"},
		{"opa error", `{"errors":[{"message":"unexpected eof token","location":{"file":"policy/deploy.rego","row":3}}]}`, errors.New("opa: exit status 2"), "unexpected eof token"},
		{"undefined deny", syntheticOPAUndefinedOut, nil, "undefined"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			installGateFakes(t, fixedRule("plan_not_empty", policy.Pass))
			h := regoGateHarness(t)
			fakeGateOPA(t, tc.out, tc.err)
			if code := h.gate("--target", "prod"); code != exitcode.Error {
				t.Fatalf("exit %d, want 2. output:\n%s", code, h.out)
			}
			p := readPermit(t, h)
			verifySigned(t, p)
			if p.Allow {
				t.Fatal("the PERMIT allows although Rego could not be evaluated")
			}
			if c, ok := findCheck(p.Checks, "rego"); !ok || c.Outcome != policy.Fail || !strings.Contains(c.Detail, tc.want) {
				t.Fatalf("rego check: %+v", p.Checks)
			}
			if checkStatus(h.result(t, "gate"), "policy.evaluate") != run.StatusError {
				t.Fatalf("result: %+v", h.result(t, "gate").Checks)
			}
		})
	}
}

func TestGateRegoMissingFileFailsClosed(t *testing.T) {
	installGateFakes(t, fixedRule("plan_not_empty", policy.Pass))
	h, _ := gateHarness(t, regoGateConfig)
	calls := fakeGateOPA(t, syntheticOPAEmptyOut, nil)
	if code := h.gate("--target", "prod"); code != exitcode.Error {
		t.Fatalf("exit %d, want 2. output:\n%s", code, h.out)
	}
	if len(*calls) != 0 {
		t.Fatal("opa ran although a policy file is missing")
	}
	if c, ok := findCheck(readPermit(t, h).Checks, "rego"); !ok || !strings.Contains(c.Detail, "does not exist") {
		t.Fatalf("rego check: %+v", readPermit(t, h).Checks)
	}
}

// opaApp returns an App whose runner answers opa with the given outputs, keyed by the
// arguments, and fails everything else.
func opaApp(outs map[string]string) *App {
	return &App{Run: func(_ string, name string, args ...string) ([]byte, error) {
		if out, ok := outs[strings.Join(append([]string{name}, args...), " ")]; ok {
			return []byte(out), nil
		}
		return nil, errors.New("exit status 1")
	}}
}

const opaCheckArgs = "opa eval --format json --data policy/deploy.rego data.glassops.deny"

func TestDoctorOPA(t *testing.T) {
	cases := []struct {
		name      string
		outs      map[string]string
		writeFile bool
		tool      run.Status
		files     run.Status
		contain   string
	}{
		{"healthy", map[string]string{"opa version": syntheticOPAVersion, opaCheckArgs: syntheticOPAEmptyOut}, true, run.StatusPass, run.StatusPass, "policy/deploy.rego"},
		{"no opa", nil, true, run.StatusFail, run.StatusSkipped, "opa is not available"},
		{"undefined deny", map[string]string{"opa version": syntheticOPAVersion, opaCheckArgs: syntheticOPAUndefinedOut}, true, run.StatusPass, run.StatusFail, "undefined"},
		{"opa eval fails", map[string]string{"opa version": syntheticOPAVersion}, true, run.StatusPass, run.StatusFail, "opa eval failed"},
		{"missing file", map[string]string{"opa version": syntheticOPAVersion, opaCheckArgs: syntheticOPAEmptyOut}, false, run.StatusPass, run.StatusFail, "does not exist"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()
			if tc.writeFile {
				if err := os.MkdirAll(filepath.Join(dir, "policy"), 0o755); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(filepath.Join(dir, "policy", "deploy.rego"), []byte("package glassops\n"), 0o644); err != nil {
					t.Fatal(err)
				}
			}
			cfg := &config.Config{Policy: config.Policy{Rego: config.Rego{Files: []string{"policy/deploy.rego"}}}}
			rememberProjectDir(cfg, dir)
			checks := doctorOPA(opaApp(tc.outs), cfg)
			if len(checks) != 2 || checks[0].Name != "tool.opa" || checks[1].Name != "policy.rego" {
				t.Fatalf("checks = %+v", checks)
			}
			if checks[0].Status != tc.tool || checks[1].Status != tc.files || !strings.Contains(checks[1].Detail, tc.contain) {
				t.Fatalf("checks = %+v", checks)
			}
			if tc.tool == run.StatusPass && checks[0].Detail != "1.22.0" {
				t.Fatalf("version detail %q", checks[0].Detail)
			}
		})
	}
}

func TestDoctorOPASilentWithoutRego(t *testing.T) {
	if checks := doctorOPA(opaApp(nil), &config.Config{}); checks != nil {
		t.Fatalf("checks = %+v", checks)
	}
	h := newHarness(t)
	h.write(t, "sfdx-project.json", "{}")
	if code := h.exec("doctor"); code != exitcode.OK {
		t.Fatalf("exit %d, output:\n%s", code, h.out)
	}
	if got := checkStatus(h.result(t, "doctor"), "tool.opa"); got != "" {
		t.Fatalf("tool.opa ran without policy.rego: %s", got)
	}
}

// TestDoctorOPACommand runs doctor end to end. The Rego path is absolute because doctor does
// not record the project dir for its config yet, so a relative path would resolve against the
// working directory.
func TestDoctorOPACommand(t *testing.T) {
	h := newHarness(t)
	h.write(t, "sfdx-project.json", "{}")
	h.write(t, "policy/deploy.rego", "package glassops\n")
	abs := filepath.ToSlash(filepath.Join(h.dir, "policy", "deploy.rego"))
	h.write(t, "glassops.yaml", "version: 1\npolicy:\n  rego:\n    files: ['"+abs+"']\n")

	if code := h.exec("doctor"); code != exitcode.Fail {
		t.Fatalf("exit %d without opa, output:\n%s", code, h.out)
	}
	if got := checkStatus(h.result(t, "doctor"), "tool.opa"); got != run.StatusFail {
		t.Fatalf("tool.opa = %s", got)
	}

	h.tools["opa version"] = syntheticOPAVersion
	h.tools["opa eval --format json --data "+abs+" data.glassops.deny"] = syntheticOPAEmptyOut
	if code := h.exec("doctor"); code != exitcode.OK {
		t.Fatalf("exit %d, output:\n%s", code, h.out)
	}
	r := h.result(t, "doctor")
	if checkStatus(r, "tool.opa") != run.StatusPass || checkStatus(r, "policy.rego") != run.StatusPass {
		t.Fatalf("checks: %+v", r.Checks)
	}
}
