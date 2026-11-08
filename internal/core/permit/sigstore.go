package permit

import (
	"errors"

	"github.com/glassops-platform/glassops/internal/core/config"
)

// errSigstoreNotImplemented is what sigstore signing returns until it lands.
var errSigstoreNotImplemented = errors.New("permit: sigstore signing is not implemented yet")

// newSigstoreSigner returns the keyless signer for cfg. It is not implemented yet.
func newSigstoreSigner(cfg config.PermitSigning) (Signer, error) {
	return nil, errSigstoreNotImplemented
}

// newSigstoreVerifier returns the verifier that checks a sigstore signature against cfg's
// identity and issuer. It is not implemented yet.
func newSigstoreVerifier(cfg config.PermitSigning) (Signer, error) {
	return nil, errSigstoreNotImplemented
}
