package permit

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/glassops-platform/glassops/internal/core/config"
)

// The fuzz targets in this file check that PERMIT parsing and verification never panic, and
// that no input makes Verify accept a PERMIT it should refuse.

// mutKey and mutOtherKey are synthetic 32 byte keys, not real ones. Each holds 32 different
// bytes, because ParseKey and NewHMAC refuse a key with fewer than 8.
var (
	mutKey      = []byte("mut-synthetic-key-0123456789abcd")
	mutOtherKey = []byte("mut-other-synthetic-key-01234567")
)

var mutNow = time.Date(2026, 11, 8, 12, 0, 0, 0, time.UTC)

func mutPermit() Permit {
	return Permit{
		Schema:       Schema,
		ID:           "0123456789abcdef",
		Allow:        true,
		IssuedAt:     mutNow.Add(-time.Minute),
		ExpiresAt:    mutNow.Add(time.Hour),
		Target:       Target{Name: "prod", Tier: "prod", ID: "00D000000000001AAA", Kind: "production"},
		Commit:       "c0ffee",
		PlanSHA256:   "plan-hash",
		ConfigSHA256: "config-hash",
		// The change set hash is bound like the plan hash. The fuzzed fields below cover them.
		ChangeSetSHA256: "change-set-hash",
	}
}

func mutExpect() Expect {
	return Expect{TargetName: "prod", TargetID: "00D000000000001AAA", Commit: "c0ffee", PlanSHA256: "plan-hash", ConfigSHA256: "config-hash", ChangeSetSHA256: "change-set-hash", Now: mutNow}
}

// FuzzMutPermitRead feeds arbitrary bytes to Read. Whatever it returns, Canonical and Verify
// must not panic, and a PERMIT nobody signed with the key must never verify.
func FuzzMutPermitRead(f *testing.F) {
	dir := f.TempDir()
	good := mutPermit()
	if err := Sign(&good, NewHMAC(mutKey)); err != nil {
		f.Fatal(err)
	}
	goodPath := filepath.Join(dir, "good.json")
	if err := Write(goodPath, good); err != nil {
		f.Fatal(err)
	}
	seed, err := os.ReadFile(goodPath)
	if err != nil {
		f.Fatal(err)
	}
	f.Add(seed)
	f.Add([]byte(`{}`))
	f.Add([]byte(`{"schema":"glassops/permit/v1","allow":true}`))
	f.Add([]byte(`{"allow":true} {"allow":false}`))
	f.Add([]byte(`{"unknown":1}`))
	f.Add([]byte(`null`))
	f.Add([]byte(``))
	f.Add([]byte("{\"signature\":{\"alg\":\"HMAC-SHA256\",\"value\":\"AAAA\"}}"))
	path := filepath.Join(dir, "fuzz.json")
	f.Fuzz(func(t *testing.T, data []byte) {
		if err := os.WriteFile(path, data, 0o644); err != nil {
			t.Fatal(err)
		}
		p, err := Read(path)
		if err != nil {
			return
		}
		_, _ = Canonical(p)
		verr := Verify(p, NewHMAC(mutKey), mutExpect())
		if verr == nil {
			// A parsed file may only verify when it really carries a valid signature for this key.
			want := p
			want.Signature.Value = ""
			if err := Sign(&want, NewHMAC(mutKey)); err != nil || want.Signature.Value != p.Signature.Value {
				t.Fatalf("accepted a PERMIT whose signature is not ours: %q", data)
			}
			if !p.Allow || !mutNow.Before(p.ExpiresAt) {
				t.Fatalf("accepted a denying or expired PERMIT: %q", data)
			}
		}
	})
}

// FuzzMutPermitVerify signs a PERMIT built from fuzzed fields and checks Verify against fuzzed
// expectations. Verify accepts only when every bound field matches exactly.
func FuzzMutPermitVerify(f *testing.F) {
	f.Add("prod", "00D000000000001AAA", "c0ffee", "plan-hash", "config-hash", true, int64(3600), "prod", "00D000000000001AAA", "c0ffee", "plan-hash", "config-hash")
	f.Add("prod", "00D000000000001AAA", "c0ffee", "plan-hash", "config-hash", true, int64(3600), "qa", "00D000000000001AAA", "c0ffee", "plan-hash", "config-hash")
	f.Add("", "", "", "", "", true, int64(1), "", "", "", "", "")
	f.Add("prod", "id", "sha", "p", "c", false, int64(60), "prod", "id", "sha", "p", "c")
	f.Add("prod\n::set-output name=x::y", "id", "sha", "p", "c", true, int64(-5), "prod", "id", "sha", "p", "c")
	f.Fuzz(func(t *testing.T, name, id, commit, plan, cfgHash string, allow bool, ttlSeconds int64, wName, wID, wCommit, wPlan, wCfg string) {
		p := mutPermit()
		p.Allow = allow
		p.Target = Target{Name: name, ID: id}
		p.Commit, p.PlanSHA256, p.ConfigSHA256 = commit, plan, cfgHash
		// Keep the offset small enough that the time arithmetic cannot overflow.
		p.ExpiresAt = mutNow.Add(time.Duration(ttlSeconds%1_000_000) * time.Second)
		signer := NewHMAC(mutKey)
		if err := Sign(&p, signer); err != nil {
			t.Fatal(err)
		}
		want := Expect{TargetName: wName, TargetID: wID, Commit: wCommit, PlanSHA256: wPlan, ConfigSHA256: wCfg, ChangeSetSHA256: "change-set-hash", Now: mutNow}
		err := Verify(p, signer, want)
		matches := allow && mutNow.Before(p.ExpiresAt) && p.ExpiresAt.Sub(p.IssuedAt) <= config.MaxPermitTTL &&
			name == wName && id == wID && wID != "" &&
			commit == wCommit && wCommit != "" && plan == wPlan && wPlan != "" && cfgHash == wCfg && wCfg != ""
		if matches && err != nil {
			t.Fatalf("refused a PERMIT that matches: %v", err)
		}
		if !matches && err == nil {
			t.Fatalf("accepted a PERMIT that does not match: permit %+v want %+v", p, want)
		}
		// Changing any bound field after signing must break the signature.
		tampered := p
		tampered.Commit += "x"
		if err := Verify(tampered, signer, want); !errors.Is(err, ErrSignature) {
			t.Fatalf("a tampered PERMIT must fail the signature check, got %v", err)
		}
		// A different key never verifies.
		if err := Verify(p, NewHMAC(mutOtherKey), want); err == nil {
			t.Fatal("verified with the wrong key")
		}
	})
}

// FuzzMutParseKey checks that no key text panics ParseKey, that a key it accepts is long
// enough, and that an error never repeats the key.
func FuzzMutParseKey(f *testing.F) {
	f.Add("MDEyMzQ1Njc4OWFiY2RlZjAxMjM0NTY3ODlhYmNkZWY=")
	f.Add("MDEyMzQ1Njc4OWFiY2RlZjAxMjM0NTY3ODlhYmNkZWY")
	f.Add("")
	f.Add("   ")
	f.Add("c2hvcnQ=")
	f.Add("not base64 at all!!")
	f.Add("AAAA\nAAAA")
	f.Fuzz(func(t *testing.T, s string) {
		key, err := ParseKey(s)
		if err == nil {
			if len(key) < MinKeyBytes {
				t.Fatalf("accepted a %d byte key", len(key))
			}
			return
		}
		// The message must never contain the key text.
		if text := strings.TrimSpace(s); len(text) >= 16 && strings.Contains(err.Error(), text) {
			t.Fatalf("the error repeats the key text")
		}
	})
}
