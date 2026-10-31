package report

import (
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/glassops-platform/glassops/internal/core/policy"
	"github.com/glassops-platform/glassops/internal/core/redact"
	"github.com/glassops-platform/glassops/internal/core/sarif"
)

// ToolName is the name of the glassops run inside the SARIF log.
const ToolName = "glassops"

// SARIF builds the log for a bundle: glassops's own findings first, then any SARIF files
// that were already in the results directory, each with its own tool.
func SARIF(b *Bundle, version string) *sarif.Log {
	own := sarif.Run{Tool: sarif.Tool{Driver: sarif.Driver{
		Name:           ToolName,
		Version:        version,
		InformationURI: "https://github.com/glassops-platform/glassops",
	}}}
	own.Results = []sarif.Result{}

	own.Results = append(own.Results, findings(b)...)
	seen := map[string]bool{}
	for _, r := range own.Results {
		if !seen[r.RuleID] {
			seen[r.RuleID] = true
			own.Tool.Driver.Rules = append(own.Tool.Driver.Rules, sarif.Rule{ID: r.RuleID})
		}
	}
	sort.Slice(own.Tool.Driver.Rules, func(i, j int) bool {
		return own.Tool.Driver.Rules[i].ID < own.Tool.Driver.Rules[j].ID
	})

	logs := []*sarif.Log{{Schema: sarif.SchemaURI, Version: sarif.Version, Runs: []sarif.Run{own}}}
	logs = append(logs, b.Foreign...)
	return sarif.Merge(logs...)
}

// findings collects one SARIF result per gate denial and per validate or deploy problem.
func findings(b *Bundle) []sarif.Result {
	var out []sarif.Result

	// Gate denials come from the PERMIT when there is one, because it holds every rule.
	// Without a PERMIT the gate result's failed checks stand in for them.
	if b.Permit != nil {
		for _, c := range b.Permit.Checks {
			if c.Outcome == policy.Fail {
				out = append(out, finding("glassops.gate."+c.Rule, c.Rule, c.Detail))
			}
		}
	} else if g := b.stepOf("gate"); g != nil {
		for _, c := range failingChecks(*g) {
			out = append(out, finding("glassops.gate."+c.Name, c.Name, c.Detail))
		}
	}

	for _, step := range []string{"validate", "deploy"} {
		r := b.stepOf(step)
		if r == nil {
			continue
		}
		for _, c := range failingChecks(*r) {
			rule := "glassops." + step + "." + c.Name
			if strings.HasPrefix(c.Name, "problem.") {
				rule = "glassops." + step + ".problem"
			}
			out = append(out, finding(rule, c.Name, c.Detail))
		}
	}
	return out
}

// finding builds one error-level result, with a location when the detail names a file.
func finding(ruleID, name, detail string) sarif.Result {
	msg := plain(detail)
	if msg == "" {
		msg = name
	} else if !strings.HasPrefix(name, "problem.") {
		msg = name + ": " + msg
	}
	res := sarif.Result{RuleID: ruleID, Level: sarif.LevelError, Message: sarif.Text{Text: msg}}
	if file, line := locate(detail); file != "" {
		loc := sarif.Location{PhysicalLocation: sarif.PhysicalLocation{
			ArtifactLocation: sarif.ArtifactLocation{URI: file},
		}}
		if line > 0 {
			loc.PhysicalLocation.Region = &sarif.Region{StartLine: line}
		}
		res.Locations = []sarif.Location{loc}
	}
	return res
}

var (
	fileRe = regexp.MustCompile(`(?:[\w.\-]+[/\\])*[\w\-]+(?:\.[\w\-]+)*\.(?:cls|trigger|js|ts|html|css|xml|cmp|app|page|component|flow|json|ya?ml)\b`)
	// lineAfterFile matches ":12" right after the file name.
	lineAfterFile = regexp.MustCompile(`^:(\d+)`)
	lineWords     = regexp.MustCompile(`(?i)\(?\bline[ :]+(\d+)\b`)
)

// locate finds a file path and a line number in a problem's text. It returns an empty file
// when the text names none. The line is 0 when none is named.
func locate(problem string) (string, int) {
	loc := fileRe.FindStringIndex(problem)
	if loc == nil {
		return "", 0
	}
	file := strings.ReplaceAll(problem[loc[0]:loc[1]], "\\", "/")
	if m := lineAfterFile.FindStringSubmatch(problem[loc[1]:]); m != nil {
		if n, err := strconv.Atoi(m[1]); err == nil && n > 0 {
			return file, n
		}
	}
	if m := lineWords.FindStringSubmatch(problem); m != nil {
		if n, err := strconv.Atoi(m[1]); err == nil && n > 0 {
			return file, n
		}
	}
	return file, 0
}

// plain masks emails and folds whitespace, for SARIF text.
func plain(s string) string { return strings.Join(strings.Fields(redact.Text(s)), " ") }
