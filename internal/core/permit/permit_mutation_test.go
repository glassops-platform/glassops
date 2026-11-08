package permit

import (
	"errors"
	"testing"
	"time"

	"github.com/glassops-platform/glassops/internal/core/config"
	"github.com/glassops-platform/glassops/internal/providers"
)

// permit.go checkKey: `distinct < MinKeyDistinctBytes`. Exactly the minimum is accepted and one
// fewer is refused, which a boundary mutant (`<=`) would get wrong.
func TestMutKeyDistinctBoundary(t *testing.T) {
	mk := func(n int) []byte {
		k := make([]byte, 32)
		for i := range k {
			k[i] = byte(i % n)
		}
		return k
	}
	if err := checkKey(mk(MinKeyDistinctBytes)); err != nil {
		t.Fatalf("exactly %d distinct byte values is enough: %v", MinKeyDistinctBytes, err)
	}
	if err := checkKey(mk(MinKeyDistinctBytes - 1)); err == nil {
		t.Fatal("one fewer distinct byte value must be refused")
	}
}

// permit.go MaxClockSkew: the constant is five minutes and the edge is inclusive.
func TestMutClockSkewEdge(t *testing.T) {
	if MaxClockSkew != 5*time.Minute {
		t.Fatalf("MaxClockSkew = %s, want 5m", MaxClockSkew)
	}
	s := NewHMAC(testKey(1))
	issue := func(at time.Time) error {
		p := syntheticPermit()
		p.IssuedAt = at
		p.ExpiresAt = at.Add(time.Hour)
		if err := Sign(&p, s); err != nil {
			t.Fatal(err)
		}
		want := syntheticExpect()
		want.MaxTTL = 2 * time.Hour
		return Verify(p, s, want)
	}
	if err := issue(testNow.Add(MaxClockSkew)); err != nil {
		t.Fatalf("issued exactly at the skew limit verifies: %v", err)
	}
	if err := issue(testNow.Add(MaxClockSkew + time.Second)); !errors.Is(err, ErrLifetime) {
		t.Fatalf("issued past the skew limit: %v", err)
	}
	z := syntheticPermit()
	z.IssuedAt = time.Time{}
	if err := Sign(&z, s); err != nil {
		t.Fatal(err)
	}
	if err := Verify(z, s, syntheticExpect()); !errors.Is(err, ErrLifetime) {
		t.Fatalf("no issued_at: %v", err)
	}
}

// permit.go checkLifetime: `life <= 0`, `life > limit`, and the default limit.
func TestMutLifetimeBounds(t *testing.T) {
	s := NewHMAC(testKey(1))
	try := func(life time.Duration, maxTTL time.Duration) error {
		p := syntheticPermit()
		p.ExpiresAt = p.IssuedAt.Add(life)
		if err := Sign(&p, s); err != nil {
			t.Fatal(err)
		}
		want := syntheticExpect()
		want.MaxTTL = maxTTL
		return Verify(p, s, want)
	}
	// IssuedAt is a minute before testNow, so every life below keeps the PERMIT unexpired.
	if err := try(time.Hour, time.Hour); err != nil {
		t.Fatalf("a life equal to the limit is allowed: %v", err)
	}
	if err := try(time.Hour+time.Nanosecond, time.Hour); !errors.Is(err, ErrLifetime) {
		t.Fatalf("a life one tick over the limit: %v", err)
	}
	if err := try(time.Hour+time.Nanosecond, 0); err != nil && config.MaxPermitTTL >= time.Hour+time.Nanosecond {
		t.Fatalf("the default limit is config.MaxPermitTTL: %v", err)
	}
	if err := try(config.MaxPermitTTL+time.Nanosecond, 0); !errors.Is(err, ErrLifetime) {
		t.Fatalf("a life over the default limit: %v", err)
	}
	if err := try(config.MaxPermitTTL, 0); err != nil && errors.Is(err, ErrLifetime) {
		t.Fatalf("a life equal to the default limit is allowed: %v", err)
	}
	// A negative limit falls back to the default, so a one hour life passes it.
	if err := try(time.Hour, -time.Hour); err != nil {
		t.Fatalf("a negative limit means the default: %v", err)
	}
}

// permit.go checkLifetime: expiry before issue. An expiry at the issue time and one before it
// are refused. Verify checks expiry against Now first, so Now is set before the issue time.
func TestMutLifetimeExpiresBeforeIssued(t *testing.T) {
	s := NewHMAC(testKey(1))
	for _, d := range []time.Duration{0, -time.Second} {
		p := syntheticPermit()
		p.IssuedAt = testNow.Add(time.Minute)
		p.ExpiresAt = p.IssuedAt.Add(d)
		if err := Sign(&p, s); err != nil {
			t.Fatal(err)
		}
		want := syntheticExpect()
		err := Verify(p, s, want)
		if err == nil {
			t.Fatalf("expiry %s from issue verified", d)
		}
	}
}

// permit.go deploy command binding: a permit that binds nothing refuses a wanted command, a
// different command refuses, the same command passes, and no wanted command skips the check.
func TestMutDeployCommandBinding(t *testing.T) {
	s := NewHMAC(testKey(1))
	cmdA := DeployCommandDigest([]string{"deploy", "--all"})
	cmdB := DeployCommandDigest([]string{"deploy", "--none"})
	mkPermit := func(bound string) Permit {
		p := syntheticPermit()
		p.DeployCommandSHA256 = bound
		if err := Sign(&p, s); err != nil {
			t.Fatal(err)
		}
		return p
	}
	want := func(d string) Expect {
		e := syntheticExpect()
		e.DeployCommandSHA256 = d
		return e
	}
	if err := Verify(mkPermit(""), s, want(cmdA)); !errors.Is(err, ErrMismatch) {
		t.Fatalf("a permit with no bound command must refuse one: %v", err)
	}
	if err := Verify(mkPermit(cmdB), s, want(cmdA)); !errors.Is(err, ErrMismatch) {
		t.Fatalf("a different command: %v", err)
	}
	if err := Verify(mkPermit(cmdA), s, want(cmdA)); err != nil {
		t.Fatalf("the same command: %v", err)
	}
	if err := Verify(mkPermit(cmdA), s, want("")); err != nil {
		t.Fatalf("no wanted command skips the check: %v", err)
	}
}

// permit.go ChangeSetDigest: a nil destructive list digests like an empty one, and a list with
// members digests differently from both. The mutant `destructive != nil` would drop a real list
// and turn nil into JSON null.
func TestMutChangeSetDigestDestructive(t *testing.T) {
	base := providers.ChangeSet{Provider: "salesforce", Dir: "d", Manifest: "m", SHA256: "plan"}
	nilList := ChangeSetDigest(base)
	empty := base
	empty.Destructive = []string{}
	if got := ChangeSetDigest(empty); got != nilList {
		t.Fatal("a nil and an empty destructive list must digest alike")
	}
	one := base
	one.Destructive = []string{"ApexClass:Old"}
	other := base
	other.Destructive = []string{"ApexClass:Other"}
	if ChangeSetDigest(one) == nilList {
		t.Fatal("a destructive member must change the digest")
	}
	if ChangeSetDigest(one) == ChangeSetDigest(other) {
		t.Fatal("different destructive members must digest differently")
	}
}
