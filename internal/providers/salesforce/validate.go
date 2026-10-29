package salesforce

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/glassops-platform/glassops/internal/providers"
)

// ErrNotCheckOnly means sf reported a deploy that was not check-only. Validate never calls
// deploy start, so this should be impossible, and commands treat it as a tool error.
var ErrNotCheckOnly = errors.New("salesforce: the target reported a deploy that is not check-only")

// validateWaitMinutes is how long sf waits for the validate to finish.
const validateWaitMinutes = "60"

// Validate runs a check-only deploy validate and reads the full result from deploy report.
// The validate envelope alone is not enough: a failed validate carries only the deploy Id.
func (p *Provider) Validate(ctx context.Context, s providers.Session, cs providers.ChangeSet, opts providers.ValidateOptions) (providers.ValidateResult, error) {
	sess, ok := asSession(s)
	if !ok {
		return providers.ValidateResult{}, errors.New("salesforce: validate needs a session from this provider")
	}
	dir, err := filepath.Abs(opts.Dir)
	if err != nil {
		return providers.ValidateResult{}, fmt.Errorf("salesforce: resolve project dir: %w", err)
	}

	args := []string{"project", "deploy", "validate",
		"--manifest", filepath.Join(dir, filepath.FromSlash(cs.Manifest)),
		"--target-org", sess.alias,
		"--wait", validateWaitMinutes,
	}
	for _, d := range cs.Destructive {
		args = append(args, "--post-destructive-changes", filepath.Join(dir, filepath.FromSlash(d)))
	}
	if opts.TestLevel != "" {
		args = append(args, "--test-level", opts.TestLevel)
		if opts.TestLevel == "RunSpecifiedTests" {
			for _, t := range opts.Tests {
				args = append(args, "--tests", t)
			}
		}
	}

	id, err := p.runValidate(ctx, sess, dir, args)
	if err != nil {
		return providers.ValidateResult{}, err
	}

	rep, err := p.sf(ctx, sess, dir, "project", "deploy", "report", "--job-id", id, "--target-org", sess.alias)
	if err != nil {
		return providers.ValidateResult{}, fmt.Errorf("salesforce: read validate result %s: %w", id, err)
	}
	var body deployBody
	if err := json.Unmarshal(rep.Result, &body); err != nil {
		return providers.ValidateResult{}, fmt.Errorf("salesforce: read deploy report: %w", err)
	}
	if body.CheckOnly == nil || !*body.CheckOnly {
		return providers.ValidateResult{}, ErrNotCheckOnly
	}
	return body.toResult(id), nil
}

// runValidate runs the validate command and returns the deploy Id from either shape: the
// result of a pass or the data of a failure. A failure that carries no Id is an error.
func (p *Provider) runValidate(ctx context.Context, sess *session, dir string, args []string) (string, error) {
	env, err := p.sf(ctx, sess, dir, args...)
	if err == nil {
		var body deployBody
		if err := json.Unmarshal(env.Result, &body); err != nil {
			return "", fmt.Errorf("salesforce: read validate output: %w", err)
		}
		if body.CheckOnly == nil || !*body.CheckOnly {
			return "", ErrNotCheckOnly
		}
		if body.ID == "" {
			return "", errors.New("salesforce: validate reported no deploy Id")
		}
		return body.ID, nil
	}
	var sfErr *SFError
	if errors.As(err, &sfErr) {
		var data struct {
			DeployID string `json:"deployId"`
		}
		if len(sfErr.Data) > 0 && json.Unmarshal(sfErr.Data, &data) == nil && data.DeployID != "" {
			return data.DeployID, nil
		}
	}
	return "", err
}

// flexList reads a JSON list or a single object into a slice. The Metadata API returns a
// bare object when there is exactly one item.
type flexList[T any] []T

func (l *flexList[T]) UnmarshalJSON(b []byte) error {
	b = []byte(strings.TrimSpace(string(b)))
	switch {
	case len(b) == 0 || string(b) == "null":
		*l = nil
		return nil
	case b[0] == '[':
		var items []T
		if err := json.Unmarshal(b, &items); err != nil {
			return err
		}
		*l = items
		return nil
	}
	var one T
	if err := json.Unmarshal(b, &one); err != nil {
		return err
	}
	*l = []T{one}
	return nil
}

// flexInt reads a JSON number or a numeric string. The Metadata API uses both.
type flexInt int

