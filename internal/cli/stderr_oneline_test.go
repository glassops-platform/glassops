package cli

import (
	"bytes"
	"errors"
	"strings"
	"testing"

	"github.com/glassops-platform/glassops/internal/core/exitcode"
)

// An error printed to stderr stays one line, so a message holding a line break and a
// workflow command cannot start a new line that a CI runner would read as a command.
func TestStderrErrorStaysOneLine(t *testing.T) {
	msg := "bad target\n::warning title=x::forged"
	for _, err := range []error{errors.New(msg), &exitcode.ExitError{Code: exitcode.Error, Err: errors.New(msg)}} {
		var buf bytes.Buffer
		app := &App{Stderr: &buf}
		finishCode(app, err)
		out := buf.String()
		if strings.Count(out, "\n") != 1 || strings.Contains(out, "\n::") {
			t.Fatalf("stderr is not one line: %q", out)
		}
	}
}
