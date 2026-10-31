package cli

import (
	"bytes"
	"encoding/base64"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/glassops-platform/glassops/internal/core/config"
	"github.com/glassops-platform/glassops/internal/core/exitcode"
	"github.com/glassops-platform/glassops/internal/core/permit"
	"github.com/glassops-platform/glassops/internal/core/run"
	"github.com/glassops-platform/glassops/internal/providers"
)

const testCommit = "0123456789abcdef0123456789abcdef01234567"

// testPlanHash is the digest of the synthetic manifest permitHarness writes. Every harness
// writes the same bytes at the same path, so the digest is the same each time.
var testPlanHash string

// testPermitKey is a synthetic key of 32 bytes, not a real one.
func testPermitKey(b byte) string {
	return base64.StdEncoding.EncodeToString(bytes.Repeat([]byte{b}, 32))
}

// permitHarness has a config, a plan with a known hash, and the CI environment of a push
// run, so the commit is known. The PERMIT key is set to testPermitKey(1).
func permitHarness(t *testing.T) *harness {
	t.Helper()
	h := newHarness(t)
	h.write(t, "glassops.yaml", validateConfig)
	h.app.Providers = providers.Registry{&fakeProvider{detect: true}}
	h.write(t, "sfdx-project.json", "{}")
	h.write(t, ".glassops/plan/package/package.xml", "<Package/>")
	digest, err := providers.PlanDigest(h.dir, filepath.Join(h.dir, ".glassops", "plan"))
	if err != nil {
		t.Fatal(err)
	}
	testPlanHash = digest
	cs := providers.ChangeSet{
		Schema: providers.PlanSchema, Provider: "salesforce", SHA256: testPlanHash,
		Dir: ".glassops/plan", Manifest: ".glassops/plan/package/package.xml",
		Destructive: []string{}, Facts: map[string]any{},
	}
	if err := providers.WriteChangeSet(filepath.Join(h.dir, ".glassops", "plan.json"), cs); err != nil {
		t.Fatal(err)
	}
	h.env["GITHUB_ACTIONS"] = "true"
	h.env["GITHUB_SHA"] = testCommit
	h.env["GITHUB_EVENT_NAME"] = "push"
	h.env[permit.KeyEnv] = testPermitKey(1)
	h.env["SF_QA_AUTH_URL"] = "force://qa"
	h.env["SF_SCRATCH_AUTH_URL"] = "force://scratch"
	return h
}

func (h *harness) execGitHub(args ...string) int {
	return Execute(h.app, append(args, "--dir", h.dir, "--ci", "github"))
}

func configHash(t *testing.T, h *harness) string {
	t.Helper()
	l, err := config.Load(h.dir, "", "salesforce")
	if err != nil {
		t.Fatal(err)
	}
	return l.SHA256()
}

// writePermit signs a PERMIT that matches the harness, after edit, and writes it.
func writePermit(t *testing.T, h *harness, keyByte byte, edit func(p *permit.Permit)) {
	t.Helper()
	p := permit.Permit{
		Schema: permit.Schema, ID: "synthetic-id", Allow: true,
		IssuedAt: h.app.Now().Add(-time.Minute), ExpiresAt: h.app.Now().Add(time.Hour),
		// The 18 character form of the qa Id, to prove the Ids compare in canonical form.
		Target: permit.Target{Name: "qa", Tier: "qa", ID: "00D000000000002AAA", Kind: "sandbox"},
		Commit: testCommit, PlanSHA256: testPlanHash, ConfigSHA256: configHash(t, h),
	}
	if edit != nil {
		edit(&p)
	}
	if err := permit.Sign(&p, permit.NewHMAC(bytes.Repeat([]byte{keyByte}, 32))); err != nil {
		t.Fatal(err)
	}
	if err := permit.Write(filepath.Join(h.dir, ".glassops", "permit.json"), p); err != nil {
		t.Fatal(err)
	}
}

func TestPermitVerifyPasses(t *testing.T) {
	h := permitHarness(t)
	writePermit(t, h, 1, nil)
	if code := h.execGitHub("permit", "verify", "--target", "qa"); code != exitcode.OK {
		t.Fatalf("exit %d, output:\n%s", code, h.out)
	}
	r := h.result(t, "permit-verify")
	if r.Status != run.StatusPass || r.Target != "qa" || checkStatus(r, "permit") != run.StatusPass {
		t.Fatalf("result: %+v", r)
	}
	if r.Facts["permit.id"] != "synthetic-id" {
		t.Fatalf("facts: %v", r.Facts)
	}
}

