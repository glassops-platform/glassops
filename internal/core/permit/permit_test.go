package permit

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/glassops-platform/glassops/internal/core/policy"
)

var testNow = time.Date(2026, 10, 31, 12, 0, 0, 0, time.UTC)

// testKey is a synthetic key of 32 bytes, not a real one.
func testKey(b byte) []byte { return bytes.Repeat([]byte{b}, 32) }

func canonID(id string) string {
	if len(id) > 15 {
		return id[:15]
	}
	return id
}

// syntheticPermit is a PERMIT built by the test. The Ids and hashes are made up.
func syntheticPermit() Permit {
	return Permit{
		Schema:       Schema,
		ID:           "0123456789abcdef",
		Allow:        true,
		IssuedAt:     testNow.Add(-time.Minute),
		ExpiresAt:    testNow.Add(time.Hour),
		Target:       Target{Name: "prod", Tier: "production", ID: "00D000000000001AAA", Kind: "production"},
		Commit:       "c0ffee",
		PlanSHA256:   "plan-hash",
		ConfigSHA256: "config-hash",
		Validation:   &Validation{DeployID: "0Af000000000001", TestLevel: "RunLocalTests", Passed: true},
		Checks:       []policy.Result{{Rule: "plan_not_empty", Outcome: policy.Pass, Detail: "1 change"}},
	}
}

func syntheticExpect() Expect {
	return Expect{
		TargetName: "prod", TargetID: "00D000000000001", Commit: "c0ffee",
		PlanSHA256: "plan-hash", ConfigSHA256: "config-hash", Now: testNow, CanonicalID: canonID,
	}
}

func signed(t *testing.T, s Signer) Permit {
	t.Helper()
	p := syntheticPermit()
	if err := Sign(&p, s); err != nil {
		t.Fatal(err)
	}
	return p
}

func TestSignVerifyRoundTrip(t *testing.T) {
	s := NewHMAC(testKey(1))
	p := signed(t, s)
	if p.Signature.Alg != "HMAC-SHA256" || len(p.Signature.KeyID) != 8 || p.Signature.Value == "" {
		t.Fatalf("signature: %+v", p.Signature)
	}
	if err := Verify(p, s, syntheticExpect()); err != nil {
		t.Fatal(err)
	}
}

