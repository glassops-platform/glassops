package salesforce

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/glassops-platform/glassops/internal/core/junit"
	"github.com/glassops-platform/glassops/internal/core/redact"
	"github.com/glassops-platform/glassops/internal/providers"
)

// defaultTestWait is how long sf waits for a test run when the request sets no wait.
const defaultTestWait = 60 * time.Minute

// apexBody is the result sf prints for a finished test run. An unfinished run prints only
// testRunId.
type apexBody struct {
	TestRunID string `json:"testRunId"`
	Summary   *struct {
		Outcome         string  `json:"outcome"`
		Failing         flexInt `json:"failing"`
		Skipped         flexInt `json:"skipped"`
		TestRunID       string  `json:"testRunId"`
		TestsRan        flexInt `json:"testsRan"`
		OrgWideCoverage string  `json:"orgWideCoverage"`
	} `json:"summary"`
	Tests []struct {
		Message    string  `json:"Message"`
		StackTrace string  `json:"StackTrace"`
		MethodName string  `json:"MethodName"`
		Outcome    string  `json:"Outcome"`
		RunTime    float64 `json:"RunTime"`
		FullName   string  `json:"FullName"`
		ApexClass  struct {
			Name string `json:"Name"`
		} `json:"ApexClass"`
	} `json:"tests"`
	Coverage *struct {
		Coverage []struct {
			Name           string  `json:"name"`
			TotalLines     flexInt `json:"totalLines"`
			TotalCovered   flexInt `json:"totalCovered"`
			CoveredPercent float64 `json:"coveredPercent"`
		} `json:"coverage"`
	} `json:"coverage"`
}

// RunTests runs Apex tests with code coverage, or resumes an earlier run. A failing test makes
// sf exit 100 with a full body, so the body is read whatever the exit code. A run that has not
// finished by the end of the wait comes back with Finished false and its run Id.
func (p *Provider) RunTests(ctx context.Context, s providers.Session, req providers.TestRequest) (providers.TestResult, error) {
	sess, ok := asSession(s)
	if !ok {
		return providers.TestResult{}, errors.New("salesforce: tests need a session from this provider")
	}
	if req.TestLevel == "NoTestRun" {
		return providers.TestResult{}, errors.New("salesforce: the test level NoTestRun runs no tests, so a test step cannot use it")
	}
	wait := req.Wait
	if wait <= 0 {
		wait = defaultTestWait
	}
	mins := strconv.Itoa(int(math.Ceil(wait.Minutes())))

	var args []string
	if req.ResumeID != "" {
		args = []string{"apex", "get", "test", "--test-run-id", req.ResumeID}
	} else {
		args = []string{"apex", "run", "test"}
		if req.TestLevel != "" {
			args = append(args, "--test-level", req.TestLevel)
		}
		if len(req.Tests) > 0 {
			args = append(args, "--tests", strings.Join(req.Tests, ","))
		}
	}
	args = append(args, "--code-coverage", "--result-format", "json", "--target-org", sess.alias, "--wait", mins)

	body, err := p.apexRun(ctx, sess, req.Dir, args)
	if err != nil {
		var sfErr *SFError
		if errors.As(err, &sfErr) && strings.Contains(strings.ToLower(sfErr.Message), "no tests found") {
			// An org with no Apex tests is a result, not a tool failure.
			return providers.TestResult{Finished: true, OrgCoverage: -1,
				Facts: map[string]any{"test.none_found": true}}, nil
		}
		return providers.TestResult{}, err
	}
	if body.Summary == nil && body.TestRunID != "" && req.ResumeID == "" {
		// The run was started and has not finished. Watch it once through get test.
		body, err = p.apexRun(ctx, sess, req.Dir, []string{"apex", "get", "test", "--test-run-id", body.TestRunID,
			"--code-coverage", "--result-format", "json", "--target-org", sess.alias, "--wait", mins})
		if err != nil {
			return providers.TestResult{}, err
		}
	}
	return body.toResult(req.ResumeID)
}

