// Package permit signs and verifies PERMITs, the record that a deploy of one plan to one
// target was allowed.
package permit

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/glassops-platform/glassops/internal/core/config"
	"github.com/glassops-platform/glassops/internal/core/policy"
	"github.com/glassops-platform/glassops/internal/providers"
)

// Schema identifies the PERMIT file.
const Schema = "glassops.permit/v1"

// KeyEnv is the variable the HMAC key is read from.
const KeyEnv = "GLASSOPS_PERMIT_KEY"

// MinKeyBytes is the shortest HMAC key accepted, after base64 decoding.
const MinKeyBytes = 32

// MinKeyDistinctBytes is the fewest different byte values an HMAC key must hold, so a
// placeholder such as 32 zero bytes or one character repeated never signs.
const MinKeyDistinctBytes = 8

// MaxClockSkew is how far in the future a PERMIT's issued_at may be and still verify. It
// allows for small clock differences between the gate and the deploy machines.
const MaxClockSkew = 5 * time.Minute

// Algorithm is the signature algorithm NewHMAC signs with.
const Algorithm = "HMAC-SHA256"

// Errors Verify returns. Callers tell them apart with errors.Is.
var (
	// ErrDenied means the PERMIT records a refusal, so it never allows a deploy.
	ErrDenied = errors.New("permit: the permit denies the deploy")
	// ErrExpired means the PERMIT is past its expiry time.
	ErrExpired = errors.New("permit: the permit has expired")
	// ErrMismatch means the PERMIT is for a different target, commit, plan or config.
	ErrMismatch = errors.New("permit: the permit does not match this deploy")
	// ErrSignature means the signature is missing, uses another algorithm or does not verify.
	ErrSignature = errors.New("permit: the permit signature is not valid")
	// ErrUnreadable means a file exists but is not a PERMIT this version can read.
	ErrUnreadable = errors.New("permit: the file is not a readable permit")
	// ErrLifetime means the PERMIT's issued_at is missing or in the future, or it lives longer
	// than the maximum PERMIT lifetime.
	ErrLifetime = errors.New("permit: the permit's lifetime is not valid")
)

// Permit is the signed record that a deploy of one plan to one target was allowed.
type Permit struct {
	Schema string `json:"schema"`
	// ID is random, in hex.
	ID        string    `json:"id"`
	Allow     bool      `json:"allow"`
	IssuedAt  time.Time `json:"issued_at"`
	ExpiresAt time.Time `json:"expires_at"`
	Target    Target    `json:"target"`
	Commit    string    `json:"commit"`
	// PlanSHA256 and ConfigSHA256 bind the PERMIT to one change set and one config.
	PlanSHA256   string `json:"plan_sha256"`
	ConfigSHA256 string `json:"config_sha256"`
	// ChangeSetSHA256 binds the rest of the change set, which plan_sha256 does not cover: the
	// provider, the plan dir, which file is the manifest and which are destructive, and the
	// empty and full flags. See ChangeSetDigest.
	ChangeSetSHA256 string `json:"change_set_sha256"`
	// DeployCommandSHA256 is set when gate was given the deploy command for provider exec. It
	// is DeployCommandDigest of that command, never the command itself.
	DeployCommandSHA256 string          `json:"deploy_command_sha256,omitempty"`
	Validation          *Validation     `json:"validation,omitempty"`
	SeparationOfDuties  *SoD            `json:"separation_of_duties,omitempty"`
	Checks              []policy.Result `json:"checks"`
	Signature           Signature       `json:"signature"`
}

// Target is the target a PERMIT is for.
type Target struct {
	Name string `json:"name"`
	Tier string `json:"tier"`
	ID   string `json:"id"`
	Kind string `json:"kind"`
}

// Validation is the check-only validation a PERMIT is bound to.
type Validation struct {
	DeployID  string `json:"deploy_id"`
	TestLevel string `json:"test_level,omitempty"`
	Passed    bool   `json:"passed"`
}

// SoD records how separation of duties was met.
type SoD struct {
	// Mode is "enforced", "waived_by_config", "satisfied_by_approval" or "not_applicable".
	Mode string `json:"mode"`
	// Mechanism is "environment" or "label".
	Mechanism string `json:"mechanism,omitempty"`
	Approver  string `json:"approver,omitempty"`
	Author    string `json:"author,omitempty"`
	Deployer  string `json:"deployer,omitempty"`
}

// Signature is the signature over a PERMIT's canonical bytes.
type Signature struct {
	// Alg is the signing algorithm, such as "HMAC-SHA256".
	Alg   string `json:"alg"`
	KeyID string `json:"key_id,omitempty"`
	// Value is base64.
	Value string `json:"value"`
}

