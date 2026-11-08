//go:build redteam_open

package permit

// Red team attacks on the permit package that currently succeed. Each test asserts the
// behaviour Verify or ParseKey should have, so it fails until the finding is fixed.
// Reproduce with:
//
//	go test -tags redteam_open -run TestRTOpen ./internal/core/permit/

import (
	"bytes"
	"encoding/base64"
	"strings"
	"testing"
	"time"

	"github.com/glassops-platform/glassops/internal/core/config"
)

// Finding: Verify never compares issued_at with the clock, so a PERMIT issued in the future,
// by a gate with a skewed clock, verifies now and for its whole lifetime plus the skew.
func TestRTOpenPermitIssuedInTheFuture(t *testing.T) {
	p := rtPermit()
	p.IssuedAt = rtNow.Add(6 * time.Hour)
	p.ExpiresAt = p.IssuedAt.Add(2 * time.Hour)
	if err := Sign(&p, NewHMAC(rtKeyA)); err != nil {
		t.Fatal(err)
	}
	if err := Verify(p, NewHMAC(rtKeyA), rtExpect()); err == nil {
		t.Fatal("a PERMIT issued six hours from now verified")
	}
}

// Finding: Verify does not bound expires_at minus issued_at. gate keeps it within
// config.MaxPermitTTL, but nothing on the verify side does, so a PERMIT signed by anything
// else holding the key, or a gate with a broken TTL, can live for years.
func TestRTOpenPermitLifetimeIsUnbounded(t *testing.T) {
	p := rtPermit()
	p.ExpiresAt = p.IssuedAt.AddDate(5, 0, 0)
	if err := Sign(&p, NewHMAC(rtKeyA)); err != nil {
		t.Fatal(err)
	}
	if p.ExpiresAt.Sub(p.IssuedAt) <= config.MaxPermitTTL {
		t.Fatal("setup: the lifetime is not over the maximum")
	}
	if err := Verify(p, NewHMAC(rtKeyA), rtExpect()); err == nil {
		t.Fatal("a PERMIT valid for five years verified")
	}
}

// Finding: ParseKey checks the length only. A key of 32 zero bytes, or one byte repeated, is
// accepted, so a placeholder key copied from an example signs PERMITs.
func TestRTOpenLowEntropyKeyAccepted(t *testing.T) {
	for _, key := range [][]byte{make([]byte, 32), bytes.Repeat([]byte("A"), 32), []byte(strings.Repeat("ab", 16))} {
		if _, err := ParseKey(base64.StdEncoding.EncodeToString(key)); err == nil {
			t.Errorf("ParseKey accepted the low entropy key %q", key)
		}
	}
}

// Finding, informational: the signature value is decoded with non-strict base64, so the
// unused low bits of its last character can be changed and the PERMIT still verifies. The
// content stays the same, so nothing the PERMIT allows changes, but two different files are
// the same PERMIT.
func TestRTOpenSignatureEncodingIsMalleable(t *testing.T) {
	p := rtSigned(t, rtKeyA)
	v := p.Signature.Value
	// 32 bytes encode to 43 characters and one pad. The 43rd character carries 2 unused bits.
	const alphabet = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/"
	i := strings.IndexByte(alphabet, v[42])
	alt := alphabet[i^1]
	p.Signature.Value = v[:42] + string(alt) + v[43:]
	if p.Signature.Value == v {
		t.Fatal("setup: the value did not change")
	}
	if err := Verify(p, NewHMAC(rtKeyA), rtExpect()); err == nil {
		t.Fatalf("a signature value written differently (%s, signed %s) verified", p.Signature.Value, v)
	}
}
