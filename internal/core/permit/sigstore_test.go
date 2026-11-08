package permit

import (
	"crypto/sha256"
	"encoding/base64"
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
// "version", "sign-blob" and "verify-blob" instead of running tests. Its bundle is synthetic
// but has the legacy shape cosign v2.5.3 "sign-blob --bundle" writes, {base64Signature, cert,
// rekorBundle} with no mediaType [confirmed by the orchestrator's probe, log P56L-023]. In the
// synthetic bundle base64Signature is the blob's SHA-256, and cert is base64 JSON of the
// identity, issuer and repository from FAKE_COSIGN_IDENTITY, FAKE_COSIGN_ISSUER and
// FAKE_COSIGN_REPOSITORY. A real cert is a base64 PEM certificate whose extensions hold them.
func init() {
	if os.Getenv("GLASSOPS_FAKE_COSIGN") == "1" {
		os.Exit(fakeCosign(os.Args[1:], os.Stdout, os.Stderr))
	}
}

// syntheticLegacyBundle has the field names of a cosign v2 legacy bundle.
type syntheticLegacyBundle struct {
	Base64Signature string          `json:"base64Signature"`
	Cert            string          `json:"cert"`
	RekorBundle     json.RawMessage `json:"rekorBundle"`
}

// syntheticCertClaims is what the synthetic cert carries in place of Fulcio's extensions.
type syntheticCertClaims struct {
	Identity   string `json:"identity"`
	Issuer     string `json:"issuer"`
	Repository string `json:"repository"`
}

// syntheticRekorBundle has the shape of a legacy rekorBundle, with no real log entry.
const syntheticRekorBundle = `{"SignedEntryTimestamp":"c3ludGhldGlj","Payload":{"body":"c3ludGhldGlj","integratedTime":1762600000,"logIndex":1,"logID":"synthetic"}}`

func syntheticBundleFor(blob []byte, c syntheticCertClaims) []byte {
	sum := sha256.Sum256(blob)
	claims, _ := json.Marshal(c)
	b, _ := json.Marshal(syntheticLegacyBundle{
		Base64Signature: base64.StdEncoding.EncodeToString(sum[:]),
		Cert:            base64.StdEncoding.EncodeToString(claims),
		RekorBundle:     json.RawMessage(syntheticRekorBundle),
	})
	return b
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
		case "--bundle", "--certificate-identity", "--certificate-oidc-issuer", "--certificate-github-workflow-repository":
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
		b := syntheticBundleFor(blob, syntheticCertClaims{os.Getenv("FAKE_COSIGN_IDENTITY"), os.Getenv("FAKE_COSIGN_ISSUER"), os.Getenv("FAKE_COSIGN_REPOSITORY")})
		if os.Getenv("FAKE_COSIGN_ODD_BUNDLE") != "" {
			b = []byte(`{"synthetic":true}`)
		}
		if err := os.WriteFile(flags["--bundle"], b, 0o600); err != nil {
			return fail(err.Error())
		}
		sum := sha256.Sum256(blob)
		fmt.Fprintln(stdout, base64.StdEncoding.EncodeToString(sum[:]))
		return 0
	case "verify-blob":
		if len(pos) != 1 || flags["--bundle"] == "" || flags["--certificate-identity"] == "" || flags["--certificate-oidc-issuer"] == "" || flags["--certificate-github-workflow-repository"] == "" {
			return fail("verify-blob needs one blob, --bundle, --certificate-identity, --certificate-oidc-issuer and --certificate-github-workflow-repository")
		}
		blob, err := os.ReadFile(pos[0])
		if err != nil {
			return fail(err.Error())
		}
		raw, err := os.ReadFile(flags["--bundle"])
		if err != nil {
			return fail(err.Error())
		}
		var b syntheticLegacyBundle
		var c syntheticCertClaims
		if err := json.Unmarshal(raw, &b); err != nil {
			return fail("bundle could not be parsed")
		}
		if certJSON, err := base64.StdEncoding.DecodeString(b.Cert); err != nil || json.Unmarshal(certJSON, &c) != nil {
			return fail("could not parse the certificate")
		}
		sum := sha256.Sum256(blob)
		switch {
		case b.Base64Signature != base64.StdEncoding.EncodeToString(sum[:]):
			return fail("invalid signature when validating ASN.1 encoded signature")
		case c.Identity != flags["--certificate-identity"]:
			return fail("none of the expected identities matched what was in the certificate")
		case c.Issuer != flags["--certificate-oidc-issuer"]:
			return fail("expected oidc issuer not found in certificate")
		case c.Repository != flags["--certificate-github-workflow-repository"]:
			return fail("expected GitHub Workflow Repository not found in certificate")
		}
		fmt.Fprintln(stderr, "Verified OK")
		return 0
	}
	return fail("unknown command " + args[0])
}