func TestPermitVerifyFailsExitOne(t *testing.T) {
	cases := []struct {
		name   string
		keyB   byte
		edit   func(p *permit.Permit)
		after  func(t *testing.T, h *harness)
		reason string
	}{
		{"denied", 1, func(p *permit.Permit) { p.Allow = false }, nil, "denies"},
		{"expired", 1, func(p *permit.Permit) { p.ExpiresAt = time.Date(2026, 10, 26, 11, 0, 0, 0, time.UTC) }, nil, "expired"},
		{"other commit", 1, func(p *permit.Permit) { p.Commit = "ffff" }, nil, "commit"},
		{"other plan", 1, func(p *permit.Permit) { p.PlanSHA256 = "ffff" }, nil, "plan hash"},
		{"other config", 1, func(p *permit.Permit) { p.ConfigSHA256 = "ffff" }, nil, "config hash"},
		{"other target", 1, func(p *permit.Permit) { p.Target.Name = "scratch" }, nil, "target"},
		{"other target id", 1, func(p *permit.Permit) { p.Target.ID = "00D000000000009AAA" }, nil, "target Id"},
		{"signed with another key", 2, nil, nil, "signature"},
		{"edited after signing", 1, nil, func(t *testing.T, h *harness) {
			path := filepath.Join(h.dir, ".glassops", "permit.json")
			data, _ := os.ReadFile(path)
			data = bytes.Replace(data, []byte(`"tier": "qa"`), []byte(`"tier": "dev"`), 1)
			if err := os.WriteFile(path, data, 0o644); err != nil {
				t.Fatal(err)
			}
		}, "signature"},
		{"not a permit", 1, nil, func(t *testing.T, h *harness) {
			if err := os.WriteFile(filepath.Join(h.dir, ".glassops", "permit.json"), []byte("{nope"), 0o644); err != nil {
				t.Fatal(err)
			}
		}, "not a readable permit"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			h := permitHarness(t)
			writePermit(t, h, c.keyB, c.edit)
			if c.after != nil {
				c.after(t, h)
			}
			if code := h.execGitHub("permit", "verify", "--target", "qa"); code != exitcode.Fail {
				t.Fatalf("exit %d, output:\n%s", code, h.out)
			}
			r := h.result(t, "permit-verify")
			if r.Status != run.StatusFail || !strings.Contains(detail(r, "permit"), c.reason) {
				t.Fatalf("result: %+v", r)
			}
		})
	}
}

func TestPermitVerifyChangedFilesFail(t *testing.T) {
	t.Run("config edited", func(t *testing.T) {
		h := permitHarness(t)
		writePermit(t, h, 1, nil)
		h.write(t, "glassops.yaml", validateConfig+"# edited\n")
		if code := h.execGitHub("permit", "verify", "--target", "qa"); code != exitcode.Fail {
			t.Fatalf("exit %d, output:\n%s", code, h.out)
		}
	})
	t.Run("plan hash changed", func(t *testing.T) {
		h := permitHarness(t)
		writePermit(t, h, 1, nil)
		cs := providers.ChangeSet{Schema: providers.PlanSchema, Provider: "salesforce", SHA256: "other", Destructive: []string{}, Facts: map[string]any{}}
		if err := providers.WriteChangeSet(filepath.Join(h.dir, ".glassops", "plan.json"), cs); err != nil {
			t.Fatal(err)
		}
		if code := h.execGitHub("permit", "verify", "--target", "qa"); code != exitcode.Fail {
			t.Fatalf("exit %d, output:\n%s", code, h.out)
		}
	})
	t.Run("plan files edited after plan", func(t *testing.T) {
		h := permitHarness(t)
		writePermit(t, h, 1, nil)
		h.write(t, ".glassops/plan/package/package.xml", "<Package><types/></Package>")
		if code := h.execGitHub("permit", "verify", "--target", "qa"); code != exitcode.Fail {
			t.Fatalf("exit %d, output:\n%s", code, h.out)
		}
		if !strings.Contains(h.out.String(), "changed after plan wrote them") {
			t.Fatalf("output:\n%s", h.out)
		}
	})
	t.Run("other commit", func(t *testing.T) {
		h := permitHarness(t)
		writePermit(t, h, 1, nil)
		h.env["GITHUB_SHA"] = "fedcba9876543210fedcba9876543210fedcba98"
		if code := h.execGitHub("permit", "verify", "--target", "qa"); code != exitcode.Fail {
			t.Fatalf("exit %d, output:\n%s", code, h.out)
		}
	})
}

