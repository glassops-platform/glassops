package permit

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/glassops-platform/glassops/internal/core/config"
	"github.com/glassops-platform/glassops/internal/core/redact"
)

// SigstoreAlgorithm is the signature algorithm a sigstore PERMIT records. The signature value
// is the cosign bundle in standard base64.
const SigstoreAlgorithm = "sigstore-bundle"

// Cosign is the program the sigstore signer runs. It is looked up on PATH.
const Cosign = "cosign"

// GitHubIssuer is the OIDC issuer of GitHub Actions job tokens.
const GitHubIssuer = "https://token.actions.githubusercontent.com"

// The variables GitHub Actions sets when a job may request an OIDC token, which needs the
// id-token: write permission. cosign reads them itself, so GlassOps checks only that they are
// present and never reads their values into output.
const (
	OIDCRequestURLEnv   = "ACTIONS_ID_TOKEN_REQUEST_URL"
	OIDCRequestTokenEnv = "ACTIONS_ID_TOKEN_REQUEST_TOKEN"
)

// cosignTimeout bounds one cosign run. Keyless signing talks to Fulcio and Rekor, and
// verifying may refresh the Sigstore trust root.
const cosignTimeout = 3 * time.Minute

// Errors the sigstore signer returns before it runs cosign. Callers tell them apart with
// errors.Is. Each one stops gate before a PERMIT is written.
var (
	// ErrSigstoreNotGitHub means sigstore was chosen on a run where it is not supported.
	ErrSigstoreNotGitHub = errors.New("permit: sigstore signing works on GitHub Actions only")
	// ErrNoOIDCToken means the GitHub job cannot request an OIDC token.
	ErrNoOIDCToken = errors.New("permit: the job has no OIDC token for sigstore signing")
	// ErrNoCosign means cosign is not on PATH.
	ErrNoCosign = errors.New("permit: cosign is not installed")
	// ErrNoRepository means there is no source repository to pin a signature to. It is a
	// config error: set policy.permit.repository.
	ErrNoRepository = errors.New("permit: sigstore verifying needs policy.permit.repository")
)

// RepositoryEnv is the variable GitHub Actions sets to the running repository, as owner/name.
const RepositoryEnv = "GITHUB_REPOSITORY"

// repositoryPattern is what policy.permit.repository accepts, owner/name as GitHub writes it.
// schemas/config.json carries the same pattern.
var repositoryPattern = regexp.MustCompile(`^[A-Za-z0-9-]+/[A-Za-z0-9._-]+$`)

// PinnedRepository returns the repository a sigstore signature must come from, and where it
// came from: "policy.permit.repository" when the config names one, otherwise RepositoryEnv on
// GitHub Actions. Off GitHub Actions with no repository in the config it returns
// ErrNoRepository. The certificate identity names the workflow that signed, and every
// repository calling the same reusable workflow shares that identity, so the repository is
// what tells one consumer's PERMITs from another's.
func PinnedRepository(p config.PermitSigning, getenv func(string) string) (repo, source string, err error) {
	repo, source = strings.TrimSpace(p.Repository), "policy.permit.repository"
	if repo == "" && getenv("GITHUB_ACTIONS") == "true" {
		repo, source = strings.TrimSpace(getenv(RepositoryEnv)), RepositoryEnv
	}
	if repo == "" {
		return "", "", fmt.Errorf("%w (owner/name) off GitHub Actions, the repository whose workflow signed the PERMIT", ErrNoRepository)
	}
	if !repositoryPattern.MatchString(repo) {
		return "", "", fmt.Errorf("permit: %s %q is not owner/name", source, repo)
	}
	return repo, source, nil
}

// sigstoreGetenv and cosignLookPath are how the signer reads the environment and finds cosign.
// Tests replace them. cosign itself runs with the process environment, because it reads the
// OIDC request variables on its own.
var (
	sigstoreGetenv = os.Getenv
	cosignLookPath = exec.LookPath
)

// sigstoreSigner signs and verifies PERMITs with cosign, keyless. Sign runs "cosign sign-blob"
// with the job's OIDC token and returns the bundle. Verify runs "cosign verify-blob" and
// accepts only a certificate for identity issued by issuer to a workflow run in repository.
type sigstoreSigner struct {
	cosign     string
	identity   string
	issuer     string
	repository string
	// canSign is false for a verifier, which never needs an OIDC token.
	canSign bool
}

