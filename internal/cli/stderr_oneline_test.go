package cli

import (
	"bytes"
	"errors"
	"strings"
	"testing"
)

// An error printed to stderr stays one line, so a message holding a line break and a
// workflow command cannot start a new line that a CI runner would read as a command.
func TestPrintErrStaysOneLine(t *testing.T) {
	for _, msg := range []string{"bad\n::warning title=x::forged", "bad\r\n::error::forged", "plain"} {
		var buf bytes.Buffer
		printErr(&buf, errors.New(msg))
		out := buf.String()
		if strings.Count(out, "\n") != 1 || !strings.HasSuffix(out, "\n") {
			t.Fatalf("%q: not exactly one line: %q", msg, out)
		}
		if strings.Contains(out, "\r") || !strings.HasPrefix(out, "glassops: ") {
			t.Fatalf("%q: unexpected output %q", msg, out)
		}
	}
}