const (
	testIdentity = "https://github.com/glassops-platform/glassops/.github/workflows/glassops-deploy.yml@refs/tags/v1"
	testIssuer   = GitHubIssuer
	testRepo     = "acme/app"
)

func sigstoreCfg(identity, issuer string) config.PermitSigning {
	return config.PermitSigning{Signing: config.SigningSigstore, Identity: identity, Issuer: issuer}
}

func sigstoreCfgRepo(repo string) config.PermitSigning {
	p := sigstoreCfg(testIdentity, testIssuer)
	p.Repository = repo
	return p
}

// onGitHub makes this process look like a GitHub Actions job in testRepo with id-token: write
// and makes cosign the fake. The fake signs as testIdentity from testIssuer in testRepo.
func onGitHub(t *testing.T) {
	t.Helper()
	for k, v := range map[string]string{
		"GLASSOPS_FAKE_COSIGN":   "1",
		"FAKE_COSIGN_IDENTITY":   testIdentity,
		"FAKE_COSIGN_ISSUER":     testIssuer,
		"FAKE_COSIGN_REPOSITORY": testRepo,
		"FAKE_COSIGN_FAIL":       "",
		"GITHUB_ACTIONS":         "true",
		RepositoryEnv:            testRepo,
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
	return verifierWith(t, sigstoreCfg(identity, issuer))
}

func verifierWith(t *testing.T, p config.PermitSigning) Signer {
	t.Helper()
	v, err := VerifierFor(&config.Config{Policy: config.Policy{Permit: p}}, noHMAC(t))
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
	if err != nil || checkBundle(raw) != nil {
		t.Fatalf("the signature is not a base64 legacy bundle: %v, %v", err, checkBundle(raw))
	}
	var shape map[string]json.RawMessage
	if err := json.Unmarshal(raw, &shape); err != nil || len(shape) != 3 || shape["mediaType"] != nil {
		t.Fatalf("the synthetic bundle is not the legacy shape: %s", raw)
	}
	// On GitHub the verifier pins GITHUB_REPOSITORY.
	if err := Verify(p, verifierFor(t, testIdentity, testIssuer), sigstoreExpect(now)); err != nil {
		t.Fatalf("round trip: %v", err)
	}
	// A verifier needs no OIDC token, and runs outside any CI with the repository in the config.
	t.Setenv("GITHUB_ACTIONS", "")
	t.Setenv(RepositoryEnv, "")
	t.Setenv(OIDCRequestTokenEnv, "")
	if err := Verify(p, verifierWith(t, sigstoreCfgRepo(testRepo)), sigstoreExpect(now)); err != nil {
		t.Fatalf("verify outside a CI: %v", err)
	}
}

func TestSigstoreRepositoryPinning(t *testing.T) {
	onGitHub(t)
	p, now := signedPermit(t)
	// The same reusable workflow identity, signed in another repository.
	t.Setenv("FAKE_COSIGN_REPOSITORY", "evil/fork")
	other, _ := signedPermit(t)
	t.Setenv("FAKE_COSIGN_REPOSITORY", testRepo)

	if err := Verify(p, verifierWith(t, sigstoreCfgRepo(testRepo)), sigstoreExpect(now)); err != nil {
		t.Fatalf("the right repository was refused: %v", err)
	}
	for name, tc := range map[string]struct {
		p   Permit
		cfg config.PermitSigning
	}{
		"PERMIT from another repo, pinned by GITHUB_REPOSITORY": {other, sigstoreCfg(testIdentity, testIssuer)},
		"PERMIT from another repo, pinned by the config":        {other, sigstoreCfgRepo(testRepo)},
		"config pins another repo":                              {p, sigstoreCfgRepo("acme/other")},
	} {
		t.Run(name, func(t *testing.T) {
			err := Verify(tc.p, verifierWith(t, tc.cfg), sigstoreExpect(now))
			if !errors.Is(err, ErrSignature) || !strings.Contains(err.Error(), "GitHub Workflow Repository") {
				t.Fatalf("err = %v, want a refused repository", err)
			}
		})
	}
	// The config wins over GITHUB_REPOSITORY.
	t.Setenv(RepositoryEnv, "acme/other")
	if err := Verify(p, verifierWith(t, sigstoreCfgRepo(testRepo)), sigstoreExpect(now)); err != nil {
		t.Fatalf("the configured repository did not win: %v", err)
	}
}

func TestPinnedRepository(t *testing.T) {
	env := func(kv map[string]string) func(string) string { return func(k string) string { return kv[k] } }
	gh := map[string]string{"GITHUB_ACTIONS": "true", RepositoryEnv: "acme/app"}
	cases := []struct {
		name               string
		repo               string
		env                map[string]string
		want, source, fail string
	}{
		{"GitHub default", "", gh, "acme/app", RepositoryEnv, ""},
		{"config wins on GitHub", "acme/other", gh, "acme/other", "policy.permit.repository", ""},
		{"config off GitHub", "acme/app", nil, "acme/app", "policy.permit.repository", ""},
		{"missing off GitHub", "", nil, "", "", "policy.permit.repository"},
		{"GITHUB_REPOSITORY ignored off GitHub", "", map[string]string{RepositoryEnv: "acme/app"}, "", "", "policy.permit.repository"},
		{"GitHub without GITHUB_REPOSITORY", "", map[string]string{"GITHUB_ACTIONS": "true"}, "", "", "policy.permit.repository"},
		{"not owner/name", "app", nil, "", "", "is not owner/name"},
		{"bad GITHUB_REPOSITORY", "", map[string]string{"GITHUB_ACTIONS": "true", RepositoryEnv: "a/b/c"}, "", "", "GITHUB_REPOSITORY \"a/b/c\" is not owner/name"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			p := sigstoreCfgRepo(tc.repo)
			got, source, err := PinnedRepository(p, env(tc.env))
			if tc.fail != "" {
				if err == nil || !strings.Contains(err.Error(), tc.fail) {
					t.Fatalf("err = %v, want %q", err, tc.fail)
				}
				return
			}
			if err != nil || got != tc.want || source != tc.source {
				t.Fatalf("got %q from %q, %v", got, source, err)
			}
		})
	}
}

