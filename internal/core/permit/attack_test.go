package permit

// Red team tests for PERMIT signing and verification. A test here passes when the attack is
// refused. Attacks that currently succeed are in attack_open_test.go, behind the redteam_open
// build tag.

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/glassops-platform/glassops/internal/core/config"
	"github.com/glassops-platform/glassops/internal/core/policy"
)

var rtNow = time.Date(2026, 10, 26, 12, 0, 0, 0, time.UTC)

// rtKeyA and rtKeyB are synthetic 32 byte keys.
var (
	rtKeyA = []byte("rt-synthetic-key-a-0123456789abc")
	rtKeyB = []byte("rt-synthetic-key-b-0123456789abc")
)

func rtCanon(id string) string {
	if len(id) > 15 {
		return id[:15]
	}
	return id
}

// rtPermit is a synthetic allow PERMIT for target qa with every optional block filled in.
func rtPermit() Permit {
	return Permit{
		Schema: Schema, ID: "0123456789abcdef0123456789abcdef", Allow: true,
		IssuedAt:           rtNow.Add(-10 * time.Minute),
		ExpiresAt:          rtNow.Add(110 * time.Minute),
		Target:             Target{Name: "qa", Tier: "qa", ID: "00D000000000002AAA", Kind: "sandbox"},
		Commit:             strings.Repeat("6", 40),
		PlanSHA256:         strings.Repeat("a", 64),
		ConfigSHA256:       strings.Repeat("c", 64),
		Validation:         &Validation{DeployID: "0Af00000000RT01AAA", TestLevel: "RunLocalTests", Passed: true},
		SeparationOfDuties: &SoD{Mode: "enforced", Author: "author", Deployer: "deployer"},
		Checks:             []policy.Result{{Rule: "rt_rule", Outcome: policy.Pass, Detail: "rt"}},
	}
}

func rtExpect() Expect {
	return Expect{
		TargetName: "qa", TargetID: rtCanon("00D000000000002AAA"),
		Commit: strings.Repeat("6", 40), PlanSHA256: strings.Repeat("a", 64), ConfigSHA256: strings.Repeat("c", 64),
		Now: rtNow, CanonicalID: rtCanon,
	}
}

func rtSigned(t *testing.T, key []byte) Permit {
	t.Helper()
	p := rtPermit()
	if err := Sign(&p, NewHMAC(key)); err != nil {
		t.Fatal(err)
	}
	return p
}

func TestRTControlSignedPermitVerifies(t *testing.T) {
	if err := Verify(rtSigned(t, rtKeyA), NewHMAC(rtKeyA), rtExpect()); err != nil {
		t.Fatalf("control: %v", err)
	}
}

// Attack 2: every field of a signed PERMIT is covered by the signature.
func TestRTTamperEveryField(t *testing.T) {
	cases := []struct {
		name string
		edit func(p *Permit)
	}{
		{"schema", func(p *Permit) { p.Schema = "glassops.permit/v0" }},
		{"id", func(p *Permit) { p.ID = "ffffffffffffffffffffffffffffffff" }},
		{"allow on a deny", func(p *Permit) { p.Allow = !p.Allow }},
		{"issued_at", func(p *Permit) { p.IssuedAt = p.IssuedAt.Add(-time.Hour) }},
		{"expires_at", func(p *Permit) { p.ExpiresAt = p.ExpiresAt.Add(24 * time.Hour) }},
		{"expires_at by a nanosecond", func(p *Permit) { p.ExpiresAt = p.ExpiresAt.Add(time.Nanosecond) }},
		{"target.name", func(p *Permit) { p.Target.Name = "prod" }},
		{"target.tier", func(p *Permit) { p.Target.Tier = "prod" }},
		{"target.id", func(p *Permit) { p.Target.ID = "00D000000000001AAA" }},
		{"target.id suffix only", func(p *Permit) { p.Target.ID = "00D000000000002AAB" }},
		{"target.kind", func(p *Permit) { p.Target.Kind = "production" }},
		{"commit", func(p *Permit) { p.Commit = strings.Repeat("7", 40) }},
		{"plan_sha256", func(p *Permit) { p.PlanSHA256 = strings.Repeat("b", 64) }},
		{"config_sha256", func(p *Permit) { p.ConfigSHA256 = strings.Repeat("d", 64) }},
		{"validation.deploy_id", func(p *Permit) { p.Validation.DeployID = "0Af00000000RT09AAA" }},
		{"validation.test_level", func(p *Permit) { p.Validation.TestLevel = "NoTestRun" }},
		{"validation.passed", func(p *Permit) { p.Validation.Passed = false }},
		{"validation removed", func(p *Permit) { p.Validation = nil }},
		{"sod.mode", func(p *Permit) { p.SeparationOfDuties.Mode = "waived_by_config" }},
		{"sod.approver", func(p *Permit) { p.SeparationOfDuties.Approver = "someone" }},
		{"sod removed", func(p *Permit) { p.SeparationOfDuties = nil }},
		{"check outcome", func(p *Permit) { p.Checks[0].Outcome = policy.Fail }},
		{"check added", func(p *Permit) { p.Checks = append(p.Checks, policy.Result{Rule: "x", Outcome: policy.Pass}) }},
		{"checks emptied", func(p *Permit) { p.Checks = nil }},
		{"signature.key_id", func(p *Permit) { p.Signature.KeyID = "deadbeef" }},
		{"signature.alg", func(p *Permit) { p.Signature.Alg = "HMAC-SHA512" }},
		{"signature.value bit flip", func(p *Permit) {
			raw, _ := base64.StdEncoding.DecodeString(p.Signature.Value)
			raw[0] ^= 1
			p.Signature.Value = base64.StdEncoding.EncodeToString(raw)
		}},
		{"signature.value truncated", func(p *Permit) { p.Signature.Value = p.Signature.Value[:20] }},
		{"signature.value not base64", func(p *Permit) { p.Signature.Value = "!!!!" }},
		{"signature.value empty", func(p *Permit) { p.Signature.Value = "" }},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			p := rtSigned(t, rtKeyA)
			tc.edit(&p)
			err := Verify(p, NewHMAC(rtKeyA), rtExpect())
			if err == nil {
				t.Fatal("a tampered PERMIT verified")
			}
			if !errors.Is(err, ErrSignature) {
				t.Fatalf("refused, but not by the signature: %v", err)
			}
		})
	}
}