func TestVerifyThroughWriteAndRead(t *testing.T) {
	s := NewHMAC(testKey(1))
	p := signed(t, s)
	path := filepath.Join(t.TempDir(), "sub", "permit.json")
	if err := Write(path, p); err != nil {
		t.Fatal(err)
	}
	got, err := Read(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := Verify(got, s, syntheticExpect()); err != nil {
		t.Fatal(err)
	}
}

func TestCanonicalIsStable(t *testing.T) {
	a := syntheticPermit()
	// The same PERMIT read from JSON whose keys come in another order, with times written
	// in another zone and no checks array detail difference.
	b, err := json.Marshal(a)
	if err != nil {
		t.Fatal(err)
	}
	var generic map[string]any
	if err := json.Unmarshal(b, &generic); err != nil {
		t.Fatal(err)
	}
	generic["issued_at"] = a.IssuedAt.In(time.FixedZone("x", 2*3600)).Format(time.RFC3339Nano)
	reordered := `{"signature":{"value":"zzz","alg":""},"checks":[{"detail":"1 change","outcome":"pass","rule":"plan_not_empty"}],`
	for _, k := range []string{"validation", "target", "plan_sha256", "config_sha256", "commit", "expires_at", "issued_at", "allow", "id", "schema"} {
		v, _ := json.Marshal(generic[k])
		reordered += `"` + k + `":` + string(v) + `,`
	}
	reordered = strings.TrimSuffix(reordered, ",") + `}`
	var other Permit
	if err := json.Unmarshal([]byte(reordered), &other); err != nil {
		t.Fatal(err)
	}
	ca, err := Canonical(a)
	if err != nil {
		t.Fatal(err)
	}
	cb, err := Canonical(other)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(ca, cb) {
		t.Fatalf("canonical bytes differ:\n%s\n%s", ca, cb)
	}
	if bytes.ContainsAny(ca, "\n\t") || bytes.Contains(ca, []byte(`": `)) || bytes.Contains(ca, []byte(`, "`)) {
		t.Fatalf("canonical bytes have insignificant whitespace: %s", ca)
	}
	// Keys sorted at the top level and in a nested object.
	s := string(ca)
	if !(strings.Index(s, `"allow"`) < strings.Index(s, `"checks"`) && strings.Index(s, `"checks"`) < strings.Index(s, `"commit"`)) {
		t.Fatalf("top level keys not sorted: %s", s)
	}
	if !strings.Contains(s, `"checks":[{"detail":"1 change","outcome":"pass","rule":"plan_not_empty"}]`) {
		t.Fatalf("nested keys not sorted: %s", s)
	}
	if !strings.Contains(s, `"signature":{"alg":"","value":""}`) {
		t.Fatalf("signature value not emptied: %s", s)
	}
}

func TestCanonicalNilAndEmptyChecksMatch(t *testing.T) {
	a, b := syntheticPermit(), syntheticPermit()
	a.Checks, b.Checks = nil, []policy.Result{}
	ca, _ := Canonical(a)
	cb, _ := Canonical(b)
	if !bytes.Equal(ca, cb) {
		t.Fatalf("%s != %s", ca, cb)
	}
}

func TestVerifyTamper(t *testing.T) {
	s := NewHMAC(testKey(1))
	cases := []struct {
		name string
		edit func(p *Permit)
	}{
		{"allow", func(p *Permit) { p.Allow = !p.Allow }},
		{"id", func(p *Permit) { p.ID = "other" }},
		{"expires", func(p *Permit) { p.ExpiresAt = p.ExpiresAt.Add(24 * time.Hour) }},
		{"issued", func(p *Permit) { p.IssuedAt = p.IssuedAt.Add(-time.Hour) }},
		{"target name", func(p *Permit) { p.Target.Name = "dev" }},
		{"target tier", func(p *Permit) { p.Target.Tier = "dev" }},
		{"target id", func(p *Permit) { p.Target.ID = "00D000000000002AAA" }},
		{"target kind", func(p *Permit) { p.Target.Kind = "scratch" }},
		{"commit", func(p *Permit) { p.Commit = "bad" }},
		{"plan", func(p *Permit) { p.PlanSHA256 = "bad" }},
		{"config", func(p *Permit) { p.ConfigSHA256 = "bad" }},
		{"validation", func(p *Permit) { p.Validation.Passed = false }},
		{"checks", func(p *Permit) { p.Checks[0].Outcome = policy.Fail }},
		{"sod", func(p *Permit) { p.SeparationOfDuties = &SoD{Mode: "waived_by_config"} }},
		{"key id", func(p *Permit) { p.Signature.KeyID = "deadbeef" }},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			p := signed(t, s)
			c.edit(&p)
			err := Verify(p, s, syntheticExpect())
			if !errors.Is(err, ErrSignature) {
				t.Fatalf("want ErrSignature, got %v", err)
			}
		})
	}
}

func TestVerifySignatureProblems(t *testing.T) {
	s := NewHMAC(testKey(1))
	t.Run("unsigned", func(t *testing.T) {
		p := syntheticPermit()
		if err := Verify(p, s, syntheticExpect()); !errors.Is(err, ErrSignature) {
			t.Fatal(err)
		}
	})
	t.Run("wrong algorithm", func(t *testing.T) {
		p := signed(t, s)
		p.Signature.Alg = "none"
		if err := Verify(p, s, syntheticExpect()); !errors.Is(err, ErrSignature) {
			t.Fatal(err)
		}
	})
	t.Run("another key", func(t *testing.T) {
		p := signed(t, NewHMAC(testKey(2)))
		err := Verify(p, s, syntheticExpect())
		if !errors.Is(err, ErrSignature) {
			t.Fatal(err)
		}
	})
	t.Run("another key without key id", func(t *testing.T) {
		p := signed(t, NewHMAC(testKey(2)))
		p.Signature.KeyID = signed(t, s).Signature.KeyID
		if err := Verify(p, s, syntheticExpect()); !errors.Is(err, ErrSignature) {
			t.Fatal(err)
		}
	})
	t.Run("value not base64", func(t *testing.T) {
		p := signed(t, s)
		p.Signature.Value = "!!!"
		if err := Verify(p, s, syntheticExpect()); !errors.Is(err, ErrSignature) {
			t.Fatal(err)
		}
	})
}