func TestSigstoreVerifierNeedsRepositoryOffGitHub(t *testing.T) {
	onGitHub(t)
	t.Setenv("GITHUB_ACTIONS", "")
	t.Setenv(RepositoryEnv, "acme/app")
	v, err := VerifierFor(&config.Config{Policy: config.Policy{Permit: sigstoreCfg(testIdentity, testIssuer)}}, noHMAC(t))
	if !errors.Is(err, ErrNoRepository) || v != nil {
		t.Fatalf("VerifierFor = %v, %v, want ErrNoRepository", v, err)
	}
}

func TestCheckBundle(t *testing.T) {
	legacy := string(syntheticBundleFor([]byte("x"), syntheticCertClaims{testIdentity, testIssuer, testRepo}))
	// syntheticProtobufBundle has the top-level shape of a Sigstore protobuf bundle in JSON.
	syntheticProtobufBundle := `{"mediaType":"application/vnd.dev.sigstore.bundle.v0.3+json","verificationMaterial":{"certificate":{"rawBytes":"c3ludGhldGlj"},"tlogEntries":[]},"messageSignature":{"messageDigest":{"algorithm":"SHA2_256","digest":"c3ludGhldGlj"},"signature":"c3ludGhldGlj"}}`
	cases := []struct {
		name, data, fail string
	}{
		{"synthetic legacy", legacy, ""},
		{"synthetic protobuf", syntheticProtobufBundle, ""},
		{"protobuf v0.1", `{"mediaType":"application/vnd.dev.sigstore.bundle+json;version=0.1"}`, ""},
		{"other mediaType", `{"mediaType":"application/json"}`, "is not a Sigstore bundle"},
		{"mediaType not a string", `{"mediaType":7}`, "is not a Sigstore bundle"},
		{"empty object", `{}`, "neither a mediaType nor a base64Signature"},
		{"null", `null`, "not a JSON object"},
		{"array", `[]`, "not a JSON object"},
		{"legacy, signature not base64", `{"base64Signature":"%%","cert":"Yw==","rekorBundle":{"a":1}}`, "not base64"},
		{"legacy without cert", `{"base64Signature":"c2ln","rekorBundle":{"a":1}}`, "no signing certificate"},
		{"legacy without rekorBundle", `{"base64Signature":"c2ln","cert":"Yw=="}`, "no transparency log entry"},
		{"legacy with an empty rekorBundle", `{"base64Signature":"c2ln","cert":"Yw==","rekorBundle":{}}`, "no transparency log entry"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := checkBundle([]byte(tc.data))
			if tc.fail == "" && err != nil {
				t.Fatalf("refused: %v", err)
			}
			if tc.fail != "" && (err == nil || !strings.Contains(err.Error(), tc.fail)) {
				t.Fatalf("err = %v, want %q", err, tc.fail)
			}
		})
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
	// A bundle of neither known shape is refused before it reaches a PERMIT.
	t.Setenv("FAKE_COSIGN_FAIL", "")
	t.Setenv("FAKE_COSIGN_ODD_BUNDLE", "1")
	if err := Sign(&p, s); err == nil || !strings.Contains(err.Error(), "bundle GlassOps cannot read") || p.Signature.Value != "" {
		t.Fatalf("odd bundle: err = %v, signature %q", err, p.Signature.Value)
	}
	t.Setenv("FAKE_COSIGN_ODD_BUNDLE", "")
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