// newSigstoreSigner returns the keyless signer for cfg. It refuses a run that is not GitHub
// Actions, a job without an OIDC token (the id-token: write permission), a missing cosign and
// a repository that is not owner/name, in that order, so gate stops before it writes a PERMIT.
func newSigstoreSigner(cfg config.PermitSigning) (Signer, error) {
	if err := sigstoreIdentity(cfg); err != nil {
		return nil, err
	}
	if sigstoreGetenv("GITHUB_ACTIONS") != "true" {
		return nil, fmt.Errorf("%w: this run is not on GitHub Actions. Use policy.permit.signing: hmac on other CIs", ErrSigstoreNotGitHub)
	}
	if sigstoreGetenv(OIDCRequestURLEnv) == "" || sigstoreGetenv(OIDCRequestTokenEnv) == "" {
		return nil, fmt.Errorf("%w: give the job that runs gate the permission id-token: write", ErrNoOIDCToken)
	}
	path, err := findCosign()
	if err != nil {
		return nil, err
	}
	repo, _, err := PinnedRepository(cfg, sigstoreGetenv)
	if err != nil {
		return nil, err
	}
	return &sigstoreSigner{cosign: path, identity: cfg.Identity, issuer: cfg.Issuer, repository: repo, canSign: true}, nil
}

// newSigstoreVerifier returns the verifier that checks a sigstore signature against cfg's
// identity, issuer and source repository. Verifying needs no OIDC token, so it runs on GitHub
// Actions and on a machine with no CI, where policy.permit.repository must be set. It refuses
// Bitbucket Pipelines and GitLab CI, where sigstore is not supported, a missing cosign and a
// missing repository, in that order.
func newSigstoreVerifier(cfg config.PermitSigning) (Signer, error) {
	if err := sigstoreIdentity(cfg); err != nil {
		return nil, err
	}
	if name := OtherCI(sigstoreGetenv); name != "" {
		return nil, fmt.Errorf("%w: this run is on %s. Use policy.permit.signing: hmac there", ErrSigstoreNotGitHub, name)
	}
	path, err := findCosign()
	if err != nil {
		return nil, err
	}
	repo, _, err := PinnedRepository(cfg, sigstoreGetenv)
	if err != nil {
		return nil, err
	}
	return &sigstoreSigner{cosign: path, identity: cfg.Identity, issuer: cfg.Issuer, repository: repo}, nil
}

// sigstoreIdentity refuses a config without the identity and issuer a signature must carry.
// Config checking already requires both, so this guards callers that skipped it.
func sigstoreIdentity(cfg config.PermitSigning) error {
	if strings.TrimSpace(cfg.Identity) == "" || strings.TrimSpace(cfg.Issuer) == "" {
		return errors.New("permit: sigstore signing needs policy.permit.identity and policy.permit.issuer")
	}
	return nil
}

// OtherCI names the CI that getenv describes when it is one GlassOps knows other than GitHub
// Actions, where sigstore is not supported. It returns "" for GitHub Actions and for a run
// outside any known CI.
func OtherCI(getenv func(string) string) string {
	switch {
	case getenv("GITHUB_ACTIONS") == "true":
		return ""
	case getenv("BITBUCKET_BUILD_NUMBER") != "":
		return "Bitbucket Pipelines"
	case getenv("GITLAB_CI") == "true":
		return "GitLab CI"
	}
	return ""
}

func findCosign() (string, error) {
	path, err := cosignLookPath(Cosign)
	if err != nil {
		return "", fmt.Errorf("%w: %s is not on PATH. Install it from https://docs.sigstore.dev/cosign/system_config/installation/", ErrNoCosign, Cosign)
	}
	return path, nil
}

func (s *sigstoreSigner) Alg() string { return SigstoreAlgorithm }

// Sign signs canonical keyless and returns the cosign bundle in standard base64. Signing
// records the PERMIT's hash and the signing certificate in the public Rekor log, never the
// PERMIT itself.
func (s *sigstoreSigner) Sign(canonical []byte) (string, error) {
	if !s.canSign {
		return "", errors.New("a sigstore verifier does not sign")
	}
	dir, err := os.MkdirTemp("", "glassops-sigstore-")
	if err != nil {
		return "", fmt.Errorf("create a temp dir: %w", err)
	}
	defer os.RemoveAll(dir)
	blob, bundle := filepath.Join(dir, "permit.json"), filepath.Join(dir, "permit.sigstore.json")
	if err := os.WriteFile(blob, canonical, 0o600); err != nil {
		return "", fmt.Errorf("write the PERMIT for cosign: %w", err)
	}
	// --yes answers cosign's upload prompt, which a CI job cannot answer.
	if err := s.run("sign-blob", "--yes", "--bundle", bundle, blob); err != nil {
		return "", err
	}
	data, err := os.ReadFile(bundle)
	if err != nil {
		return "", errors.New("cosign sign-blob wrote no bundle")
	}
	if err := checkBundle(data); err != nil {
		return "", fmt.Errorf("cosign sign-blob wrote a bundle GlassOps cannot read: %w", err)
	}
	return base64.StdEncoding.EncodeToString(data), nil
}

