// Package redact masks values that must never appear in results or CI logs. Results and logs
// are public on public repositories.
package redact

import (
	"regexp"
	"strings"
)

var email = regexp.MustCompile(`[A-Za-z0-9._%+'-]+@[A-Za-z0-9.-]+\.[A-Za-z]{2,}`)

// Username masks a username to its first character and its domain: "u***@example.com".
// A value without an @ is masked whole.
func Username(u string) string {
	local, domain, ok := strings.Cut(u, "@")
	if !ok || local == "" {
		return "***"
	}
	return local[:1] + "***@" + domain
}

// Text masks every email-shaped value in free text, such as an error message from a tool.
func Text(s string) string {
	return email.ReplaceAllStringFunc(s, Username)
}
