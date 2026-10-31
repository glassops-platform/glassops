package cli

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/glassops-platform/glassops/internal/core/exitcode"
	"github.com/glassops-platform/glassops/internal/core/permit"
	"github.com/glassops-platform/glassops/internal/core/policy"
	"github.com/glassops-platform/glassops/internal/core/run"
)

// The cli test binary doubles as a fake cosign, as the permit package's does: copied onto PATH
// as cosign and run with GLASSOPS_FAKE_COSIGN=1, it signs and verifies synthetic bundles. Each
// bundle names the blob's SHA-256 and the identity and issuer in FAKE_COSIGN_IDENTITY and
// FAKE_COSIGN_ISSUER. Real bundles differ [unverified, the orchestrator's probe captures one].
func init() {
	if os.Getenv("GLASSOPS_FAKE_COSIGN") == "1" {
		os.Exit(fakeCosignMain(os.Args[1:], os.Stderr))
	}
}

type syntheticBundle struct {
	SHA256, Identity, Issuer string
}

func fakeCosignMain(args []string, stderr io.Writer) int {
	fail := func(msg string) int { fmt.Fprintln(stderr, "Error: "+msg); return 1 }
	if len(args) == 0 {
		return fail("no command")
	}
	flags, pos := map[string]string{}, []string{}
	for i := 1; i < len(args); i++ {
		if a := args[i]; strings.HasPrefix(a, "--") && a != "--yes" && i+1 < len(args) {
			flags[a], i = args[i+1], i+1
		} else if a == "--yes" {
			flags[a] = "true"
		} else {
			pos = append(pos, a)
		}
	}
	if len(pos) != 1 {
		return fail("want one blob")
	}
	blob, err := os.ReadFile(pos[0])
	if err != nil {
		return fail(err.Error())
	}
	sum := sha256.Sum256(blob)
	switch args[0] {
	case "sign-blob":
		if flags["--yes"] == "" || os.Getenv(permit.OIDCRequestTokenEnv) == "" {
			return fail("cannot sign without --yes and an OIDC token")
		}
		b, _ := json.Marshal(syntheticBundle{hex.EncodeToString(sum[:]), os.Getenv("FAKE_COSIGN_IDENTITY"), os.Getenv("FAKE_COSIGN_ISSUER")})
		if err := os.WriteFile(flags["--bundle"], b, 0o600); err != nil {
			return fail(err.Error())
		}
		return 0
	case "verify-blob":
		raw, err := os.ReadFile(flags["--bundle"])
		var b syntheticBundle
		if err != nil || json.Unmarshal(raw, &b) != nil {
			return fail("bundle could not be parsed")
		}
		switch {
		case b.SHA256 != hex.EncodeToString(sum[:]):
			return fail("invalid signature when validating ASN.1 encoded signature")
		case b.Identity != flags["--certificate-identity"]:
			return fail("none of the expected identities matched what was in the certificate")
		case b.Issuer != flags["--certificate-oidc-issuer"]:
			return fail("expected oidc issuer not found in certificate")
		}
		return 0
	}
	return fail("unknown command " + args[0])
}

const sigstoreIdentity = "https://github.com/acme/app/.github/workflows/deploy.yml@refs/heads/main"

// fakeCosignOnPath copies the test binary onto a fresh PATH entry as cosign and makes this
// process look like a GitHub Actions job with id-token: write. The PERMIT signer reads the
// process environment, not the harness's, because cosign does.
func fakeCosignOnPath(t *testing.T) {
	t.Helper()
	self, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(self)
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	name := "cosign"
	if runtime.GOOS == "windows" {
		name += ".exe"
	}
	if err := os.WriteFile(filepath.Join(dir, name), data, 0o755); err != nil {
		t.Fatal(err)
	}
	for k, v := range map[string]string{
		"PATH":                     dir + string(os.PathListSeparator) + os.Getenv("PATH"),
		"GLASSOPS_FAKE_COSIGN":     "1",
		"FAKE_COSIGN_IDENTITY":     sigstoreIdentity,
		"FAKE_COSIGN_ISSUER":       permit.GitHubIssuer,
		"GITHUB_ACTIONS":           "true",
		"BITBUCKET_BUILD_NUMBER":   "",
		"GITLAB_CI":                "",
		permit.OIDCRequestURLEnv:   "https://example.invalid/token",
		permit.OIDCRequestTokenEnv: "syntheticRequestToken",
	} {
		t.Setenv(k, v)
	}
}

// sigstoreGate runs as gateHarness with sigstore signing and one passing rule, signing with
// the real PERMIT code. It has no HMAC key, so nothing can fall back to hmac.
func sigstoreGate(t *testing.T) (*harness, *fakeProvider) {
	t.Helper()
	old := gateRules
	t.Cleanup(func() { gateRules = old })
	gateRules = func() []policy.Rule { return []policy.Rule{fixedRule("plan_not_empty", policy.Pass)} }
	h, fp := gateHarness(t, sigstoreConfig)
	delete(h.env, permit.KeyEnv)
	return h, fp
}

