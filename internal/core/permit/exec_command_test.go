package permit

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"
)

func TestExecCommandBinding(t *testing.T) {
	s := NewHMAC(testKey(1))
	a := ExecCommandDigest("node scripts/api-reserve.js")
	b := ExecCommandDigest("node scripts/other.js")
	mk := func(bound ...string) Permit {
		p := syntheticPermit()
		p.ExecCommandsSHA256 = ExecCommandDigests(bound)
		if err := Sign(&p, s); err != nil {
			t.Fatal(err)
		}
		return p
	}
	want := func(d string) Expect {
		e := syntheticExpect()
		e.ExecCommandSHA256 = d
		return e
	}
	if err := Verify(mk(), s, want(a)); !errors.Is(err, ErrMismatch) || !strings.Contains(err.Error(), "--exec-command") {
		t.Fatalf("a permit that binds no command must refuse one: %v", err)
	}
	if err := Verify(mk("node scripts/other.js"), s, want(a)); !errors.Is(err, ErrMismatch) {
		t.Fatalf("another command: %v", err)
	}
	if err := Verify(mk("node scripts/other.js", "node scripts/api-reserve.js"), s, want(a)); err != nil {
		t.Fatalf("a listed command: %v", err)
	}
	if err := Verify(mk("node scripts/other.js"), s, want(b)); err != nil {
		t.Fatalf("the one listed command: %v", err)
	}
	if err := Verify(mk(), s, want("")); err != nil {
		t.Fatalf("no wanted command skips the check: %v", err)
	}
	// The list is signed: adding a command after signing breaks the signature.
	p := mk("node scripts/other.js")
	p.ExecCommandsSHA256 = append(p.ExecCommandsSHA256, a)
	if err := Verify(p, s, want(a)); !errors.Is(err, ErrSignature) {
		t.Fatalf("a command added after signing: %v", err)
	}
}

func TestExecCommandDigests(t *testing.T) {
	if got := ExecCommandDigests(nil); got != nil {
		t.Fatalf("no commands = %v, want nil", got)
	}
	got := ExecCommandDigests([]string{"b", "a", "b", ""})
	if len(got) != 2 || got[0] > got[1] {
		t.Fatalf("digests = %v, want two sorted", got)
	}
	// The run text is one argument: "a b" is not ["a", "b"].
	if ExecCommandDigest("a b") == DeployCommandDigest([]string{"a", "b"}) {
		t.Fatal("the run text must digest as one element")
	}
	if ExecCommandDigest("a b") != DeployCommandDigest([]string{"a b"}) {
		t.Fatal("ExecCommandDigest is DeployCommandDigest of the one element argv")
	}
}

// A PERMIT with no exec commands encodes exactly as before the field existed, so its canonical
// bytes and signature are unchanged.
func TestNoExecCommandsLeavesThePermitAlone(t *testing.T) {
	data, err := json.Marshal(syntheticPermit())
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(data), "exec_commands") {
		t.Fatalf("permit JSON = %s, want no exec_commands key", data)
	}
}