func (n *flexInt) UnmarshalJSON(b []byte) error {
	s := strings.Trim(strings.TrimSpace(string(b)), `"`)
	if s == "" || s == "null" {
		*n = 0
		return nil
	}
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return fmt.Errorf("not a number: %q", s)
	}
	*n = flexInt(f)
	return nil
}

type componentFailure struct {
	ComponentType string  `json:"componentType"`
	FullName      string  `json:"fullName"`
	Problem       string  `json:"problem"`
	LineNumber    flexInt `json:"lineNumber"`
}

type testFailure struct {
	Name       string `json:"name"`
	MethodName string `json:"methodName"`
	Message    string `json:"message"`
}

type classCoverage struct {
	Name                   string  `json:"name"`
	NumLocations           flexInt `json:"numLocations"`
	NumLocationsNotCovered flexInt `json:"numLocationsNotCovered"`
}

// deployBody is the part of a deploy result glassops reads. It is the same shape in the
// validate output and in deploy report.
type deployBody struct {
	CheckOnly             *bool   `json:"checkOnly"`
	ID                    string  `json:"id"`
	Status                string  `json:"status"`
	NumberComponentsTotal flexInt `json:"numberComponentsTotal"`
	NumberComponentErrors flexInt `json:"numberComponentErrors"`
	NumberTestsTotal      flexInt `json:"numberTestsTotal"`
	NumberTestErrors      flexInt `json:"numberTestErrors"`
	Details               struct {
		ComponentFailures flexList[componentFailure] `json:"componentFailures"`
		RunTestResult     struct {
			Failures     flexList[testFailure]   `json:"failures"`
			CodeCoverage flexList[classCoverage] `json:"codeCoverage"`
			NumTestsRun  flexInt                 `json:"numTestsRun"`
			NumFailures  flexInt                 `json:"numFailures"`
		} `json:"runTestResult"`
	} `json:"details"`
}

// toResult builds the neutral result. It never reads createdByName or any user field.
func (b deployBody) toResult(id string) providers.ValidateResult {
	res := providers.ValidateResult{Passed: b.Status == "Succeeded", ID: id}
	for _, f := range b.Details.ComponentFailures {
		line := fmt.Sprintf("%s %s: %s", f.ComponentType, f.FullName, oneLine(f.Problem))
		if f.LineNumber > 0 {
			line += fmt.Sprintf(" (line %d)", f.LineNumber)
		}
		res.Problems = append(res.Problems, line)
	}
	for _, f := range b.Details.RunTestResult.Failures {
		res.Problems = append(res.Problems, fmt.Sprintf("%s.%s: %s", f.Name, f.MethodName, oneLine(f.Message)))
	}
	if !res.Passed && len(res.Problems) == 0 {
		res.Problems = append(res.Problems, "validate did not succeed, status "+b.Status)
	}

	testsTotal := int(b.NumberTestsTotal)
	if n := int(b.Details.RunTestResult.NumTestsRun); n > testsTotal {
		testsTotal = n
	}
	testsFailed := int(b.NumberTestErrors)
	if n := int(b.Details.RunTestResult.NumFailures); n > testsFailed {
		testsFailed = n
	}
	res.Facts = map[string]any{
		"salesforce.deploy_id":         id,
		"salesforce.check_only":        true,
		"salesforce.deploy_status":     b.Status,
		"salesforce.components_total":  int(b.NumberComponentsTotal),
		"salesforce.components_failed": int(b.NumberComponentErrors),
		"salesforce.tests_total":       testsTotal,
		"salesforce.tests_failed":      testsFailed,
	}
	classes := append([]classCoverage(nil), b.Details.RunTestResult.CodeCoverage...)
	sort.Slice(classes, func(i, j int) bool { return classes[i].Name < classes[j].Name })
	for _, c := range classes {
		total := int(c.NumLocations)
		covered := total - int(c.NumLocationsNotCovered)
		pct := 100
		if total > 0 {
			pct = covered * 100 / total
		}
		key := "salesforce.coverage." + c.Name
		res.Facts[key+".lines"] = total
		res.Facts[key+".lines_covered"] = covered
		res.Facts[key+".percent"] = pct
	}
	return res
}

// oneLine keeps a problem on one line so it reads as one entry in a list.
func oneLine(s string) string {
	return strings.Join(strings.Fields(s), " ")
}