func (h *harness) permitVerify() int {
	return Execute(h.app, []string{"permit", "verify", "--target", "prod", "--dir", h.dir, "--ci", "github"})
}

func TestGateSigstoreRoundTrip(t *testing.T) {
	fakeCosignOnPath(t)
	h, fp := sigstoreGate(t)
	if code := h.gate("--target", "prod"); code != exitcode.OK {
		t.Fatalf("gate exit %d, output:\n%s", code, h.out)
	}
	p := readPermit(t, h)
	if !p.Allow || p.Signature.Alg != permit.SigstoreAlgorithm || p.Signature.Value == "" {
		t.Fatalf("PERMIT: allow %v, signature %+v", p.Allow, p.Signature.Alg)
	}
	if code := h.permitVerify(); code != exitcode.OK {
		t.Fatalf("permit verify exit %d, output:\n%s", code, h.out)
	}
	if r := h.result(t, "permit-verify"); checkStatus(r, "permit") != run.StatusPass {
		t.Fatalf("permit verify: %+v", r.Checks)
	}
	// deploy checks the same signature before it reaches the target.
	Execute(h.app, []string{"deploy", "--target", "prod", "--dir", h.dir, "--ci", "github"})
	if d := detailOf(h.result(t, "deploy"), "permit"); strings.Contains(d, "does not allow") {
		t.Fatalf("deploy refused a good sigstore PERMIT: %s", d)
	}
	if fp.deployCalls != 1 {
		t.Fatalf("deploy calls %d, output:\n%s", fp.deployCalls, h.out)
	}
}

func TestSigstorePermitRefused(t *testing.T) {
	cases := []struct {
		name    string
		change  func(t *testing.T, h *harness)
		contain string
	}{
		{"tampered PERMIT", func(t *testing.T, h *harness) {
			path := filepath.Join(h.dir, ".glassops", "permit.json")
			data, _ := os.ReadFile(path)
			var p permit.Permit
			if err := json.Unmarshal(data, &p); err != nil {
				t.Fatal(err)
			}
			p.ExpiresAt = p.ExpiresAt.Add(24 * 60 * 60 * 1e9)
			if err := permit.Write(path, p); err != nil {
				t.Fatal(err)
			}
		}, "invalid signature"},
		{"wrong identity", func(t *testing.T, h *harness) {
			t.Setenv("FAKE_COSIGN_IDENTITY", "https://github.com/acme/app/.github/workflows/pr.yml@refs/heads/feature")
			h.gate("--target", "prod")
		}, "identities"},
		{"wrong issuer", func(t *testing.T, h *harness) {
			t.Setenv("FAKE_COSIGN_ISSUER", "https://accounts.example.com")
			h.gate("--target", "prod")
		}, "issuer"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			fakeCosignOnPath(t)
			h, fp := sigstoreGate(t)
			if code := h.gate("--target", "prod"); code != exitcode.OK {
				t.Fatalf("gate exit %d, output:\n%s", code, h.out)
			}
			tc.change(t, h)
			if code := h.permitVerify(); code != exitcode.Fail {
				t.Fatalf("permit verify exit %d, want 1. output:\n%s", code, h.out)
			}
			if d := detailOf(h.result(t, "permit-verify"), "permit"); !strings.Contains(d, tc.contain) {
				t.Fatalf("permit detail %q, want %q", d, tc.contain)
			}
			Execute(h.app, []string{"deploy", "--target", "prod", "--dir", h.dir, "--ci", "github"})
			if fp.deployCalls != 0 {
				t.Fatal("deploy reached the target with a refused PERMIT")
			}
		})
	}
}

func TestGateSigstoreRefusesWithoutAPermit(t *testing.T) {
	cases := []struct {
		name    string
		env     map[string]string
		contain string
	}{
		{"no OIDC token", map[string]string{permit.OIDCRequestTokenEnv: ""}, "id-token: write"},
		{"not on GitHub Actions", map[string]string{"GITHUB_ACTIONS": ""}, "GitHub Actions only"},
		{"no cosign", map[string]string{"PATH": ""}, "cosign is not installed"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			fakeCosignOnPath(t)
			for k, v := range tc.env {
				if k == "PATH" {
					v = t.TempDir()
				}
				t.Setenv(k, v)
			}
			h, _ := sigstoreGate(t)
			if code := h.gate("--target", "prod"); code != exitcode.Error {
				t.Fatalf("exit %d, want 2. output:\n%s", code, h.out)
			}
			if _, err := os.Stat(filepath.Join(h.dir, ".glassops", "permit.json")); !errors.Is(err, os.ErrNotExist) {
				t.Fatalf("a PERMIT was written: %v", err)
			}
			if d := detailOf(h.result(t, "gate"), "permit.key"); !strings.Contains(d, tc.contain) {
				t.Fatalf("permit.key detail %q, want %q", d, tc.contain)
			}
			if strings.Contains(h.out.String(), "syntheticRequestToken") {
				t.Fatal("the OIDC request token reached output")
			}
		})
	}
}

