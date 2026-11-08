package ci

// Red team tests for the CI output channels. A test here passes when the injection is refused.

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRTGitHubOutputRefusesInjectedLines(t *testing.T) {
	out := filepath.Join(t.TempDir(), "github-output")
	env := func(k string) string {
		if k == "GITHUB_OUTPUT" {
			return out
		}
		return ""
	}
	g := newGitHub(env, &bytes.Buffer{})
	for _, v := range []string{"pass\nstatus=fail", "pass\r\nx=y", "pass\rx"} {
		if err := g.SetOutput("status", v); err == nil {
			t.Fatalf("SetOutput accepted %q", v)
		}
	}
	data, _ := os.ReadFile(out)
	if len(data) != 0 {
		t.Fatalf("GITHUB_OUTPUT was written: %q", data)
	}
}

func TestRTGitHubAnnotationStaysOneLine(t *testing.T) {
	var buf bytes.Buffer
	g := newGitHub(func(string) string { return "" }, &buf)
	g.Annotate(Errors, "rt\n::warning::rt-forged\r\n%0A::add-mask::x")
	lines := strings.Split(strings.TrimRight(buf.String(), "\n"), "\n")
	if len(lines) != 1 || !strings.HasPrefix(lines[0], "::error::") {
		t.Fatalf("annotation output %q", buf.String())
	}
	if want := "::error::rt%0A::warning::rt-forged%0D%0A%250A::add-mask::x"; lines[0] != want {
		t.Fatalf("annotation %q, want %q", lines[0], want)
	}
}

func TestRTDotenvOutputRefusesInjectedLines(t *testing.T) {
	path := filepath.Join(t.TempDir(), "outputs.env")
	cases := []struct{ name, value string }{
		{"status", "pass\nGLASSOPS_PERMIT_KEY=x"},
		{"status", "pass\r\nx=y"},
		{"status\nX", "pass"},
		{"status=x", "pass"},
		{"1status", "pass"},
		{"", "pass"},
	}
	for _, tc := range cases {
		if err := writeOutput(path, tc.name, tc.value); err == nil {
			t.Fatalf("writeOutput(%q, %q) accepted", tc.name, tc.value)
		}
	}
	if data, _ := os.ReadFile(path); len(data) != 0 {
		t.Fatalf("the outputs file was written: %q", data)
	}
}

// Finding F5, now refused: the Bitbucket and GitLab layers printed an annotation as it was,
// so a line break in a message started a log line of its own.
func TestRTPlainAnnotationStaysOneLine(t *testing.T) {
	for name, mk := range map[string]func(*bytes.Buffer) CI{
		"bitbucket": func(b *bytes.Buffer) CI { return newBitbucket(func(string) string { return "" }, b, t.TempDir()) },
		"gitlab":    func(b *bytes.Buffer) CI { return newGitLab(func(string) string { return "" }, b, t.TempDir()) },
	} {
		t.Run(name, func(t *testing.T) {
			var buf bytes.Buffer
			mk(&buf).Annotate(Errors, "rt\nerror: rt-forged\r\nwarning: x")
			lines := strings.Split(strings.TrimRight(buf.String(), "\n"), "\n")
			if len(lines) != 1 || !strings.Contains(lines[0], `rt\nerror: rt-forged\r\nwarning: x`) {
				t.Fatalf("annotation output %q", buf.String())
			}
		})
	}
}