// Signer signs and checks the canonical bytes of a PERMIT.
type Signer interface {
	Alg() string
	Sign(canonical []byte) (string, error)
	Verify(canonical []byte, value string) error
}

// KeyIdentifier is implemented by signers that can name their key without exposing it. Sign
// records the name in the PERMIT, and Verify refuses a PERMIT signed with another key.
type KeyIdentifier interface {
	KeyID() string
}

// Expect is what a PERMIT must match to be honoured.
type Expect struct {
	TargetName, TargetID, Commit, PlanSHA256, ConfigSHA256 string
	// ChangeSetSHA256 is ChangeSetDigest of the change set being deployed.
	ChangeSetSHA256 string
	// DeployCommandSHA256, when set, is DeployCommandDigest of the command deploy will run in
	// place of exec.deploy. The PERMIT must carry the same value. When empty the PERMIT's
	// command hash is not checked, because the command comes from the signed config.
	DeployCommandSHA256 string
	Now                 time.Time
	// MaxTTL is the longest lifetime, expires_at minus issued_at, a PERMIT may have. Zero
	// means config.MaxPermitTTL.
	MaxTTL time.Duration
	// CanonicalID, when set, is applied to the PERMIT's target Id before it is compared with
	// TargetID. The caller passes its provider's canonical form for TargetID and sets this to
	// the same function, so a 15 character and an 18 character Id compare equal. When nil the
	// two Ids are compared as written.
	CanonicalID func(string) string
}

// Canonical returns the bytes a PERMIT's signature covers: the PERMIT with signature.value
// emptied, as JSON with keys sorted at every level, no insignificant whitespace, and times in
// UTC. The same PERMIT gives the same bytes however its fields or maps were built.
func Canonical(p Permit) ([]byte, error) {
	p.Signature.Value = ""
	p.IssuedAt = p.IssuedAt.UTC()
	p.ExpiresAt = p.ExpiresAt.UTC()
	if p.Checks == nil {
		p.Checks = []policy.Result{}
	}
	raw, err := json.Marshal(p)
	if err != nil {
		return nil, fmt.Errorf("permit: encode: %w", err)
	}
	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.UseNumber()
	var generic any
	if err := dec.Decode(&generic); err != nil {
		return nil, fmt.Errorf("permit: decode: %w", err)
	}
	var out bytes.Buffer
	enc := json.NewEncoder(&out)
	enc.SetEscapeHTML(false)
	// encoding/json writes map keys sorted, so the generic form is canonical.
	if err := enc.Encode(generic); err != nil {
		return nil, fmt.Errorf("permit: encode canonical: %w", err)
	}
	return bytes.TrimRight(out.Bytes(), "\n"), nil
}

// Sign sets p's algorithm, key id (when the signer has one) and signature value from s.
func Sign(p *Permit, s Signer) error {
	p.Signature.Alg = s.Alg()
	p.Signature.KeyID = ""
	if k, ok := s.(KeyIdentifier); ok {
		p.Signature.KeyID = k.KeyID()
	}
	p.Signature.Value = ""
	canon, err := Canonical(*p)
	if err != nil {
		return err
	}
	v, err := s.Sign(canon)
	if err != nil {
		return fmt.Errorf("permit: sign: %w", err)
	}
	p.Signature.Value = v
	return nil
}