func TestGateHMACUnchangedBySigstore(t *testing.T) {
	// A sigstore environment changes nothing for a project on hmac.
	fakeCosignOnPath(t)
	installGateFakes(t, fixedRule("plan_not_empty", policy.Pass))
	h, _ := gateHarness(t, gateConfig)
	if code := h.gate("--target", "prod"); code != exitcode.OK {
		t.Fatalf("exit %d, output:\n%s", code, h.out)
	}
	verifySigned(t, readPermit(t, h))
}

func TestDoctorSigstore(t *testing.T) {
	gh := map[string]string{"GITHUB_ACTIONS": "true", "GITHUB_EVENT_NAME": "push",
		permit.OIDCRequestURLEnv: "https://example.invalid/token", permit.OIDCRequestTokenEnv: "syntheticRequestToken"}
	with := func(base map[string]string, kv ...string) map[string]string {
		out := map[string]string{}
		for k, v := range base {
			out[k] = v
		}
		for i := 0; i+1 < len(kv); i += 2 {
			out[kv[i]] = kv[i+1]
		}
		return out
	}
	cases := []struct {
		name     string
		env      map[string]string
		cosign   bool
		cfg      string
		want     map[string]run.Status
		contains map[string]string
	}{
		{"ready on GitHub", gh, true, sigstoreConfig,
			map[string]run.Status{"tool.cosign": run.StatusPass, "permit.sigstore": run.StatusPass, "permit.identity": run.StatusPass},
			map[string]string{"tool.cosign": "v2.4.1", "permit.identity": sigstoreIdentity}},
		{"no cosign", gh, false, sigstoreConfig,
			map[string]run.Status{"tool.cosign": run.StatusFail},
			map[string]string{"tool.cosign": "cosign not found"}},
		{"no OIDC token on push", with(gh, permit.OIDCRequestTokenEnv, ""), true, sigstoreConfig,
			map[string]run.Status{"permit.sigstore": run.StatusFail},
			map[string]string{"permit.sigstore": "id-token: write"}},
		{"no OIDC token on a pull request", with(gh, permit.OIDCRequestURLEnv, "", "GITHUB_EVENT_NAME", "pull_request"), true, sigstoreConfig,
			map[string]run.Status{"permit.sigstore": run.StatusPass},
			map[string]string{"permit.sigstore": "warning:"}},
		{"GitLab", map[string]string{"GITLAB_CI": "true"}, true, sigstoreConfig,
			map[string]run.Status{"permit.sigstore": run.StatusFail},
			map[string]string{"permit.sigstore": "GitLab CI"}},
		{"Bitbucket", map[string]string{"BITBUCKET_BUILD_NUMBER": "12"}, true, sigstoreConfig,
			map[string]run.Status{"permit.sigstore": run.StatusFail},
			map[string]string{"permit.sigstore": "Bitbucket Pipelines"}},
		{"local", map[string]string{}, true, sigstoreConfig,
			map[string]run.Status{"permit.sigstore": run.StatusSkipped},
			map[string]string{"permit.sigstore": "permit verify works here"}},
		{"odd identity and issuer", gh, true, strings.NewReplacer(sigstoreIdentity, "deployer", permit.GitHubIssuer, "https://accounts.example.com").Replace(sigstoreConfig),
			map[string]run.Status{"permit.identity": run.StatusPass},
			map[string]string{"permit.identity": "warning: issuer is not https://token.actions.githubusercontent.com, identity is not shaped"}},
		{"hmac is silent", gh, false, gateConfig, map[string]run.Status{"tool.cosign": "", "permit.sigstore": "", "permit.identity": ""}, nil},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			h, _ := gateHarness(t, tc.cfg)
			h.env = tc.env
			if tc.cosign {
				h.tools["cosign version"] = "  ______   ______\nGitVersion:    v2.4.1\nGitCommit:     synthetic\n"
			}
			h.exec("doctor")
			r := h.result(t, "doctor")
			for name, want := range tc.want {
				if got := checkStatus(r, name); got != want {
					t.Errorf("%s: %q, want %q (%s)", name, got, want, detailOf(r, name))
				}
			}
			for name, sub := range tc.contains {
				if d := detailOf(r, name); !strings.Contains(d, sub) {
					t.Errorf("%s detail %q, want %q", name, d, sub)
				}
			}
			if strings.Contains(h.out.String(), "syntheticRequestToken") {
				t.Fatal("the OIDC request token reached output")
			}
		})
	}
}