func TestPermitVerifyToolErrorsExitTwo(t *testing.T) {
	cases := []struct {
		name  string
		setup func(t *testing.T, h *harness)
		check string
	}{
		{"key unset", func(t *testing.T, h *harness) { delete(h.env, permit.KeyEnv) }, "key"},
		{"key empty", func(t *testing.T, h *harness) { h.env[permit.KeyEnv] = "" }, "key"},
		{"key short", func(t *testing.T, h *harness) {
			h.env[permit.KeyEnv] = base64.StdEncoding.EncodeToString([]byte("too short"))
		}, "key"},
		{"key not base64", func(t *testing.T, h *harness) { h.env[permit.KeyEnv] = "not base64 !!" }, "key"},
		{"permit missing", func(t *testing.T, h *harness) {
			if err := os.Remove(filepath.Join(h.dir, ".glassops", "permit.json")); err != nil {
				t.Fatal(err)
			}
		}, "permit"},
		{"plan missing", func(t *testing.T, h *harness) {
			if err := os.Remove(filepath.Join(h.dir, ".glassops", "plan.json")); err != nil {
				t.Fatal(err)
			}
		}, "plan"},
		{"target not configured", func(t *testing.T, h *harness) {
			h.write(t, "glassops.yaml", "version: 1\n")
		}, "target"},
		{"no commit", func(t *testing.T, h *harness) { delete(h.env, "GITHUB_SHA") }, "commit"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			h := permitHarness(t)
			writePermit(t, h, 1, nil)
			c.setup(t, h)
			if code := h.execGitHub("permit", "verify", "--target", "qa"); code != exitcode.Error {
				t.Fatalf("exit %d, output:\n%s", code, h.out)
			}
			r := h.result(t, "permit-verify")
			if r.Status != run.StatusError || checkStatus(r, c.check) != run.StatusError {
				t.Fatalf("result: %+v", r)
			}
		})
	}
}

func TestPermitVerifyNeverPrintsTheKey(t *testing.T) {
	h := permitHarness(t)
	key := h.env[permit.KeyEnv]
	short := base64.StdEncoding.EncodeToString([]byte("too short"))
	for name, setup := range map[string]func(){
		"pass":      func() { writePermit(t, h, 1, nil) },
		"wrong key": func() { writePermit(t, h, 2, nil) },
		"bad key":   func() { h.env[permit.KeyEnv] = short },
	} {
		h.out.Reset()
		setup()
		h.execGitHub("permit", "verify", "--target", "qa")
		data, _ := os.ReadFile(filepath.Join(h.dir, ".glassops", "results", "permit-verify.json"))
		for _, secret := range []string{key, short, testPermitKey(2)} {
			if strings.Contains(h.out.String(), secret) || strings.Contains(string(data), secret) {
				t.Fatalf("%s: output or result contains a key", name)
			}
		}
	}
}

func TestDoctorWarnsWhenPermitKeyIsInAPullRequestRun(t *testing.T) {
	for _, tc := range []struct {
		event   string
		keySet  bool
		warning bool
	}{
		{"pull_request", true, true},
		{"pull_request_target", true, true},
		{"pull_request", false, false},
		{"push", true, false},
		{"workflow_dispatch", true, false},
	} {
		t.Run(tc.event, func(t *testing.T) {
			h := permitHarness(t)
			h.env["GITHUB_EVENT_NAME"] = tc.event
			if !tc.keySet {
				delete(h.env, permit.KeyEnv)
			}
			if code := h.execGitHub("doctor"); code != exitcode.OK {
				t.Fatalf("a warning must not fail doctor: exit %d, output:\n%s", code, h.out)
			}
			r := h.result(t, "doctor")
			d := detail(r, "permit.key")
			if tc.warning != strings.HasPrefix(d, "warning:") {
				t.Fatalf("detail %q, want warning %v", d, tc.warning)
			}
			if tc.warning && !strings.Contains(h.out.String(), "::warning::") {
				t.Fatalf("no CI warning annotation in:\n%s", h.out)
			}
			if strings.Contains(h.out.String(), testPermitKey(1)) {
				t.Fatal("doctor printed the key")
			}
		})
	}
}