// Verify checks p's signature with s and that p allows a deploy matching want. It checks, in
// order, that p is signed and binds a change set, the signature, allow, expiry and the
// lifetime, then the target, commit, plan, config and change set, and the deploy command when
// want names one. The error wraps the matching sentinel and says which check failed, never
// any key material.
func Verify(p Permit, s Signer, want Expect) error {
	if p.Signature.Value == "" {
		return fmt.Errorf("%w: the permit is not signed", ErrSignature)
	}
	if p.ChangeSetSHA256 == "" {
		// Checked before the signature so a PERMIT from an older gate gets this message rather
		// than a signature error. A missing field can only refuse.
		return fmt.Errorf("%w: the permit has no change_set_sha256, so an older gate wrote it and it does not bind the change set. Run gate again", ErrMismatch)
	}
	if p.Signature.Alg != s.Alg() {
		return fmt.Errorf("%w: algorithm is %q, want %q", ErrSignature, p.Signature.Alg, s.Alg())
	}
	if k, ok := s.(KeyIdentifier); ok && p.Signature.KeyID != k.KeyID() {
		return fmt.Errorf("%w: the permit was signed with key %q, this key is %q", ErrSignature, p.Signature.KeyID, k.KeyID())
	}
	canon, err := Canonical(p)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrSignature, err)
	}
	if err := s.Verify(canon, p.Signature.Value); err != nil {
		return fmt.Errorf("%w: %v", ErrSignature, err)
	}
	if p.Schema != Schema {
		return fmt.Errorf("%w: schema is %q, want %q", ErrMismatch, p.Schema, Schema)
	}
	if !p.Allow {
		return fmt.Errorf("%w: allow is false", ErrDenied)
	}
	if want.Now.IsZero() {
		return fmt.Errorf("%w: no clock to check expiry against", ErrExpired)
	}
	if !want.Now.Before(p.ExpiresAt) {
		return fmt.Errorf("%w: expired at %s", ErrExpired, p.ExpiresAt.UTC().Format(time.RFC3339))
	}
	if err := checkLifetime(p, want); err != nil {
		return err
	}
	if p.Target.Name != want.TargetName {
		return fmt.Errorf("%w: the permit is for target %q, not %q", ErrMismatch, p.Target.Name, want.TargetName)
	}
	gotID := p.Target.ID
	if want.CanonicalID != nil {
		gotID = want.CanonicalID(gotID)
	}
	if want.TargetID == "" || gotID != want.TargetID {
		return fmt.Errorf("%w: the permit is for a different target Id", ErrMismatch)
	}
	for _, c := range []struct{ name, got, want string }{
		{"commit", p.Commit, want.Commit},
		{"plan hash", p.PlanSHA256, want.PlanSHA256},
		{"config hash", p.ConfigSHA256, want.ConfigSHA256},
		{"change set hash", p.ChangeSetSHA256, want.ChangeSetSHA256},
	} {
		if c.want == "" || c.got != c.want {
			return fmt.Errorf("%w: the %s differs", ErrMismatch, c.name)
		}
	}
	if want.DeployCommandSHA256 != "" {
		switch {
		case p.DeployCommandSHA256 == "":
			return fmt.Errorf("%w: the permit binds no deploy command. Run gate with --deploy-command to allow one", ErrMismatch)
		case p.DeployCommandSHA256 != want.DeployCommandSHA256:
			return fmt.Errorf("%w: the deploy command differs from the one the permit was issued for", ErrMismatch)
		}
	}
	return nil
}

// checkLifetime refuses a PERMIT with no issued_at, one issued more than MaxClockSkew after
// want.Now, one that expires before it was issued, and one that lives longer than want.MaxTTL.
// gate never writes such a PERMIT, so each means a broken clock or another signer.
func checkLifetime(p Permit, want Expect) error {
	limit := want.MaxTTL
	if limit <= 0 {
		limit = config.MaxPermitTTL
	}
	life := p.ExpiresAt.Sub(p.IssuedAt)
	switch {
	case p.IssuedAt.IsZero():
		return fmt.Errorf("%w: issued_at is not set", ErrLifetime)
	case p.IssuedAt.After(want.Now.Add(MaxClockSkew)):
		return fmt.Errorf("%w: issued at %s, more than %s after this clock (%s)", ErrLifetime,
			p.IssuedAt.UTC().Format(time.RFC3339), MaxClockSkew, want.Now.UTC().Format(time.RFC3339))
	case life <= 0:
		return fmt.Errorf("%w: it expires before it was issued", ErrLifetime)
	case life > limit:
		return fmt.Errorf("%w: it lives %s, longer than the maximum %s", ErrLifetime, life, limit)
	}
	return nil
}

// ChangeSetDigest is the change_set_sha256 a PERMIT binds a change set by. It covers
// plan_sha256 and the change set fields that say how the plan files are used: the provider,
// the plan dir, the manifest, the destructive list in order, and the empty and full flags. It
// is the sha256, in hex, of plan_sha256, a newline, and the canonical JSON of those fields.
// The facts are not covered: gate reads them once, and the PERMIT records what it decided.
func ChangeSetDigest(cs providers.ChangeSet) string {
	destructive := cs.Destructive
	if destructive == nil {
		destructive = []string{}
	}
	fields := map[string]any{
		"provider":    cs.Provider,
		"dir":         cs.Dir,
		"manifest":    cs.Manifest,
		"destructive": destructive,
		"empty":       cs.Empty,
		"full":        cs.Full,
	}
	h := sha256.New()
	h.Write([]byte(cs.SHA256))
	h.Write([]byte{'\n'})
	h.Write(canonicalJSON(fields))
	return hex.EncodeToString(h.Sum(nil))
}

// DeployCommandDigest is the deploy_command_sha256 a PERMIT binds a provider exec deploy
// command by: the sha256, in hex, of the argv as a canonical JSON array of strings. Each
// argument stays one element, so ["a b"] and ["a", "b"] differ. An empty argv gives "".
func DeployCommandDigest(argv []string) string {
	if len(argv) == 0 {
		return ""
	}
	sum := sha256.Sum256(canonicalJSON(argv))
	return hex.EncodeToString(sum[:])
}

