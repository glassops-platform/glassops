package permit

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/glassops-platform/glassops/internal/core/config"
	"github.com/glassops-platform/glassops/internal/core/policy"
)

// The test binary doubles as a fake cosign: run with GLASSOPS_FAKE_COSIGN=1 it answers
// "version", "sign-blob" and "verify-blob" instead of running tests. Its bundle is synthetic,
// JSON naming the blob's SHA-256 and the identity and issuer from FAKE_COSIGN_IDENTITY and
// FAKE_COSIGN_ISSUER. Real bundles hold a certificate and a Rekor entry instead [unverified
// shape, the orchestrator's probe run captures one].
func init() {
	if os.Getenv("GLASSOPS_FAKE_COSIGN") == "1" {
		os.Exit(fakeCosign(os.Args[1:], os.Stdout, os.Stderr))
	}
}

type fakeBundle struct {
	Synthetic bool   `json:"synthetic"`
	SHA256    string `json:"sha256"`
	Identity  string `json:"identity"`
	Issuer    string `json:"issuer"`
}

func fakeCosign(args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		fmt.Fprintln(stderr, "Error: no command")
		return 1
	}
	flags, yes, pos := map[string]string{}, false, []string{}
	for i := 1; i < len(args); i++ {
		switch a := args[i]; a {
		case "--yes":
			yes = true
		case "--bundle", "--certificate-identity", "--certificate-oidc-issuer":
			if i+1 < len(args) {
				flags[a] = args[i+1]
				i++
			}
		default:
			if strings.HasPrefix(a, "-") {
				fmt.Fprintf(stderr, "Error: unknown flag: %s\n", a)
				return 1
			}
			pos = append(pos, a)
		}
	}
	fail := func(msg string) int { fmt.Fprintln(stderr, "Error: "+msg); return 1 }
	switch args[0] {
	case "version":
		fmt.Fprintln(stdout, "GitVersion:    v2.4.1")
		return 0
	case "sign-blob":
		switch {
		case !yes:
			return fail("signing needs confirmation and stdin is not a terminal")
		case os.Getenv(OIDCRequestTokenEnv) == "":
			return fail("getting signer: no OIDC token for user@example.com")
		case os.Getenv("FAKE_COSIGN_FAIL") != "":
			return fail("signing: rekor is unavailable")
		case len(pos) != 1 || flags["--bundle"] == "":
			return fail("sign-blob needs one blob and --bundle")
		}
		blob, err := os.ReadFile(pos[0])
		if err != nil {
			return fail(err.Error())
		}
		sum := sha256.Sum256(blob)
		b, _ := json.Marshal(fakeBundle{true, hex.EncodeToString(sum[:]), os.Getenv("FAKE_COSIGN_IDENTITY"), os.Getenv("FAKE_COSIGN_ISSUER")})
		if err := os.WriteFile(flags["--bundle"], b, 0o600); err != nil {
			return fail(err.Error())
		}
		fmt.Fprintln(stdout, base64.StdEncoding.EncodeToString(sum[:]))
		return 0
	case "verify-blob":
		if len(pos) != 1 || flags["--bundle"] == "" || flags["--certificate-identity"] == "" || flags["--certificate-oidc-issuer"] == "" {
			return fail("verify-blob needs one blob, --bundle, --certificate-identity and --certificate-oidc-issuer")
		}
		blob, err := os.ReadFile(pos[0])
		if err != nil {
			return fail(err.Error())
		}
		raw, err := os.ReadFile(flags["--bundle"])
		if err != nil {
			return fail(err.Error())
		}
		var b fakeBundle
		if err := json.Unmarshal(raw, &b); err != nil || !b.Synthetic {
			return fail("bundle could not be parsed")
		}
		sum := sha256.Sum256(blob)
		switch {
		case b.SHA256 != hex.EncodeToString(sum[:]):
			return fail("invalid signature when validating ASN.1 encoded signature")
		case b.Identity != flags["--certificate-identity"]:
			return fail("none of the expected identities matched what was in the certificate")
		case b.Issuer != flags["--certificate-oidc-issuer"]:
			return fail("expected oidc issuer not found in certificate")
		}
		fmt.Fprintln(stderr, "Verified OK")
		return 0
	}
	return fail("unknown command " + args[0])
}