func TestVerifyDeniedExpiredAndMismatch(t *testing.T) {
	s := NewHMAC(testKey(1))
	resign := func(edit func(p *Permit)) Permit {
		p := syntheticPermit()
		edit(&p)
		if err := Sign(&p, s); err != nil {
			t.Fatal(err)
		}
		return p
	}
	cases := []struct {
		name   string
		permit Permit
		want   func(e *Expect)
		err    error
	}{
		{"denied", resign(func(p *Permit) { p.Allow = false }), nil, ErrDenied},
		{"expired", resign(func(p *Permit) {}), func(e *Expect) { e.Now = testNow.Add(2 * time.Hour) }, ErrExpired},
		{"expires exactly now", resign(func(p *Permit) {}), func(e *Expect) { e.Now = testNow.Add(time.Hour) }, ErrExpired},
		{"no clock", resign(func(p *Permit) {}), func(e *Expect) { e.Now = time.Time{} }, ErrExpired},
		{"target name", resign(func(p *Permit) {}), func(e *Expect) { e.TargetName = "staging" }, ErrMismatch},
		{"target id", resign(func(p *Permit) {}), func(e *Expect) { e.TargetID = "00D000000000002" }, ErrMismatch},
		{"empty target id", resign(func(p *Permit) {}), func(e *Expect) { e.TargetID = "" }, ErrMismatch},
		{"commit", resign(func(p *Permit) {}), func(e *Expect) { e.Commit = "other" }, ErrMismatch},
		{"empty commit", resign(func(p *Permit) {}), func(e *Expect) { e.Commit = "" }, ErrMismatch},
		{"plan", resign(func(p *Permit) {}), func(e *Expect) { e.PlanSHA256 = "other" }, ErrMismatch},
		{"config", resign(func(p *Permit) {}), func(e *Expect) { e.ConfigSHA256 = "other" }, ErrMismatch},
		{"schema", resign(func(p *Permit) { p.Schema = "glassops.permit/v0" }), nil, ErrMismatch},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			want := syntheticExpect()
			if c.want != nil {
				c.want(&want)
			}
			err := Verify(c.permit, s, want)
			if !errors.Is(err, c.err) {
				t.Fatalf("want %v, got %v", c.err, err)
			}
			if strings.Contains(err.Error(), base64.StdEncoding.EncodeToString(testKey(1))) {
				t.Fatal("error contains the key")
			}
		})
	}
}

func TestVerifyOrder(t *testing.T) {
	s := NewHMAC(testKey(1))
	p := syntheticPermit()
	p.Allow = false
	p.ExpiresAt = testNow.Add(-time.Hour)
	if err := Sign(&p, s); err != nil {
		t.Fatal(err)
	}
	want := syntheticExpect()
	want.Commit = "other"
	if err := Verify(p, s, want); !errors.Is(err, ErrDenied) {
		t.Fatalf("denied comes before expiry and mismatch: %v", err)
	}
	p.Allow = true
	if err := Sign(&p, s); err != nil {
		t.Fatal(err)
	}
	if err := Verify(p, s, want); !errors.Is(err, ErrExpired) {
		t.Fatalf("expiry comes before mismatch: %v", err)
	}
	// A bad signature comes before everything.
	p.Commit = "tampered"
	if err := Verify(p, s, want); !errors.Is(err, ErrSignature) {
		t.Fatalf("signature comes first: %v", err)
	}
}