// Attack 1 and 3: a correctly signed PERMIT for something else is refused.
func TestRTReplayAndSwapExpectations(t *testing.T) {
	cases := []struct {
		name string
		edit func(w *Expect)
	}{
		{"other target name", func(w *Expect) { w.TargetName = "uat" }},
		{"other org Id", func(w *Expect) { w.TargetID = rtCanon("00D000000000004AAA") }},
		{"no target Id expected", func(w *Expect) { w.TargetID = "" }},
		{"other commit", func(w *Expect) { w.Commit = strings.Repeat("7", 40) }},
		{"no commit expected", func(w *Expect) { w.Commit = "" }},
		{"other plan", func(w *Expect) { w.PlanSHA256 = strings.Repeat("b", 64) }},
		{"no plan expected", func(w *Expect) { w.PlanSHA256 = "" }},
		{"other config", func(w *Expect) { w.ConfigSHA256 = strings.Repeat("d", 64) }},
		{"no config expected", func(w *Expect) { w.ConfigSHA256 = "" }},
		{"18 character Id with no canonical form", func(w *Expect) {
			w.CanonicalID = nil
			w.TargetID = "00D000000000002"
		}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			w := rtExpect()
			tc.edit(&w)
			if err := Verify(rtSigned(t, rtKeyA), NewHMAC(rtKeyA), w); !errors.Is(err, ErrMismatch) {
				t.Fatalf("want ErrMismatch, got %v", err)
			}
		})
	}
}

// Attack 4: key confusion.
func TestRTKeyConfusion(t *testing.T) {
	t.Run("another key", func(t *testing.T) {
		if err := Verify(rtSigned(t, rtKeyA), NewHMAC(rtKeyB), rtExpect()); !errors.Is(err, ErrSignature) {
			t.Fatalf("got %v", err)
		}
	})
	t.Run("another key with the key id rewritten to match", func(t *testing.T) {
		p := rtSigned(t, rtKeyA)
		p.Signature.KeyID = NewHMAC(rtKeyB).(KeyIdentifier).KeyID()
		if err := Verify(p, NewHMAC(rtKeyB), rtExpect()); !errors.Is(err, ErrSignature) {
			t.Fatalf("got %v", err)
		}
	})
	for _, key := range [][]byte{{}, rtKeyA[:31], []byte("short")} {
		t.Run(fmt.Sprintf("a %d byte key signs nothing", len(key)), func(t *testing.T) {
			p := rtPermit()
			if err := Sign(&p, NewHMAC(key)); err == nil {
				t.Fatal("a short key signed")
			}
			// A PERMIT forged with the same short key never verifies either.
			m := hmac.New(sha256.New, key)
			canon, _ := Canonical(p)
			m.Write(canon)
			p.Signature = Signature{Alg: Algorithm, KeyID: NewHMAC(key).(KeyIdentifier).KeyID(), Value: base64.StdEncoding.EncodeToString(m.Sum(nil))}
			if err := Verify(p, NewHMAC(key), rtExpect()); !errors.Is(err, ErrSignature) {
				t.Fatalf("got %v", err)
			}
		})
	}
	t.Run("ParseKey refuses empty, short and non base64", func(t *testing.T) {
		for _, in := range []string{"", "   ", "\n", base64.StdEncoding.EncodeToString(rtKeyA[:31]), "not base64!", "%%%%"} {
			if _, err := ParseKey(in); err == nil {
				t.Fatalf("ParseKey(%q) accepted", in)
			}
		}
	})
	t.Run("the key never reaches an error", func(t *testing.T) {
		enc := base64.StdEncoding.EncodeToString(rtKeyA[:31])
		_, err := ParseKey(enc)
		if err == nil || strings.Contains(err.Error(), enc) || strings.Contains(err.Error(), string(rtKeyA[:31])) {
			t.Fatalf("error %v", err)
		}
	})
}

