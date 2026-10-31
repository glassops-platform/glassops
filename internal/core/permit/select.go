package permit

import (
	"errors"
	"fmt"

	"github.com/glassops-platform/glassops/internal/core/config"
)

// Keys is how SignerFor and VerifierFor reach key material. Each command keeps its own
// message for a missing or bad key, so it supplies the key reader.
type Keys struct {
	// HMACKey returns the HMAC key. It is called only when the config selects hmac, and its
	// error is returned unchanged.
	HMACKey func() ([]byte, error)
	// NewHMAC builds the HMAC signer from the key. Nil means NewHMAC. Tests replace it.
	NewHMAC func(key []byte) Signer
}

// Method returns the signing method cfg selects: config.SigningHMAC or
// config.SigningSigstore. A nil cfg means hmac.
func Method(cfg *config.Config) string {
	if cfg == nil {
		return config.SigningHMAC
	}
	return cfg.Policy.Permit.Method()
}

// SignerFor returns the signer gate signs PERMITs with, chosen by policy.permit.signing.
func SignerFor(cfg *config.Config, k Keys) (Signer, error) {
	switch m := Method(cfg); m {
	case config.SigningHMAC:
		return k.hmac()
	case config.SigningSigstore:
		return newSigstoreSigner(cfg.Policy.Permit)
	default:
		return nil, fmt.Errorf("permit: unknown signing method %q", m)
	}
}

// VerifierFor returns what deploy and permit verify check a PERMIT's signature with, chosen
// by policy.permit.signing. It is a Signer so it passes to Verify. A verifier for a keyless
// method may refuse to Sign.
func VerifierFor(cfg *config.Config, k Keys) (Signer, error) {
	switch m := Method(cfg); m {
	case config.SigningHMAC:
		return k.hmac()
	case config.SigningSigstore:
		return newSigstoreVerifier(cfg.Policy.Permit)
	default:
		return nil, fmt.Errorf("permit: unknown signing method %q", m)
	}
}

func (k Keys) hmac() (Signer, error) {
	if k.HMACKey == nil {
		return nil, errors.New("permit: no HMAC key source")
	}
	key, err := k.HMACKey()
	if err != nil {
		return nil, err
	}
	if k.NewHMAC != nil {
		return k.NewHMAC(key), nil
	}
	return NewHMAC(key), nil
}
