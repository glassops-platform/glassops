package cli

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	osexec "os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/glassops-platform/glassops/internal/core/exitcode"
	"github.com/glassops-platform/glassops/internal/core/permit"
	"github.com/glassops-platform/glassops/internal/core/policy"
	"github.com/glassops-platform/glassops/internal/core/run"
	"github.com/glassops-platform/glassops/internal/providers"
)

// The cli test binary doubles as a fake cosign, as the permit package's does: copied onto PATH
// as cosign and run with GLASSOPS_FAKE_COSIGN=1, it signs and verifies synthetic bundles. Each
// has the legacy shape cosign v2.5.3 writes, {base64Signature, cert, rekorBundle} [confirmed
// by the orchestrator's probe, log P56L-023]. In the synthetic bundle base64Signature is the
// blob's SHA-256 and cert is base64 JSON of FAKE_COSIGN_IDENTITY, FAKE_COSIGN_ISSUER and
// FAKE_COSIGN_REPOSITORY, where a real cert is a PEM certificate.
func init() {
	if os.Getenv("GLASSOPS_FAKE_COSIGN") == "1" {
		os.Exit(fakeCosignMain(os.Args[1:], os.Stderr))
	}
}

type syntheticLegacyBundle struct {
	Base64Signature string          `json:"base64Signature"`
	Cert            string          `json:"cert"`
	RekorBundle     json.RawMessage `json:"rekorBundle"`
}

type syntheticCert struct {
	Identity, Issuer, Repository string
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
	sig := base64.StdEncoding.EncodeToString(sum[:])
	switch args[0] {
	case "sign-blob":
		if flags["--yes"] == "" || os.Getenv(permit.OIDCRequestTokenEnv) == "" {
			return fail("cannot sign without --yes and an OIDC token")
		}
		cert, _ := json.Marshal(syntheticCert{os.Getenv("FAKE_COSIGN_IDENTITY"), os.Getenv("FAKE_COSIGN_ISSUER"), os.Getenv("FAKE_COSIGN_REPOSITORY")})
		b, _ := json.Marshal(syntheticLegacyBundle{sig, base64.StdEncoding.EncodeToString(cert), json.RawMessage(`{"SignedEntryTimestamp":"c3ludGhldGlj","Payload":{"logIndex":1}}`)})
		if err := os.WriteFile(flags["--bundle"], b, 0o600); err != nil {
			return fail(err.Error())
		}
		return 0
	case "verify-blob":
		raw, err := os.ReadFile(flags["--bundle"])
		var b syntheticLegacyBundle
		var c syntheticCert
		if err != nil || json.Unmarshal(raw, &b) != nil {
			return fail("bundle could not be parsed")
		}
		if certJSON, err := base64.StdEncoding.DecodeString(b.Cert); err != nil || json.Unmarshal(certJSON, &c) != nil {
			return fail("could not parse the certificate")
		}
		switch {
		case flags["--certificate-github-workflow-repository"] == "":
			return fail("verify-blob was not given --certificate-github-workflow-repository")
		case b.Base64Signature != sig:
			return fail("invalid signature when validating ASN.1 encoded signature")
		case c.Identity != flags["--certificate-identity"]:
			return fail("none of the expected identities matched what was in the certificate")
		case c.Issuer != flags["--certificate-oidc-issuer"]:
			return fail("expected oidc issuer not found in certificate")
		case c.Repository != flags["--certificate-github-workflow-repository"]:
			return fail("expected GitHub Workflow Repository not found in certificate")
		}
		return 0
	}
	return fail("unknown command " + args[0])
}

const (
	sigstoreIdentity = "https://github.com/acme/app/.github/workflows/deploy.yml@refs/heads/main"
	sigstoreRepo     = "acme/app"
)

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
		"FAKE_COSIGN_REPOSITORY":   sigstoreRepo,
		"GITHUB_ACTIONS":           "true",
		permit.RepositoryEnv:       sigstoreRepo,
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
	passingDeploy(fp)
	return h, fp
}

