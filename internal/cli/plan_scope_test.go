package cli

import (
	"strings"
	"testing"

	"github.com/glassops-platform/glassops/internal/core/exitcode"
	"github.com/glassops-platform/glassops/internal/core/run"
	"github.com/glassops-platform/glassops/internal/providers"
)

// A target under environments only, with a full scope on its targets entry. The project
// default stays delta.
const planScopeConfig = `version: 1
targets:
  devhub: { tier: devhub }
  qa:
    tier: qa
    branch: develop
  ephemeral:
    plan: { scope: full }
environments:
  ephemeral: { devhub: devhub }
  scratch2: { devhub: devhub }
`

// A target with scope full plans the whole source and never looks up its tag or a merge
// base, even in a shallow clone.
func TestPlanTargetScopeFull(t *testing.T) {
	cs := deltaSet()
	cs.Destructive = []string{}
	h, req, calls := planHarness(t, cs, nil, map[string]string{
		"rev-parse --is-shallow-repository": "true",
		"rev-parse HEAD":                    planHead,
	})
	h.write(t, "glassops.yaml", planScopeConfig)
	if code := h.exec("plan", "--target", "ephemeral"); code != exitcode.OK {
		t.Fatalf("exit %d, output:\n%s", code, h.out)
	}
	if !req.Full || req.From != "" || req.To != planHead {
		t.Fatalf("request = %+v, want a full plan at HEAD", req)
	}
	for _, c := range *calls {
		if strings.Contains(c, "refs/tags") || strings.Contains(c, "merge-base") || strings.Contains(c, "shallow") {
			t.Fatalf("a full scope still ran git %q", c)
		}
	}
	r := h.result(t, "plan")
	if r.Status != run.StatusPass || r.Target != "ephemeral" || r.Facts["plan.full"] != true ||
		r.Facts["plan.scope"] != "full" || r.Facts["plan.scope_source"] != "targets.ephemeral.plan.scope" ||
		r.Facts["plan.first_deploy"] != false {
		t.Fatalf("result = %+v", r)
	}
	if d := checkDetail(r, "refs"); !strings.Contains(d, "targets.ephemeral.plan.scope") {
		t.Fatalf("refs check = %q, want it to name the setting", d)
	}
}

// Another target in the same config keeps the project default, delta from its tag.
func TestPlanOtherTargetKeepsDelta(t *testing.T) {
	h, req, _ := planHarness(t, deltaSet(), nil, gitAnswers(map[string]string{
		"rev-parse --verify --quiet refs/tags/glassops/deployed/qa^{commit}": planBase,
	}))
	h.write(t, "glassops.yaml", planScopeConfig)
	if code := h.exec("plan", "--target", "qa"); code != exitcode.OK {
		t.Fatalf("exit %d, output:\n%s", code, h.out)
	}
	if req.Full || req.From != planBase {
		t.Fatalf("request = %+v, want a delta from the tag", req)
	}
	r := h.result(t, "plan")
	if r.Facts["plan.scope"] != "delta" || r.Facts["plan.scope_source"] != "default" {
		t.Fatalf("facts = %+v", r.Facts)
	}
}

// A name only under environments is accepted, and gets the project default.
func TestPlanEnvironmentOnlyTarget(t *testing.T) {
	h, req, _ := planHarness(t, deltaSet(), nil, gitAnswers(map[string]string{
		"rev-parse --verify --quiet HEAD~1^{commit}": planBase,
	}))
	h.write(t, "glassops.yaml", planScopeConfig)
	if code := h.exec("plan", "--target", "scratch2", "--from", "HEAD~1"); code != exitcode.OK {
		t.Fatalf("exit %d, output:\n%s", code, h.out)
	}
	if req.Full || req.From != planBase {
		t.Fatalf("request = %+v, want the explicit delta", req)
	}
	if r := h.result(t, "plan"); r.Target != "scratch2" || r.Facts["plan.scope"] != "delta" {
		t.Fatalf("result = %+v", r)
	}
}

