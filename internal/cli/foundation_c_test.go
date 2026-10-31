package cli

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/glassops-platform/glassops/internal/core/config"
	"github.com/glassops-platform/glassops/internal/core/exitcode"
	"github.com/glassops-platform/glassops/internal/core/policy"
	"github.com/glassops-platform/glassops/internal/core/run"
	"github.com/glassops-platform/glassops/internal/providers"
)

// sigstoreConfig is gateConfig with sigstore signing selected.
var sigstoreConfig = strings.Replace(gateConfig, "  permit_ttl: 90m\n",
	"  permit_ttl: 90m\n  permit:\n    signing: sigstore\n    identity: https://github.com/acme/app/.github/workflows/deploy.yml@refs/heads/main\n    issuer: https://token.actions.githubusercontent.com\n", 1)

func detailOf(r run.Result, name string) string {
	for _, c := range r.Checks {
		if c.Name == name {
			return c.Detail
		}
	}
	return ""
}

func TestVerifyCommandIsAStub(t *testing.T) {
	h := newHarness(t)
	c, _, err := NewRootCommand(h.app).Find([]string{"verify"})
	if err != nil || c.Name() != "verify" {
		t.Fatalf("verify not registered: %v", err)
	}
	for _, f := range []string{"target", "plan", "deploy-result"} {
		if c.Flags().Lookup(f) == nil {
			t.Fatalf("verify has no --%s", f)
		}
	}
	if code := h.exec("verify"); code != exitcode.Error {
		t.Fatalf("verify without --target exit %d", code)
	}
	if code := h.exec("verify", "--target", "prod"); code != exitcode.Error || !strings.Contains(h.out.String(), "verify is not implemented yet") {
		t.Fatalf("exit %d, output:\n%s", code, h.out)
	}
}

func TestGateRegoFailsClosed(t *testing.T) {
	installGateFakes(t, fixedRule("plan_not_empty", policy.Pass))
	cfg := strings.Replace(gateConfig, "  permit_ttl: 90m\n", "  permit_ttl: 90m\n  rego:\n    files: [policy/deploy.rego]\n", 1)
	h, _ := gateHarness(t, cfg)
	if code := h.gate("--target", "prod"); code != exitcode.Error {
		t.Fatalf("exit %d, want 2. output:\n%s", code, h.out)
	}
	p := readPermit(t, h)
	verifySigned(t, p)
	if p.Allow {
		t.Fatal("a PERMIT allows although rego could not be evaluated")
	}
	if c, ok := findCheck(p.Checks, "plan_not_empty"); !ok || c.Outcome != policy.Pass {
		t.Fatalf("the Go rules did not run first: %+v", p.Checks)
	}
	if c, ok := findCheck(p.Checks, "rego"); !ok || c.Outcome != policy.Fail {
		t.Fatalf("no failed rego check: %+v", p.Checks)
	}
	r := h.result(t, "gate")
	if checkStatus(r, "policy.evaluate") != run.StatusError {
		t.Fatalf("result: %+v", r.Checks)
	}
}

func TestGateWithoutRegoIsUnchanged(t *testing.T) {
	installGateFakes(t, fixedRule("plan_not_empty", policy.Pass))
	h, _ := gateHarness(t, gateConfig)
	if code := h.gate("--target", "prod"); code != exitcode.OK {
		t.Fatalf("exit %d, output:\n%s", code, h.out)
	}
	if _, ok := findCheck(readPermit(t, h).Checks, "rego"); ok {
		t.Fatal("a rego check appeared without policy.rego")
	}
}

func TestGateSigstoreIsNotImplemented(t *testing.T) {
	installGateFakes(t, fixedRule("plan_not_empty", policy.Pass))
	h, _ := gateHarness(t, sigstoreConfig)
	delete(h.env, "GLASSOPS_PERMIT_KEY")
	if code := h.gate("--target", "prod"); code != exitcode.Error {
		t.Fatalf("exit %d, want 2. output:\n%s", code, h.out)
	}
	if _, err := os.Stat(filepath.Join(h.dir, ".glassops", "permit.json")); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("a PERMIT was written: %v", err)
	}
	if d := detailOf(h.result(t, "gate"), "permit.key"); !strings.Contains(d, "sigstore signing is not implemented") {
		t.Fatalf("permit.key detail = %q", d)
	}
}