// passingDeploy makes fp's deploy pass as a full deploy.
func passingDeploy(fp *fakeProvider) {
	fp.deploy = func(providers.Session, providers.ChangeSet, providers.DeployOptions) (providers.DeployResult, error) {
		return providers.DeployResult{Passed: true, ID: "0Af000000000009AAA", ComponentsDeployed: 1, Facts: map[string]any{}}, nil
	}
}

func (h *harness) permitVerify() int {
	return Execute(h.app, []string{"permit", "verify", "--target", "prod", "--dir", h.dir, "--ci", "github"})
}

// sigstoreTree makes the harness project a real git repository with everything but .glassops
// committed, and points GITHUB_SHA at that commit, so gate signs the real HEAD and deploy's tree
// check runs the real git rev-parse and git status. It returns the commit.
func sigstoreTree(t *testing.T, h *harness) string {
	t.Helper()
	if _, err := osexec.LookPath("git"); err != nil {
		t.Skip("git is not on the PATH")
	}
	h.write(t, ".gitignore", ".glassops/\n")
	h.write(t, "src/a.txt", "one\n")
	gitIn(t, h.dir, "init", "-q", "-b", "main")
	gitIn(t, h.dir, "config", "user.email", "test@example.invalid")
	gitIn(t, h.dir, "config", "user.name", "test")
	gitIn(t, h.dir, "config", "core.autocrlf", "false")
	gitIn(t, h.dir, "add", "-A")
	gitIn(t, h.dir, "commit", "-q", "-m", "one")
	head := gitIn(t, h.dir, "rev-parse", "HEAD")
	h.env["GITHUB_SHA"] = head
	fake := h.app.Run
	h.app.Run = func(dir, name string, args ...string) ([]byte, error) {
		if name == "git" && len(args) > 0 && (args[0] == "status" || (len(args) == 2 && args[0] == "rev-parse" && args[1] == "HEAD")) {
			cmd := osexec.Command("git", args...)
			cmd.Dir = dir
			return cmd.Output()
		}
		return fake(dir, name, args...)
	}
	return head
}

func (h *harness) deployProd() int {
	return Execute(h.app, []string{"deploy", "--target", "prod", "--dir", h.dir, "--ci", "github"})
}

func TestGateSigstoreRoundTrip(t *testing.T) {
	fakeCosignOnPath(t)
	h, fp := sigstoreGate(t)
	head := sigstoreTree(t, h)
	if code := h.gate("--target", "prod"); code != exitcode.OK {
		t.Fatalf("gate exit %d, output:\n%s", code, h.out)
	}
	p := readPermit(t, h)
	if !p.Allow || p.Signature.Alg != permit.SigstoreAlgorithm || p.Signature.Value == "" {
		t.Fatalf("PERMIT: allow %v, signature %+v", p.Allow, p.Signature.Alg)
	}
	// A Sigstore PERMIT binds the commit and the change set as an HMAC one does.
	cs, err := providers.ReadChangeSet(filepath.Join(h.dir, ".glassops", "plan.json"))
	if err != nil {
		t.Fatal(err)
	}
	if p.Commit != head || p.ChangeSetSHA256 == "" || p.ChangeSetSHA256 != permit.ChangeSetDigest(cs) {
		t.Fatalf("PERMIT commit %q change_set_sha256 %q, want %q and %q", p.Commit, p.ChangeSetSHA256, head, permit.ChangeSetDigest(cs))
	}
	if code := h.permitVerify(); code != exitcode.OK {
		t.Fatalf("permit verify exit %d, output:\n%s", code, h.out)
	}
	if r := h.result(t, "permit-verify"); checkStatus(r, "permit") != run.StatusPass {
		t.Fatalf("permit verify: %+v", r.Checks)
	}
	// deploy checks the same signature, then the tree, before it reaches the target.
	if code := h.deployProd(); code != exitcode.OK {
		t.Fatalf("deploy exit %d, output:\n%s", code, h.out)
	}
	r := h.result(t, "deploy")
	if checkStatus(r, "permit") != run.StatusPass || checkStatus(r, "tree") != run.StatusPass {
		t.Fatalf("deploy checks: %+v", r.Checks)
	}
	if fp.deployCalls != 1 {
		t.Fatalf("deploy calls %d, output:\n%s", fp.deployCalls, h.out)
	}
}