// canonicalJSON encodes v with sorted map keys, no HTML escaping and no trailing newline. It
// is only given maps and slices of strings and bools, which always encode.
func canonicalJSON(v any) []byte {
	var out bytes.Buffer
	enc := json.NewEncoder(&out)
	enc.SetEscapeHTML(false)
	_ = enc.Encode(v)
	return bytes.TrimRight(out.Bytes(), "\n")
}

// ParseKey decodes a base64 key such as the one in GLASSOPS_PERMIT_KEY. It accepts standard
// base64, with or without padding, and rejects a key shorter than MinKeyBytes or with fewer
// than MinKeyDistinctBytes different byte values. The error never contains the key.
func ParseKey(b64 string) ([]byte, error) {
	b64 = strings.TrimSpace(b64)
	key, err := base64.StdEncoding.DecodeString(b64)
	if err != nil {
		var rawErr error
		if key, rawErr = base64.RawStdEncoding.DecodeString(b64); rawErr != nil {
			return nil, errors.New("permit: the key is not valid base64")
		}
	}
	if err := checkKey(key); err != nil {
		return nil, fmt.Errorf("permit: %w", err)
	}
	return key, nil
}

// checkKey refuses a key shorter than MinKeyBytes or with fewer than MinKeyDistinctBytes
// different byte values. The error never contains the key.
func checkKey(key []byte) error {
	if len(key) < MinKeyBytes {
		return fmt.Errorf("the key is %d bytes, want at least %d", len(key), MinKeyBytes)
	}
	var seen [256]bool
	distinct := 0
	for _, b := range key {
		if !seen[b] {
			seen[b] = true
			distinct++
		}
	}
	if distinct < MinKeyDistinctBytes {
		return fmt.Errorf("the key has %d different byte values, want at least %d. Generate a random one, such as with openssl rand -base64 32", distinct, MinKeyDistinctBytes)
	}
	return nil
}

// hmacSigner signs with HMAC-SHA256.
type hmacSigner struct {
	key   []byte
	keyID string
}

// NewHMAC returns a signer that uses HMAC-SHA256 with key. A key ParseKey refuses, too short
// or with too few different byte values, gives a signer whose Sign and Verify always return
// an error, so a weak key never signs.
func NewHMAC(key []byte) Signer {
	sum := sha256.Sum256(key)
	return &hmacSigner{key: append([]byte(nil), key...), keyID: hex.EncodeToString(sum[:])[:8]}
}

func (h *hmacSigner) Alg() string   { return Algorithm }
func (h *hmacSigner) KeyID() string { return h.keyID }

func (h *hmacSigner) mac(canonical []byte) ([]byte, error) {
	if err := checkKey(h.key); err != nil {
		return nil, err
	}
	m := hmac.New(sha256.New, h.key)
	m.Write(canonical)
	return m.Sum(nil), nil
}

func (h *hmacSigner) Sign(canonical []byte) (string, error) {
	sum, err := h.mac(canonical)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(sum), nil
}

func (h *hmacSigner) Verify(canonical []byte, value string) error {
	// Strict decoding, and the value must be exactly what Sign writes, so a signature has one
	// spelling: no other padding bits, no line breaks, no missing padding.
	got, err := base64.StdEncoding.Strict().DecodeString(value)
	if err != nil || base64.StdEncoding.EncodeToString(got) != value {
		return errors.New("the signature is not valid base64")
	}
	want, err := h.mac(canonical)
	if err != nil {
		return err
	}
	if !hmac.Equal(got, want) {
		return errors.New("the signature does not match")
	}
	return nil
}

// Write writes p to path as indented JSON, creating the directory.
func Write(path string, p Permit) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("permit: create dir: %w", err)
	}
	data, err := json.MarshalIndent(p, "", "  ")
	if err != nil {
		return fmt.Errorf("permit: encode: %w", err)
	}
	if err := os.WriteFile(path, append(data, '\n'), 0o644); err != nil {
		return fmt.Errorf("permit: write: %w", err)
	}
	return nil
}

// Read reads a PERMIT written by Write. A missing file returns an error that wraps the os
// error. A file with unknown fields or trailing data wraps ErrUnreadable, because fields
// outside the type are not covered by the signature.
func Read(path string) (Permit, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Permit{}, fmt.Errorf("permit: read: %w", err)
	}
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.DisallowUnknownFields()
	var p Permit
	if err := dec.Decode(&p); err != nil {
		return Permit{}, fmt.Errorf("%w: %s: %v", ErrUnreadable, path, err)
	}
	if dec.More() {
		return Permit{}, fmt.Errorf("%w: %s: trailing data", ErrUnreadable, path)
	}
	return p, nil
}