func TestDeploySigstoreIsNotImplemented(t *testing.T) {
	h, fp := gateHarness(t, sigstoreConfig)
	if code := Execute(h.app, []string{"deploy", "--target", "prod", "--dir", h.dir, "--ci", "github"}); code != exitcode.Error {
		t.Fatalf("exit %d, want 2. output:\n%s", code, h.out)
	}
	if d := detailOf(h.result(t, "deploy"), "permit"); !strings.Contains(d, "sigstore signing is not implemented") {
		t.Fatalf("permit detail = %q", d)
	}
	if fp.deployCalls != 0 || len(fp.sessions) != 0 {
		t.Fatal("deploy reached the target without a verified PERMIT")
	}
}

func TestPermitVerifySigstoreIsNotImplemented(t *testing.T) {
	h, _ := gateHarness(t, sigstoreConfig)
	if code := Execute(h.app, []string{"permit", "verify", "--target", "prod", "--dir", h.dir, "--ci", "github"}); code != exitcode.Error {
		t.Fatalf("exit %d, want 2. output:\n%s", code, h.out)
	}
	if d := detailOf(h.result(t, "permit-verify"), "key"); !strings.Contains(d, "sigstore signing is not implemented") {
		t.Fatalf("key detail = %q", d)
	}
}

func TestConfigureErrorExitsTwo(t *testing.T) {
	h, req, _ := planHarness(t, deltaSet(), nil, gitAnswers(nil))
	fp := h.app.Providers[0].(*fakeProvider)
	var seen *config.Config
	fp.configure = func(c *config.Config) error { seen = c; return errors.New("bad engine") }
	if code := h.exec("plan"); code != exitcode.Error || !strings.Contains(h.out.String(), "bad engine") {
		t.Fatalf("exit %d, output:\n%s", code, h.out)
	}
	if seen == nil || req.OutDir != "" {
		t.Fatalf("configure saw %v, plan ran %+v", seen, *req)
	}
}

func TestEngineHardisPlanIsAToolError(t *testing.T) {
	h := newHarness(t)
	h.write(t, "sfdx-project.json", "{}")
	h.write(t, "glassops.yaml", "version: 1\nsalesforce:\n  engine: hardis\n  delta: { enabled: false }\n")
	if code := h.exec("plan"); code != exitcode.Error {
		t.Fatalf("exit %d, want 2. output:\n%s", code, h.out)
	}
	if d := detailOf(h.result(t, "plan"), "plan"); !strings.Contains(d, "engine hardis is not implemented") {
		t.Fatalf("plan detail = %q", d)
	}
}

func TestRunExportsTelemetryAndOnlyWarns(t *testing.T) {
	var calls []shellCall
	fakeShell(t, map[string]int{}, &calls, nil)
	saved := telemetryExport
	t.Cleanup(func() { telemetryExport = saved })
	var exported []*run.Result
	telemetryExport = func(_ context.Context, cfg *config.Config, rec *run.Result) error {
		exported = append(exported, rec)
		return errors.New("collector down")
	}

	h, _, cfgPath := runHarness(t, providers.ChangeSet{}, "  pr:\n    - plan\ntelemetry:\n  otlp: { endpoint: https://otel.example.com/v1/traces }\n")
	if code := h.exec("run", "pr", "--config", cfgPath); code != exitcode.OK {
		t.Fatalf("exit %d, output:\n%s", code, h.out)
	}
	if len(exported) != 1 || exported[0].Step != "run" {
		t.Fatalf("exported %v", exported)
	}
	if !strings.Contains(h.out.String(), "telemetry export failed: collector down") {
		t.Fatalf("no warning in output:\n%s", h.out)
	}

	exported = nil
	h2, _, cfg2 := runHarness(t, providers.ChangeSet{}, "  pr:\n    - plan\n")
	if code := h2.exec("run", "pr", "--config", cfg2); code != exitcode.OK {
		t.Fatalf("exit %d, output:\n%s", code, h2.out)
	}
	if len(exported) != 0 {
		t.Fatal("telemetry exported without telemetry configured")
	}
}