// TestSigstoreDeployRefusesAnotherTree pins F2 for a Sigstore PERMIT: a good PERMIT verifies, and
// deploy still refuses when the checkout is not its commit or a tracked file is modified.
func TestSigstoreDeployRefusesAnotherTree(t *testing.T) {
	cases := []struct {
		name    string
		change  func(t *testing.T, h *harness)
		contain string
	}{
		{"modified tracked file", func(t *testing.T, h *harness) { h.write(t, "src/a.txt", "two\n") }, "src/a.txt"},
		{"another commit", func(t *testing.T, h *harness) {
			h.write(t, "src/a.txt", "two\n")
			gitIn(t, h.dir, "commit", "-q", "-am", "two")
		}, "git HEAD is"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			fakeCosignOnPath(t)
			h, fp := sigstoreGate(t)
			sigstoreTree(t, h)
			if code := h.gate("--target", "prod"); code != exitcode.OK {
				t.Fatalf("gate exit %d, output:\n%s", code, h.out)
			}
			tc.change(t, h)
			if code := h.deployProd(); code != exitcode.Error {
				t.Fatalf("deploy exit %d, want 2. output:\n%s", code, h.out)
			}
			r := h.result(t, "deploy")
			if checkStatus(r, "permit") != run.StatusPass || !strings.Contains(detailOf(r, "tree"), tc.contain) {
				t.Fatalf("deploy checks: %+v", r.Checks)
			}
			if fp.deployCalls != 0 {
				t.Fatal("deploy reached the target from another tree")
			}
		})
	}
}

// TestSigstorePermitEndsWithTheWindow pins F7 for a Sigstore PERMIT: gate caps expires_at at the
// window end, the signature covers it, and deploy after the window closes is refused.
func TestSigstorePermitEndsWithTheWindow(t *testing.T) {
	at := func(hh, mm int) func() time.Time {
		return func() time.Time { return time.Date(2026, 10, 26, hh, mm, 0, 0, time.UTC) }
	}
	fakeCosignOnPath(t)
	old := gateRules
	t.Cleanup(func() { gateRules = old })
	gateRules = func() []policy.Rule { return []policy.Rule{fixedRule("plan_not_empty", policy.Pass)} }
	cfg := strings.Replace(sigstoreConfig, "  permit_ttl: 90m\n",
		"  permit_ttl: 90m\n  windows:\n    prod:\n      - { days: [mon], from: \"09:00\", to: \"12:30\", tz: UTC }\n", 1)
	if cfg == sigstoreConfig {
		t.Fatal("no permit_ttl line to add the window after")
	}
	h, fp := gateHarness(t, cfg)
	delete(h.env, permit.KeyEnv)
	passingDeploy(fp)
	sigstoreTree(t, h)
	h.app.Now = at(12, 0)
	if code := h.gate("--target", "prod"); code != exitcode.OK {
		t.Fatalf("gate exit %d, output:\n%s", code, h.out)
	}
	p := readPermit(t, h)
	if p.Signature.Alg != permit.SigstoreAlgorithm || !p.ExpiresAt.Equal(at(12, 30)()) {
		t.Fatalf("PERMIT %s expires %s, want sigstore and 12:30", p.Signature.Alg, p.ExpiresAt.UTC().Format(time.RFC3339))
	}
	h.app.Now = at(12, 31)
	if code := h.deployProd(); code == exitcode.OK || fp.deployCalls != 0 {
		t.Fatalf("deploy after the window exit %d, calls %d. output:\n%s", code, fp.deployCalls, h.out)
	}
	if d := detailOf(h.result(t, "deploy"), "permit"); !strings.Contains(d, "expired") {
		t.Fatalf("permit detail %q, want expired", d)
	}
	h.app.Now = at(12, 29)
	if code := h.deployProd(); code != exitcode.OK || fp.deployCalls != 1 {
		t.Fatalf("deploy inside the window exit %d, calls %d. output:\n%s", code, fp.deployCalls, h.out)
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
		{"another repository, same reusable workflow identity", func(t *testing.T, h *harness) {
			t.Setenv("FAKE_COSIGN_REPOSITORY", "evil/fork")
			h.gate("--target", "prod")
		}, "GitHub Workflow Repository"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			fakeCosignOnPath(t)
			h, fp := sigstoreGate(t)
			sigstoreTree(t, h)
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
			h.deployProd()
			if fp.deployCalls != 0 {
				t.Fatal("deploy reached the target with a refused PERMIT")
			}
			if r := h.result(t, "deploy"); checkStatus(r, "permit") == run.StatusPass || checkStatus(r, "tree") != "" {
				t.Fatalf("deploy was not refused at the permit check: %+v", r.Checks)
			}
		})
	}
}