// apexRun runs one apex test command. Exit 100 means failing tests and still carries the body.
func (p *Provider) apexRun(ctx context.Context, sess *session, dir string, args []string) (apexBody, error) {
	env, err := p.sf(ctx, sess, dir, args...)
	if err != nil {
		var sfErr *SFError
		if !errors.As(err, &sfErr) || len(env.Result) == 0 {
			return apexBody{}, err
		}
		var probe apexBody
		if json.Unmarshal(env.Result, &probe) != nil || probe.Summary == nil {
			return apexBody{}, err
		}
	}
	var body apexBody
	if err := json.Unmarshal(env.Result, &body); err != nil {
		return apexBody{}, fmt.Errorf("salesforce: read apex test output: %w", err)
	}
	return body, nil
}

// toResult turns the body into a TestResult. A body without a summary is an unfinished run.
func (b apexBody) toResult(resumeID string) (providers.TestResult, error) {
	res := providers.TestResult{OrgCoverage: -1, Facts: map[string]any{}}
	if b.Summary == nil {
		if b.TestRunID == "" && resumeID == "" {
			return providers.TestResult{}, errors.New("salesforce: apex test output has neither a summary nor a run Id")
		}
		res.RunID = b.TestRunID
		if res.RunID == "" {
			res.RunID = resumeID
		}
		res.Facts["salesforce.test_run_id"] = res.RunID
		return res, nil
	}
	sum := b.Summary
	res.RunID = sum.TestRunID
	res.Ran = int(sum.TestsRan)
	res.Finished = sum.Outcome == "Passed" || sum.Outcome == "Failed"
	if pct, ok := parsePercent(sum.OrgWideCoverage); ok {
		res.OrgCoverage = pct
	}

	var suite junit.Suite
	suite.Name = "apex"
	failed := 0
	for _, t := range b.Tests {
		c := junit.Case{Name: t.MethodName, ClassName: t.ApexClass.Name, Time: t.RunTime / 1000}
		switch t.Outcome {
		case "Pass":
		case "Skip":
			c.Skipped = &junit.Skipped{Message: "skipped"}
		default:
			failed++
			msg := redact.Text(strings.TrimSpace(t.Message))
			c.Failure = &junit.Failure{Message: msg, Type: t.Outcome, Text: redact.Text(t.StackTrace)}
			res.Failures = append(res.Failures, strings.TrimSpace(t.FullName+": "+msg))
		}
		suite.Cases = append(suite.Cases, c)
	}
	res.Failed = failed
	if int(sum.Failing) > res.Failed {
		res.Failed = int(sum.Failing)
	}
	res.Passed = res.Finished && sum.Outcome == "Passed" && res.Failed == 0

	if b.Coverage != nil {
		for _, c := range b.Coverage.Coverage {
			pct := int(math.Round(c.CoveredPercent))
			if c.TotalLines > 0 {
				pct = int(math.Floor(float64(c.TotalCovered) * 100 / float64(c.TotalLines)))
			}
			res.Classes = append(res.Classes, providers.ClassCoverage{
				Name: c.Name, Lines: int(c.TotalLines), Covered: int(c.TotalCovered), Percent: pct,
			})
		}
	}

	var suites junit.Suites
	suites.Name = "apex"
	suites.Add(suite)
	if x, err := suites.Marshal(); err == nil {
		res.JUnit = x
	}
	res.Facts["salesforce.test_run_id"] = res.RunID
	res.Facts["salesforce.test_outcome"] = sum.Outcome
	return res, nil
}

// parsePercent reads "100%" or "87" into 100 or 87.
func parsePercent(s string) (int, bool) {
	s = strings.TrimSpace(strings.TrimSuffix(strings.TrimSpace(s), "%"))
	if s == "" {
		return 0, false
	}
	f, err := strconv.ParseFloat(s, 64)
	if err != nil || f < 0 || f > 100 {
		return 0, false
	}
	return int(math.Floor(f)), true
}