// Attack 4: an HMAC PERMIT where the config asks for sigstore. Only the selection logic is on
// this base. Whatever key material is around, a sigstore config never yields an HMAC verifier.
func TestRTSigstoreConfigNeverSelectsHMAC(t *testing.T) {
	cfg := &config.Config{Policy: config.Policy{Permit: config.PermitSigning{
		Signing: config.SigningSigstore, Identity: "rt-identity", Issuer: "rt-issuer",
	}}}
	called := false
	keys := Keys{HMACKey: func() ([]byte, error) { called = true; return rtKeyA, nil }}
	if v, err := VerifierFor(cfg, keys); err == nil {
		if err := Verify(rtSigned(t, rtKeyA), v, rtExpect()); err == nil {
			t.Fatal("an HMAC PERMIT verified under a sigstore config")
		}
	}
	if s, err := SignerFor(cfg, keys); err == nil && s.Alg() == Algorithm {
		t.Fatal("a sigstore config selected the HMAC signer")
	}
	if called {
		t.Fatal("a sigstore config read the HMAC key")
	}
	for _, m := range []string{"HMAC", "Sigstore", "none", " hmac"} {
		cfg := &config.Config{Policy: config.Policy{Permit: config.PermitSigning{Signing: m}}}
		if _, err := VerifierFor(cfg, keys); err == nil {
			t.Fatalf("signing %q selected a verifier", m)
		}
	}
	// A verifier with another algorithm name refuses the HMAC PERMIT before any key check.
	if err := Verify(rtSigned(t, rtKeyA), rtAlgSigner{alg: "sigstore"}, rtExpect()); !errors.Is(err, ErrSignature) {
		t.Fatalf("got %v", err)
	}
}

// rtAlgSigner accepts every signature, so only the algorithm check can refuse.
type rtAlgSigner struct{ alg string }

func (s rtAlgSigner) Alg() string                 { return s.alg }
func (s rtAlgSigner) Sign([]byte) (string, error) { return "x", nil }
func (s rtAlgSigner) Verify([]byte, string) error { return nil }

// Attack 5: expiry and clock.
func TestRTExpiryAndClock(t *testing.T) {
	p := rtSigned(t, rtKeyA)
	for _, tc := range []struct {
		name string
		now  time.Time
	}{
		{"exactly at expiry", p.ExpiresAt},
		{"a nanosecond after", p.ExpiresAt.Add(time.Nanosecond)},
		{"expiry in another zone, same instant", p.ExpiresAt.In(time.FixedZone("rt+14", 14*3600))},
		{"no clock", time.Time{}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			w := rtExpect()
			w.Now = tc.now
			if err := Verify(p, NewHMAC(rtKeyA), w); !errors.Is(err, ErrExpired) {
				t.Fatalf("want ErrExpired, got %v", err)
			}
		})
	}
	t.Run("a zero expires_at", func(t *testing.T) {
		z := rtPermit()
		z.ExpiresAt = time.Time{}
		if err := Sign(&z, NewHMAC(rtKeyA)); err != nil {
			t.Fatal(err)
		}
		if err := Verify(z, NewHMAC(rtKeyA), rtExpect()); !errors.Is(err, ErrExpired) {
			t.Fatalf("got %v", err)
		}
	})
}

// Attack 2: a PERMIT file that decodes to something the signature does not cover.
func TestRTReadRefusesSmuggledContent(t *testing.T) {
	p := rtSigned(t, rtKeyA)
	dir := t.TempDir()
	path := filepath.Join(dir, "permit.json")
	if err := Write(path, p); err != nil {
		t.Fatal(err)
	}
	good, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	// body is the file without its final closing brace, so a field can be appended.
	body := strings.TrimSuffix(strings.TrimRight(string(good), "\n"), "}")
	cases := map[string]string{
		"case variant key":    body + "\n  ,\"ALLOW\": false\n}\n",
		"unknown field":       body + "\n  ,\"approved_by\": \"x\"\n}\n",
		"unknown nested":      strings.Replace(string(good), `"target": {`, `"target": {"org": "x",`, 1),
		"second document":     string(good) + string(good),
		"trailing garbage":    string(good) + "x",
		"not json":            "allow: true\n",
		"array instead":       "[" + string(good) + "]",
		"duplicate allow key": body + "\n  ,\"allow\": false\n}\n",
	}
	for name, content := range cases {
		t.Run(name, func(t *testing.T) {
			if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
				t.Fatal(err)
			}
			got, err := Read(path)
			if err != nil {
				t.Logf("refused by Read: %v", err)
				return
			}
			// A file that decodes must still not verify unless it decodes to the signed value.
			err = Verify(got, NewHMAC(rtKeyA), rtExpect())
			if err == nil {
				t.Fatalf("%s: a smuggled PERMIT verified", name)
			}
			t.Logf("refused by Verify: %v", err)
		})
	}
}