// withRepository is cfg with policy.permit.repository set to repo.
func withRepository(t *testing.T, cfg, repo string) string {
	t.Helper()
	const issuer = "    issuer: https://token.actions.githubusercontent.com\n"
	out := strings.Replace(cfg, issuer, issuer+"    repository: "+repo+"\n", 1)
	if out == cfg {
		t.Fatal("the config has no issuer line to add the repository after")
	}
	return out
}

func TestPermitVerifySigstoreRepositoryOffGitHub(t *testing.T) {
	cases := []struct {
		name string
		cfg  string
		want int
		key  string
	}{
		// Off GitHub Actions, GITHUB_REPOSITORY is not trusted and the config names no repository.
		{"no repository is a config error", sigstoreConfig, exitcode.Error, "policy.permit.repository"},
		{"the right repository verifies", withRepository(t, sigstoreConfig, sigstoreRepo), exitcode.OK, ""},
		{"another repository is refused", withRepository(t, sigstoreConfig, "acme/other"), exitcode.Fail, ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			fakeCosignOnPath(t)
			old := gateRules
			t.Cleanup(func() { gateRules = old })
			gateRules = func() []policy.Rule { return []policy.Rule{fixedRule("plan_not_empty", policy.Pass)} }
			h, _ := gateHarness(t, tc.cfg)
			delete(h.env, permit.KeyEnv)
			// gate signs on GitHub with the pinned repository, which the config may not match.
			t.Setenv(permit.RepositoryEnv, sigstoreRepo)
			if code := h.gate("--target", "prod"); code != exitcode.OK {
				t.Fatalf("gate exit %d, output:\n%s", code, h.out)
			}
			t.Setenv("GITHUB_ACTIONS", "")
			t.Setenv(permit.RepositoryEnv, "")
			if code := h.permitVerify(); code != tc.want {
				t.Fatalf("permit verify exit %d, want %d. output:\n%s", code, tc.want, h.out)
			}
			r := h.result(t, "permit-verify")
			if d := detailOf(r, "key"); tc.key != "" && !strings.Contains(d, tc.key) {
				t.Fatalf("key detail %q, want %q", d, tc.key)
			}
			if tc.want == exitcode.Fail && !strings.Contains(detailOf(r, "permit"), "GitHub Workflow Repository") {
				t.Fatalf("permit detail %q", detailOf(r, "permit"))
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
	gh := map[string]string{"GITHUB_ACTIONS": "true", "GITHUB_EVENT_NAME": "push", permit.RepositoryEnv: sigstoreRepo,
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
			map[string]run.Status{"tool.cosign": run.StatusPass, "permit.sigstore": run.StatusPass, "permit.identity": run.StatusPass, "permit.repository": run.StatusPass},
			map[string]string{"tool.cosign": "v2.4.1", "permit.identity": sigstoreIdentity, "permit.repository": "acme/app (from GITHUB_REPOSITORY)"}},
		{"repository from the config", gh, true, withRepository(t, sigstoreConfig, sigstoreRepo),
			map[string]run.Status{"permit.repository": run.StatusPass},
			map[string]string{"permit.repository": "acme/app (from policy.permit.repository)"}},
		{"configured repository is not this one", gh, true, withRepository(t, sigstoreConfig, "acme/other"),
			map[string]run.Status{"permit.repository": run.StatusPass},
			map[string]string{"permit.repository": "warning: policy.permit.repository is acme/other but this run is in acme/app"}},
		{"local without a repository", map[string]string{permit.RepositoryEnv: sigstoreRepo}, true, sigstoreConfig,
			map[string]run.Status{"permit.repository": run.StatusPass},
			map[string]string{"permit.repository": "warning: no repository to pin"}},
		{"local with a repository", map[string]string{}, true, withRepository(t, sigstoreConfig, sigstoreRepo),
			map[string]run.Status{"permit.repository": run.StatusPass},
			map[string]string{"permit.repository": "acme/app (from policy.permit.repository)"}},
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
		{"hmac is silent", gh, false, gateConfig, map[string]run.Status{"tool.cosign": "", "permit.sigstore": "", "permit.identity": "", "permit.repository": ""}, nil},
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

// syntheticCallerWorkflow is a consumer workflow that calls the reusable deploy workflow.
const syntheticCallerWorkflow = `name: deploy
on: { push: { branches: [main] } }
jobs:
  deploy:
    uses: glassops-platform/glassops/.github/workflows/glassops-deploy.yml@v1 # pinned
    secrets: inherit
`

func TestDoctorSigstoreReusableWorkflowIdentity(t *testing.T) {
	const reusable = "https://github.com/glassops-platform/glassops/.github/workflows/glassops-deploy.yml@"
	cases := []struct {
		name, identity, status, contain string
	}{
		{"caller's own workflow", sigstoreIdentity, "warning",
			"this project calls glassops-platform/glassops/.github/workflows/glassops-deploy.yml@v1, and a PERMIT signed there carries the identity " + reusable + "<full ref>"},
		{"reusable at the called tag", reusable + "refs/tags/v1", "ok", reusable + "refs/tags/v1"},
		{"reusable at another tag", reusable + "refs/tags/v2", "warning", "identity names the reusable workflow at refs/tags/v2 but this project calls"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			h, _ := gateHarness(t, strings.Replace(sigstoreConfig, sigstoreIdentity, tc.identity, 1))
			h.write(t, ".github/workflows/deploy.yml", syntheticCallerWorkflow)
			h.env = map[string]string{"GITHUB_ACTIONS": "true", permit.RepositoryEnv: sigstoreRepo}
			h.tools["cosign version"] = "GitVersion:    v2.5.3\n"
			h.exec("doctor")
			d := detailOf(h.result(t, "doctor"), "permit.identity")
			if !strings.Contains(d, tc.contain) || strings.HasPrefix(d, "warning:") != (tc.status == "warning") {
				t.Fatalf("permit.identity detail %q, want %s containing %q", d, tc.status, tc.contain)
			}
		})
	}
}

func TestReusableWorkflowRefs(t *testing.T) {
	dir := t.TempDir()
	wf := filepath.Join(dir, ".github", "workflows")
	if err := os.MkdirAll(wf, 0o755); err != nil {
		t.Fatal(err)
	}
	files := map[string]string{
		"a.yml":  syntheticCallerWorkflow,
		"b.yaml": "jobs:\n  x:\n    uses: \"glassops-platform/glassops/.github/workflows/glassops-deploy.yml@refs/heads/main\"\n  y:\n    uses: 'glassops-platform/glassops/.github/workflows/glassops-deploy.yml@v1'\n",
		"c.yml":  "jobs:\n  z:\n    steps:\n      - uses: actions/checkout@v4\n      # uses: glassops-platform/glassops/.github/workflows/glassops-deploy.yml@commented\n    uses: glassops-platform/glassops/.github/workflows/other.yml@v1\n",
		"d.txt":  "uses: glassops-platform/glassops/.github/workflows/glassops-deploy.yml@ignored\n",
	}
	for name, body := range files {
		if err := os.WriteFile(filepath.Join(wf, name), []byte(body), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	got := reusableWorkflowRefs(dir)
	if want := []string{"refs/heads/main", "v1"}; strings.Join(got, ",") != strings.Join(want, ",") {
		t.Fatalf("refs = %v, want %v", got, want)
	}
	if got := reusableWorkflowRefs(filepath.Join(dir, "missing")); len(got) != 0 {
		t.Fatalf("missing dir: %v", got)
	}
}

func TestRefMatches(t *testing.T) {
	for ref, want := range map[string]bool{"refs/tags/v1": true, "v1": true, "refs/tags/v10": false, "refs/heads/v1x": false} {
		if got := refMatches(ref, []string{"v1"}); got != want {
			t.Errorf("%s: %v", ref, got)
		}
	}
}
