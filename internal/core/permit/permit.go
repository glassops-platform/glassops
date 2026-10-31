// Package permit signs and verifies PERMITs, the record that a deploy of one plan to one
// target was allowed.
package permit

import (
	"errors"
	"time"

	"github.com/glassops-platform/glassops/internal/core/policy"
)

// Schema identifies the PERMIT file.
const Schema = "glassops.permit/v1"

// KeyEnv is the variable the HMAC key is read from.
const KeyEnv = "GLASSOPS_PERMIT_KEY"

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
)

// errNotImplemented is what the stubs below return until signing lands.
var errNotImplemented = errors.New("permit: not implemented yet")

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
	PlanSHA256         string          `json:"plan_sha256"`
	ConfigSHA256       string          `json:"config_sha256"`
	Validation         *Validation     `json:"validation,omitempty"`
	SeparationOfDuties *SoD            `json:"separation_of_duties,omitempty"`
	Checks             []policy.Result `json:"checks"`
	Signature          Signature       `json:"signature"`
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

// Expect is what a PERMIT must match to be honoured.
type Expect struct {
	TargetName, TargetID, Commit, PlanSHA256, ConfigSHA256 string
	Now                                                    time.Time
}

// Canonical returns the bytes a PERMIT's signature covers.
func Canonical(p Permit) ([]byte, error) {
	return nil, errNotImplemented
}

// Sign sets p's signature from s.
func Sign(p *Permit, s Signer) error {
	return errNotImplemented
}

// Verify checks p's signature with s and that p allows a deploy matching want.
func Verify(p Permit, s Signer, want Expect) error {
	return errNotImplemented
}

// NewHMAC returns a signer that uses HMAC-SHA256 with key.
func NewHMAC(key []byte) Signer {
	return stubSigner{}
}

// Write writes p to path.
func Write(path string, p Permit) error {
	return errNotImplemented
}

// Read reads a PERMIT written by Write.
func Read(path string) (Permit, error) {
	return Permit{}, errNotImplemented
}

// stubSigner stands in for the HMAC signer until signing lands.
type stubSigner struct{}

func (stubSigner) Alg() string                 { return "HMAC-SHA256" }
func (stubSigner) Sign([]byte) (string, error) { return "", errNotImplemented }
func (stubSigner) Verify([]byte, string) error { return errNotImplemented }
