package redact

import "testing"

func TestUsername(t *testing.T) {
	cases := map[string]string{
		"ci-user@example.com":         "c***@example.com",
		"a@b.example.com.integration": "a***@b.example.com.integration",
		"no-at-sign":                  "***",
		"@example.com":                "***",
	}
	for in, want := range cases {
		if got := Username(in); got != want {
			t.Errorf("Username(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestText(t *testing.T) {
	got := Text("user ci.user@example.com.sandbox has not approved this consumer")
	if got != "user c***@example.com.sandbox has not approved this consumer" {
		t.Fatalf("got %q", got)
	}
}