const (
	testIdentity = "https://github.com/acme/app/.github/workflows/deploy.yml@refs/heads/main"
	testIssuer   = GitHubIssuer
)

func sigstoreCfg(identity, issuer string) config.PermitSigning {
	return config.PermitSigning{Signing: config.SigningSigstore, Identity: identity, Issuer: issuer}
}

// onGitHub makes this process look like a GitHub Actions job with id-token: write and makes
// cosign the fake. The fake signs as testIdentity from testIssuer.
func onGitHub(t *testing.T) {
	t.Helper()
	for k, v := range map[string]string{
		"GLASSOPS_FAKE_COSIGN":   "1",
		"FAKE_COSIGN_IDENTITY":   testIdentity,
		"FAKE_COSIGN_ISSUER":     testIssuer,
		"FAKE_COSIGN_FAIL":       "",
		"GITHUB_ACTIONS":         "true",
		"BITBUCKET_BUILD_NUMBER": "",
		"GITLAB_CI":              "",
		OIDCRequestURLEnv:        "https://example.invalid/token",
		OIDCRequestTokenEnv:      "syntheticRequestToken",
	} {
		t.Setenv(k, v)
	}
	self, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	old := cosignLookPath
	t.Cleanup(func() { cosignLookPath = old })
	cosignLookPath = func(string) (string, error) { return self, nil }
}

func sigstorePermit(now time.Time) Permit {
	return Permit{
		Schema: Schema, ID: "abc123", Allow: true, IssuedAt: now, ExpiresAt: now.Add(time.Hour),
		Target:     Target{Name: "prod", Tier: "prod", ID: "00D000000000001AAA", Kind: "production"},
		Commit:     "20459158b4f16da4cddc816ac7c1d695177af196",
		PlanSHA256: "aa", ConfigSHA256: "bb",
		Checks: []policy.Result{{Rule: "plan_not_empty", Outcome: policy.Pass, Detail: "synthetic"}},
	}
}

func sigstoreExpect(now time.Time) Expect {
	return Expect{TargetName: "prod", TargetID: "00D000000000001AAA", Commit: "20459158b4f16da4cddc816ac7c1d695177af196",
		PlanSHA256: "aa", ConfigSHA256: "bb", Now: now.Add(time.Minute)}
}

// signedPermit signs a synthetic PERMIT with the fake cosign through SignerFor.
func signedPermit(t *testing.T) (Permit, time.Time) {
	t.Helper()
	now := time.Date(2026, 10, 31, 12, 0, 0, 0, time.UTC)
	cfg := &config.Config{Policy: config.Policy{Permit: sigstoreCfg(testIdentity, testIssuer)}}
	s, err := SignerFor(cfg, noHMAC(t))
	if err != nil {
		t.Fatal(err)
	}
	p := sigstorePermit(now)
	if err := Sign(&p, s); err != nil {
		t.Fatal(err)
	}
	return p, now
}

func noHMAC(t *testing.T) Keys {
	return Keys{HMACKey: func() ([]byte, error) { t.Fatal("sigstore read the HMAC key"); return nil, nil }}
}

func verifierFor(t *testing.T, identity, issuer string) Signer {
	t.Helper()
	v, err := VerifierFor(&config.Config{Policy: config.Policy{Permit: sigstoreCfg(identity, issuer)}}, noHMAC(t))
	if err != nil {
		t.Fatal(err)
	}
	return v
}

