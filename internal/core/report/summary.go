package report

import (
	"fmt"
	"strings"

	"github.com/glassops-platform/glassops/internal/core/redact"
	"github.com/glassops-platform/glassops/internal/core/run"
)

// detailLimit is the longest detail shown in the step table.
const detailLimit = 90

// text masks email-shaped values and makes free text safe for a table cell or list line.
func text(s string) string { return clean(redact.Text(s)) }

// Summary renders the markdown summary. The same bundle always renders the same text.
func Summary(b *Bundle) string {
	var sb strings.Builder
	fmt.Fprintf(&sb, "## glassops report: %s\n\n", b.Worst())

	sb.WriteString("| Step | Status | Target | Detail |\n|---|---|---|---|\n")
	for _, r := range b.Results {
		fmt.Fprintf(&sb, "| %s | %s | %s | %s |\n", text(r.Step), r.Status, text(r.Target), shortDetail(r))
	}
	sb.WriteString("\n")

	for _, r := range b.Results {
		if r.Status != run.StatusFail && r.Status != run.StatusError {
			continue
		}
		fmt.Fprintf(&sb, "### %s: %s\n\n", text(r.Step), r.Status)
		for _, c := range failingChecks(r) {
			fmt.Fprintf(&sb, "- %s (%s): %s\n", text(c.Name), c.Status, text(c.Detail))
		}
		for _, e := range r.Errors {
			fmt.Fprintf(&sb, "- error: %s\n", text(e))
		}
		sb.WriteString("\n")
	}

	if b.Permit != nil {
		writePermit(&sb, b)
	}
	if id := b.DeployID(); id != "" {
		fmt.Fprintf(&sb, "### Deploy\n\nDeploy Id: `%s`\n\n", text(id))
	}
	if len(b.Problems) > 0 {
		sb.WriteString("### Unreadable evidence\n\n")
		for _, p := range b.Problems {
			fmt.Fprintf(&sb, "- %s\n", text(p))
		}
		sb.WriteString("\n")
	}
	return sb.String()
}

// shortDetail is one line for the step table.
func shortDetail(r run.Result) string {
	var d string
	switch {
	case len(failingChecks(r)) > 0:
		c := failingChecks(r)[0]
		d = c.Name
		if c.Detail != "" {
			d += ": " + c.Detail
		}
	case len(r.Errors) > 0:
		d = r.Errors[0]
	case len(r.Checks) > 0:
		d = fmt.Sprintf("%d checks", len(r.Checks))
		if len(r.Checks) == 1 {
			d = "1 check"
		}
	}
	d = text(d)
	if r := []rune(d); len(r) > detailLimit {
		d = string(r[:detailLimit-3]) + "..."
	}
	return d
}

func writePermit(sb *strings.Builder, b *Bundle) {
	p := b.Permit
	decision := "deny"
	if p.Allow {
		decision = "allow"
	}
	fmt.Fprintf(sb, "### PERMIT: %s\n\n", decision)
	fmt.Fprintf(sb, "Target: %s (%s). Issued %s, expires %s.\n\n",
		text(p.Target.Name), text(p.Target.Tier),
		p.IssuedAt.UTC().Format("2006-01-02 15:04Z"), p.ExpiresAt.UTC().Format("2006-01-02 15:04Z"))
	if len(p.Checks) > 0 {
		sb.WriteString("| Rule | Outcome | Detail |\n|---|---|---|\n")
		for _, c := range p.Checks {
			fmt.Fprintf(sb, "| %s | %s | %s |\n", text(c.Rule), c.Outcome, text(c.Detail))
		}
		sb.WriteString("\n")
	}
	if s := p.SeparationOfDuties; s != nil {
		fmt.Fprintf(sb, "Separation of duties: %s", text(s.Mode))
		if s.Mechanism != "" {
			fmt.Fprintf(sb, " by %s", text(s.Mechanism))
		}
		if s.Approver != "" {
			fmt.Fprintf(sb, ", approver %s", text(s.Approver))
		}
		sb.WriteString(".\n\n")
	}
}