// --scope wins over both the target and the project setting, in both directions.
func TestPlanScopeFlagOverridesConfig(t *testing.T) {
	t.Run("delta over a full target", func(t *testing.T) {
		h, req, _ := planHarness(t, deltaSet(), nil, gitAnswers(map[string]string{
			"rev-parse --verify --quiet HEAD~1^{commit}": planBase,
		}))
		h.write(t, "glassops.yaml", planScopeConfig)
		if code := h.exec("plan", "--target", "ephemeral", "--scope", "delta", "--from", "HEAD~1"); code != exitcode.OK {
			t.Fatalf("exit %d, output:\n%s", code, h.out)
		}
		if req.Full || req.From != planBase {
			t.Fatalf("request = %+v, want a delta", req)
		}
		if r := h.result(t, "plan"); r.Facts["plan.scope"] != "delta" || r.Facts["plan.scope_source"] != "--scope" {
			t.Fatalf("facts = %+v", r.Facts)
		}
	})
	t.Run("delta over delta.enabled false", func(t *testing.T) {
		h, req, _ := planHarness(t, deltaSet(), nil, gitAnswers(map[string]string{
			"rev-parse --verify --quiet HEAD~1^{commit}": planBase,
		}))
		h.write(t, "glassops.yaml", "version: 1\nsalesforce:\n  delta:\n    enabled: false\n")
		if code := h.exec("plan", "--scope", "delta", "--from", "HEAD~1"); code != exitcode.OK {
			t.Fatalf("exit %d, output:\n%s", code, h.out)
		}
		if req.Full || req.From != planBase {
			t.Fatalf("request = %+v, want a delta", req)
		}
	})
	t.Run("full with no config", func(t *testing.T) {
		cs := deltaSet()
		cs.Full = true
		h, req, _ := planHarness(t, cs, nil, map[string]string{"rev-parse HEAD": planHead})
		if code := h.exec("plan", "--scope", "full"); code != exitcode.OK {
			t.Fatalf("exit %d, output:\n%s", code, h.out)
		}
		if !req.Full {
			t.Fatalf("request = %+v, want a full plan", req)
		}
	})
}

func TestPlanBadScopeIsExit2(t *testing.T) {
	h, req, _ := planHarness(t, deltaSet(), nil, gitAnswers(nil))
	if code := h.exec("plan", "--scope", "since"); code != exitcode.Error {
		t.Fatalf("exit %d, want 2", code)
	}
	if req.Dir != "" {
		t.Fatalf("the provider planned after a bad --scope: %+v", req)
	}
	if !strings.Contains(h.out.String(), "want full or delta") {
		t.Fatalf("output:\n%s", h.out)
	}
}

func TestDoctorPrintsPlanScope(t *testing.T) {
	h := newHarness(t)
	h.app.Providers = providers.Registry{&fakeProvider{detect: true}}
	h.write(t, "glassops.yaml", planScopeConfig)
	h.exec("doctor")
	r := h.result(t, "doctor")
	want := "devhub delta (default), ephemeral full (targets.ephemeral.plan.scope), qa delta (default), scratch2 delta (default)"
	if checkStatus(r, "plan.scope") != run.StatusPass || checkDetail(r, "plan.scope") != want {
		t.Fatalf("plan.scope check = %q, want %q", checkDetail(r, "plan.scope"), want)
	}
}

func TestDoctorPlanScopeSilentWithoutTargets(t *testing.T) {
	h := newHarness(t)
	h.app.Providers = providers.Registry{&fakeProvider{detect: true}}
	h.write(t, "glassops.yaml", "version: 1\n")
	h.exec("doctor")
	for _, c := range h.result(t, "doctor").Checks {
		if c.Name == "plan.scope" {
			t.Fatalf("plan.scope check with no targets: %+v", c)
		}
	}
}