func TestVerifyTargetIDCanonicalForm(t *testing.T) {
	s := NewHMAC(testKey(1))
	p := syntheticPermit()
	p.Target.ID = "00D000000000001" // the 15 character form
	if err := Sign(&p, s); err != nil {
		t.Fatal(err)
	}
	if err := Verify(p, s, syntheticExpect()); err != nil {
		t.Fatal(err)
	}
	// Without a canonicaliser the Ids are compared as written.
	want := syntheticExpect()
	want.CanonicalID = nil
	q := signed(t, s)
	if err := Verify(q, s, want); !errors.Is(err, ErrMismatch) {
		t.Fatalf("want ErrMismatch, got %v", err)
	}
}

func TestParseKey(t *testing.T) {
	good := base64.StdEncoding.EncodeToString(testKey(7))
	if len(good) != 44 {
		t.Fatalf("len %d", len(good))
	}
	for name, in := range map[string]string{
		"padded":      good,
		"unpadded":    strings.TrimRight(good, "="),
		"whitespace":  " " + good + "\n",
		"longer key":  base64.StdEncoding.EncodeToString(bytes.Repeat([]byte{1}, 64)),
		"exactly 32b": good,
	} {
		if _, err := ParseKey(in); err != nil {
			t.Errorf("%s: %v", name, err)
		}
	}
	for name, in := range map[string]string{
		"empty":     "",
		"short":     base64.StdEncoding.EncodeToString(bytes.Repeat([]byte{1}, 31)),
		"not b64":   "this is not base64!!",
		"url-safe":  "_-_-_-_-_-_-_-_-_-_-_-_-_-_-_-_-_-_-_-_-_-_-_-",
		"plain key": "hunter2-hunter2-hunter2-hunter2-hunter2",
	} {
		_, err := ParseKey(in)
		if err == nil {
			t.Errorf("%s: want an error", name)
			continue
		}
		if in != "" && strings.Contains(err.Error(), in) {
			t.Errorf("%s: error contains the input", name)
		}
	}
}

func TestShortKeySignerRefuses(t *testing.T) {
	s := NewHMAC([]byte("short"))
	p := syntheticPermit()
	if err := Sign(&p, s); err == nil {
		t.Fatal("a short key signed")
	}
	if err := s.Verify([]byte("x"), base64.StdEncoding.EncodeToString([]byte("x"))); err == nil {
		t.Fatal("a short key verified")
	}
}

func TestKeyIDIsStableAndDoesNotExposeKey(t *testing.T) {
	a, b := NewHMAC(testKey(1)).(KeyIdentifier), NewHMAC(testKey(2)).(KeyIdentifier)
	if a.KeyID() == b.KeyID() || a.KeyID() != NewHMAC(testKey(1)).(KeyIdentifier).KeyID() {
		t.Fatalf("key ids: %s %s", a.KeyID(), b.KeyID())
	}
	// SHA-256 of 32 bytes of 0x01, first 8 hex characters.
	if got := a.KeyID(); len(got) != 8 {
		t.Fatalf("key id %q", got)
	}
}

func TestReadRefusesUnknownFieldsAndMissingFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "permit.json")
	p := signed(t, NewHMAC(testKey(1)))
	if err := Write(path, p); err != nil {
		t.Fatal(err)
	}
	data, _ := os.ReadFile(path)
	if err := os.WriteFile(path, bytes.Replace(data, []byte(`"allow"`), []byte(`"extra": 1, "allow"`), 1), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := Read(path); !errors.Is(err, ErrUnreadable) {
		t.Fatalf("unknown field: %v", err)
	}
	if err := os.WriteFile(path, []byte("not json"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := Read(path); !errors.Is(err, ErrUnreadable) {
		t.Fatalf("bad json: %v", err)
	}
	if _, err := Read(filepath.Join(dir, "none.json")); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("missing: %v", err)
	}
}
