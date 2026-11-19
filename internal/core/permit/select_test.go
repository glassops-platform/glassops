package permit

import (
	"errors"
	"strings"
	"testing"

	"github.com/glassops-platform/glassops/internal/core/config"
)

func TestMethod(t *testing.T) {
	if Method(nil) != config.SigningHMAC {
		t.Fatal("nil config is not hmac")
	}
	if Method(&config.Config{}) != config.SigningHMAC {
		t.Fatal("empty config is not hmac")
	}
	sig := &config.Config{Policy: config.Policy{Permit: config.PermitSigning{Signing: "sigstore"}}}
	if Method(sig) != config.SigningSigstore {
		t.Fatal("sigstore not selected")
	}
}

func TestSignerAndVerifierForHMAC(t *testing.T) {
	key := []byte("select-test-key-0123456789abcdef")
	calls := 0
	k := Keys{HMACKey: func() ([]byte, error) { calls++; return key, nil }}
	s, err := SignerFor(nil, k)
	if err != nil || s.Alg() != Algorithm {
		t.Fatalf("SignerFor: %v, %v", s, err)
	}
	v, err := VerifierFor(&config.Config{}, k)
	if err != nil {
		t.Fatal(err)
	}
	sig, err := s.Sign([]byte("canonical"))
	if err != nil {
		t.Fatal(err)
	}
	if err := v.Verify([]byte("canonical"), sig); err != nil {
		t.Fatalf("round trip: %v", err)
	}
	if calls != 2 {
		t.Fatalf("key read %d times", calls)
	}

	built := 0
	k.NewHMAC = func(b []byte) Signer { built++; return NewHMAC(b) }
	if _, err := SignerFor(nil, k); err != nil || built != 1 {
		t.Fatalf("NewHMAC not used: %v, %d", err, built)
	}

	want := errors.New("GLASSOPS_PERMIT_KEY is not set")
	if _, err := SignerFor(nil, Keys{HMACKey: func() ([]byte, error) { return nil, want }}); err != want {
		t.Fatalf("key error not passed through: %v", err)
	}
	if _, err := VerifierFor(nil, Keys{}); err == nil {
		t.Fatal("no key source accepted")
	}
}

func TestSigstoreIsSelected(t *testing.T) {
	cfg := &config.Config{Policy: config.Policy{Permit: config.PermitSigning{Signing: "sigstore", Identity: "i", Issuer: "o"}}}
	readKey := Keys{HMACKey: func() ([]byte, error) { t.Fatal("sigstore read the HMAC key"); return nil, nil }}
	t.Setenv("GITHUB_ACTIONS", "")
	if _, err := SignerFor(cfg, readKey); !errors.Is(err, ErrSigstoreNotGitHub) {
		t.Fatalf("SignerFor: %v", err)
	}
	t.Setenv("GITLAB_CI", "true")
	if _, err := VerifierFor(cfg, readKey); !errors.Is(err, ErrSigstoreNotGitHub) {
		t.Fatalf("VerifierFor: %v", err)
	}
	bad := &config.Config{Policy: config.Policy{Permit: config.PermitSigning{Signing: "gpg"}}}
	if _, err := SignerFor(bad, readKey); err == nil || !strings.Contains(err.Error(), "unknown signing method") {
		t.Fatalf("unknown method: %v", err)
	}
}