func TestSigstoreRoundTrip(t *testing.T) {
	onGitHub(t)
	p, now := signedPermit(t)
	if p.Signature.Alg != SigstoreAlgorithm || p.Signature.KeyID != "" {
		t.Fatalf("signature: %+v", p.Signature)
	}
	raw, err := base64.StdEncoding.DecodeString(p.Signature.Value)
	if err != nil || !json.Valid(raw) {
		t.Fatalf("the signature is not a base64 bundle: %v", err)
	}
	if err := Verify(p, verifierFor(t, testIdentity, testIssuer), sigstoreExpect(now)); err != nil {
		t.Fatalf("round trip: %v", err)
	}
	// A verifier needs no OIDC token, and runs outside any CI.
	t.Setenv("GITHUB_ACTIONS", "")
	t.Setenv(OIDCRequestTokenEnv, "")
	if err := Verify(p, verifierFor(t, testIdentity, testIssuer), sigstoreExpect(now)); err != nil {
		t.Fatalf("verify outside a CI: %v", err)
	}
}

func TestSigstoreRefusals(t *testing.T) {
	onGitHub(t)
	p, now := signedPermit(t)
	tampered := p
	tampered.PlanSHA256 = "cc"
	hmacSigned := sigstorePermit(now)
	if err := Sign(&hmacSigned, NewHMAC([]byte(strings.Repeat("k", MinKeyBytes)))); err != nil {
		t.Fatal(err)
	}
	notBase64, notBundle := p, p
	notBase64.Signature.Value = "%%%"
	notBundle.Signature.Value = base64.StdEncoding.EncodeToString([]byte("not json"))
	cases := []struct {
		name    string
		p       Permit
		v       Signer
		want    error
		contain string
	}{
		{"tampered PERMIT", tampered, verifierFor(t, testIdentity, testIssuer), ErrSignature, "invalid signature"},
		{"wrong identity", p, verifierFor(t, "https://github.com/acme/app/.github/workflows/other.yml@refs/heads/main", testIssuer), ErrSignature, "identities"},
		{"wrong issuer", p, verifierFor(t, testIdentity, "https://accounts.example.com"), ErrSignature, "issuer"},
		{"HMAC PERMIT", hmacSigned, verifierFor(t, testIdentity, testIssuer), ErrSignature, "algorithm"},
		{"value not base64", notBase64, verifierFor(t, testIdentity, testIssuer), ErrSignature, "base64"},
		{"value not a bundle", notBundle, verifierFor(t, testIdentity, testIssuer), ErrSignature, "not a sigstore bundle"},
		{"sigstore PERMIT, HMAC verifier", p, NewHMAC([]byte(strings.Repeat("k", MinKeyBytes))), ErrSignature, "algorithm"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := Verify(tc.p, tc.v, sigstoreExpect(now))
			if !errors.Is(err, tc.want) || !strings.Contains(err.Error(), tc.contain) {
				t.Fatalf("err = %v, want %v containing %q", err, tc.want, tc.contain)
			}
		})
	}
}

func TestSigstoreSignerRefusesBeforeCosign(t *testing.T) {
	cfg := &config.Config{Policy: config.Policy{Permit: sigstoreCfg(testIdentity, testIssuer)}}
	cases := []struct {
		name string
		env  map[string]string
		path bool
		want error
	}{
		{"not on GitHub", map[string]string{"GITHUB_ACTIONS": ""}, true, ErrSigstoreNotGitHub},
		{"GitLab", map[string]string{"GITHUB_ACTIONS": "", "GITLAB_CI": "true"}, true, ErrSigstoreNotGitHub},
		{"no OIDC request token", map[string]string{OIDCRequestTokenEnv: ""}, true, ErrNoOIDCToken},
		{"no OIDC request URL", map[string]string{OIDCRequestURLEnv: ""}, true, ErrNoOIDCToken},
		{"no cosign", nil, false, ErrNoCosign},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			onGitHub(t)
			for k, v := range tc.env {
				t.Setenv(k, v)
			}
			if !tc.path {
				cosignLookPath = func(string) (string, error) { return "", errors.New("not found") }
			}
			s, err := SignerFor(cfg, noHMAC(t))
			if !errors.Is(err, tc.want) || s != nil {
				t.Fatalf("SignerFor = %v, %v, want %v", s, err, tc.want)
			}
			if strings.Contains(err.Error(), "syntheticRequestToken") {
				t.Fatal("the OIDC request token reached an error")
			}
		})
	}
}

