package redact

import (
	"bytes"
	"strings"
	"testing"
)

// fakeToken is built from parts so the source never holds a token shaped literal.
var fakeToken = "00Dxx0000001gPL" + "!" + "AQ4AQFakeTokenValueThatIsLongEnough.abc"

func TestSecrets(t *testing.T) {
	cases := map[string]string{
		"token " + fakeToken + " end":                           "token *** end",
		"url " + "force:" + "//PlatformCLI::5Aep861@example.my": "url ***",
		"Authorization: Bearer abcdefghijklmnopqrstuvwxyz":      "Authorization: ***",
		"org 00Dxx0000001gPLEAY is fine":                        "org 00Dxx0000001gPLEAY is fine",
		"user a@example.com stays":                              "user a@example.com stays",
	}
	for in, want := range cases {
		if got := Secrets(in); got != want {
			t.Errorf("Secrets(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestLineWriterMasksAcrossWrites(t *testing.T) {
	var out bytes.Buffer
	w := NewLineWriter(&out)
	half := len(fakeToken) / 2
	for _, part := range []string{"a " + fakeToken[:half], fakeToken[half:] + " b\nnext ", "line without newline"} {
		if _, err := w.Write([]byte(part)); err != nil {
			t.Fatal(err)
		}
	}
	if strings.Contains(out.String(), "line without") {
		t.Fatal("a partial line was written before Close")
	}
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}
	if got, want := out.String(), "a *** b\nnext line without newline"; got != want {
		t.Fatalf("output = %q, want %q", got, want)
	}
}

func TestLineWriterBoundsALongLine(t *testing.T) {
	var out bytes.Buffer
	w := NewLineWriter(&out)
	_, _ = w.Write(bytes.Repeat([]byte("x"), maxLine+10))
	if out.Len() == 0 {
		t.Fatal("a line past maxLine was held instead of written")
	}
}
