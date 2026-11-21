package redact

import (
	"bytes"
	"io"
	"regexp"
	"sync"
)

// secretPatterns match login material a Salesforce tool can print: a session Id or access
// token (org Id, "!", token), an sfdx auth URL, and a bearer header.
var secretPatterns = []*regexp.Regexp{
	regexp.MustCompile(`00D[A-Za-z0-9]{12}(?:[A-Za-z0-9]{3})?![A-Za-z0-9._]{20,}`),
	regexp.MustCompile(`force://\S+`),
	regexp.MustCompile(`(?i)bearer\s+[A-Za-z0-9._!~+/=-]{20,}`),
}

// Secrets masks login material in free text as "***". Unlike Text it leaves emails alone.
func Secrets(s string) string {
	for _, re := range secretPatterns {
		s = re.ReplaceAllString(s, "***")
	}
	return s
}

// LineWriter masks Secrets in whatever is written through it, a line at a time, so a value
// split across two writes is still masked. Close writes any last partial line.
type LineWriter struct {
	mu  sync.Mutex
	w   io.Writer
	buf []byte
}

// NewLineWriter returns a LineWriter that writes masked lines to w.
func NewLineWriter(w io.Writer) *LineWriter { return &LineWriter{w: w} }

// maxLine is how long a line may grow before it is written without its newline, so a command
// that never prints one cannot hold memory forever. A value cut at that point may not mask.
const maxLine = 64 << 10

func (l *LineWriter) Write(p []byte) (int, error) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.buf = append(l.buf, p...)
	for {
		i := bytes.IndexByte(l.buf, '\n')
		if i < 0 {
			if len(l.buf) < maxLine {
				return len(p), nil
			}
			i = len(l.buf) - 1
		}
		line := Secrets(string(l.buf[:i+1]))
		l.buf = l.buf[i+1:]
		if _, err := io.WriteString(l.w, line); err != nil {
			return len(p), err
		}
	}
}

// Close writes the last partial line, masked.
func (l *LineWriter) Close() error {
	l.mu.Lock()
	defer l.mu.Unlock()
	if len(l.buf) == 0 {
		return nil
	}
	_, err := io.WriteString(l.w, Secrets(string(l.buf)))
	l.buf = nil
	return err
}