func TestSigstoreVerifierRefusals(t *testing.T) {
	cases := []struct {
		name string
		env  map[string]string
		path bool
		cfg  config.PermitSigning
		want string
	}{
		{"Bitbucket", map[string]string{"GITHUB_ACTIONS": "", "BITBUCKET_BUILD_NUMBER": "7"}, true, sigstoreCfg(testIdentity, testIssuer), "Bitbucket Pipelines"},
		{"GitLab", map[string]string{"GITHUB_ACTIONS": "", "GITLAB_CI": "true"}, true, sigstoreCfg(testIdentity, testIssuer), "GitLab CI"},
		{"no cosign", nil, false, sigstoreCfg(testIdentity, testIssuer), "cosign is not installed"},
		{"no identity", nil, true, sigstoreCfg("", testIssuer), "policy.permit.identity"},
		{"no issuer", nil, true, sigstoreCfg(testIdentity, " "), "policy.permit.issuer"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			onGitHub(t)
			for k, v := range tc.env {
				t.Setenv(k, v)
			}
			if !tc.path {
				cosignLookPath = func(string) (string, error) { return "", errors.New("not found") }
			}
			v, err := VerifierFor(&config.Config{Policy: config.Policy{Permit: tc.cfg}}, noHMAC(t))
			if err == nil || v != nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("VerifierFor = %v, %v, want %q", v, err, tc.want)
			}
		})
	}
}

func TestSigstoreSignFailures(t *testing.T) {
	onGitHub(t)
	if _, err := verifierFor(t, testIdentity, testIssuer).Sign([]byte("x")); err == nil {
		t.Fatal("a verifier signed")
	}
	cfg := &config.Config{Policy: config.Policy{Permit: sigstoreCfg(testIdentity, testIssuer)}}
	s, err := SignerFor(cfg, noHMAC(t))
	if err != nil {
		t.Fatal(err)
	}
	t.Setenv("FAKE_COSIGN_FAIL", "1")
	p := sigstorePermit(time.Now())
	err = Sign(&p, s)
	if err == nil || !strings.Contains(err.Error(), "cosign sign-blob failed: Error: signing: rekor is unavailable") {
		t.Fatalf("err = %v", err)
	}
	if p.Signature.Value != "" {
		t.Fatal("a failed sign left a signature")
	}
	// cosign's own error text is masked like any tool output.
	t.Setenv("FAKE_COSIGN_FAIL", "")
	t.Setenv(OIDCRequestTokenEnv, "")
	if _, err := s.Sign([]byte("x")); err == nil || strings.Contains(err.Error(), "user@example.com") || !strings.Contains(err.Error(), "u***@example.com") {
		t.Fatalf("err = %v", err)
	}
}

func TestLastLine(t *testing.T) {
	if got := lastLine("a\n\nError: b  \n\n"); got != "Error: b" {
		t.Fatalf("got %q", got)
	}
	if got := lastLine(strings.Repeat("x", 400)); len(got) != 300 {
		t.Fatalf("len %d", len(got))
	}
}

func TestOtherCI(t *testing.T) {
	for env, want := range map[string]string{
		"":                         "",
		"GITHUB_ACTIONS=true":      "",
		"BITBUCKET_BUILD_NUMBER=3": "Bitbucket Pipelines",
		"GITLAB_CI=true":           "GitLab CI",
	} {
		k, v, _ := strings.Cut(env, "=")
		got := OtherCI(func(name string) string {
			if name == k {
				return v
			}
			return ""
		})
		if got != want {
			t.Errorf("%q: got %q, want %q", env, got, want)
		}
	}
}