// Verify checks value, a cosign bundle in standard base64, over canonical. It passes only when
// cosign verifies the signature, the transparency log entry and a certificate for the
// configured identity from the configured issuer, issued to a workflow run in the pinned
// repository.
func (s *sigstoreSigner) Verify(canonical []byte, value string) error {
	data, err := base64.StdEncoding.DecodeString(value)
	if err != nil {
		return errors.New("the signature is not valid base64")
	}
	if err := checkBundle(data); err != nil {
		return fmt.Errorf("the signature is not a sigstore bundle: %w", err)
	}
	dir, err := os.MkdirTemp("", "glassops-sigstore-")
	if err != nil {
		return fmt.Errorf("create a temp dir: %w", err)
	}
	defer os.RemoveAll(dir)
	blob, bundle := filepath.Join(dir, "permit.json"), filepath.Join(dir, "permit.sigstore.json")
	if err := os.WriteFile(blob, canonical, 0o600); err != nil {
		return fmt.Errorf("write the PERMIT for cosign: %w", err)
	}
	if err := os.WriteFile(bundle, data, 0o600); err != nil {
		return fmt.Errorf("write the bundle for cosign: %w", err)
	}
	return s.run("verify-blob", "--bundle", bundle,
		"--certificate-identity", s.identity,
		"--certificate-oidc-issuer", s.issuer,
		"--certificate-github-workflow-repository", s.repository,
		blob)
}

// bundleMediaTypePrefix starts the mediaType of every Sigstore protobuf bundle in JSON, such
// as application/vnd.dev.sigstore.bundle.v0.3+json.
const bundleMediaTypePrefix = "application/vnd.dev.sigstore.bundle"

// checkBundle accepts the two bundle shapes cosign writes and refuses anything else before
// cosign sees it. cosign v2 "sign-blob --bundle" writes the legacy shape, a JSON object with
// base64Signature, cert and rekorBundle and no mediaType. The newer protobuf bundle, written
// by --new-bundle-format and by cosign v3, carries a mediaType that names it. cosign itself
// still checks every value in the bundle.
func checkBundle(data []byte) error {
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(data, &fields); err != nil || fields == nil {
		return errors.New("it is not a JSON object")
	}
	str := func(name string) (string, bool) {
		var s string
		raw, ok := fields[name]
		if !ok || json.Unmarshal(raw, &s) != nil {
			return "", false
		}
		return s, true
	}
	if mt, ok := fields["mediaType"]; ok {
		if s, _ := str("mediaType"); !strings.HasPrefix(s, bundleMediaTypePrefix) {
			return fmt.Errorf("mediaType %s is not a Sigstore bundle", strings.TrimSpace(string(mt)))
		}
		return nil
	}
	sig, ok := str("base64Signature")
	if !ok || sig == "" {
		return errors.New("it has neither a mediaType nor a base64Signature")
	}
	if _, err := base64.StdEncoding.DecodeString(sig); err != nil {
		return errors.New("its base64Signature is not base64")
	}
	if cert, ok := str("cert"); !ok || cert == "" {
		return errors.New("it has no signing certificate (cert)")
	}
	var rekor map[string]json.RawMessage
	if err := json.Unmarshal(fields["rekorBundle"], &rekor); err != nil || len(rekor) == 0 {
		return errors.New("it has no transparency log entry (rekorBundle)")
	}
	return nil
}

// run runs cosign with args and the process environment, with no stdin. A failure names the
// subcommand and cosign's last line of error output, with email-shaped values masked.
func (s *sigstoreSigner) run(args ...string) error {
	ctx, cancel := context.WithTimeout(context.Background(), cosignTimeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, s.cosign, args...)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err == nil {
		return nil
	}
	if ctx.Err() != nil {
		return fmt.Errorf("cosign %s did not finish within %s", args[0], cosignTimeout)
	}
	if last := lastLine(stderr.String()); last != "" {
		return fmt.Errorf("cosign %s failed: %s", args[0], redact.Text(last))
	}
	return fmt.Errorf("cosign %s failed: %v", args[0], err)
}

// lastLine returns the last non-blank line of s, cut to 300 bytes.
func lastLine(s string) string {
	lines := strings.Split(strings.TrimSpace(s), "\n")
	last := strings.TrimSpace(lines[len(lines)-1])
	if len(last) > 300 {
		last = last[:300]
	}
	return last
}
