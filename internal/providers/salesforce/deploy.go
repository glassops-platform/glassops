package salesforce

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/glassops-platform/glassops/internal/providers"
)

// ErrCheckOnlyDeploy means sf reported a check-only run where a real deploy was asked for.
// It is a tool error, never a pass.
var ErrCheckOnlyDeploy = errors.New("salesforce: the target reported a check-only run, not a deploy")

// ErrPlanChanged means the files under the plan directory no longer hash to the change set's
// plan_sha256, so the deploy would not send what the PERMIT was issued for.
var ErrPlanChanged = errors.New("salesforce: the plan files do not match the change set's plan_sha256")

// deployWaitMinutes is how long sf waits for a deploy when the caller sets no wait.
const deployWaitMinutes = 60

// cannotQuickDeploy is the sf error name for a validated Id the org no longer accepts: it was
// never validated, was already deployed, or the validation is more than 10 days old.
const cannotQuickDeploy = "CannotQuickDeployError"

// preDestructiveName is the manifest sf runs before the deploy. Any other destructive
// manifest runs after it, which is also what sf does with destructiveChanges.xml.
const preDestructiveName = "destructiveChangesPre.xml"

// deployResultBody is a deploy result body with the fields only a real deploy needs.
type deployResultBody struct {
	deployBody
	Done                     *bool   `json:"done"`
	NumberComponentsDeployed flexInt `json:"numberComponentsDeployed"`
}

// Deploy deploys cs to the session's org. With opts.ValidatedID it quick deploys that
// validation, and a validation the org refuses falls back to a full deploy unless
// opts.QuickOnly is set. With opts.ResumeID it waits on an earlier deploy instead. Success is
// read from the result body only: status Succeeded, not check-only, at least one component
// deployed and no component or test failures.
func (p *Provider) Deploy(ctx context.Context, s providers.Session, cs providers.ChangeSet, opts providers.DeployOptions) (providers.DeployResult, error) {
	sess, ok := asSession(s)
	if !ok {
		return providers.DeployResult{}, errors.New("salesforce: deploy needs a session from this provider")
	}
	dir, err := filepath.Abs(opts.Dir)
	if err != nil {
		return providers.DeployResult{}, fmt.Errorf("salesforce: resolve project dir: %w", err)
	}
	if err := checkPlanFiles(dir, cs); err != nil {
		return providers.DeployResult{}, err
	}
	wait := waitMinutes(opts.Wait)

	if opts.ResumeID != "" {
		// deploy resume has no --target-org. It finds the org in sf's local deploy cache,
		// which lives in the home of the session that started the deploy and is gone in a new
		// run. deploy report takes the org and a wait, and returns the same body.
		env, err := p.sf(ctx, sess, dir, "project", "deploy", "report", "--job-id", opts.ResumeID, "--wait", wait, "--target-org", sess.alias)
		return p.readDeploy(ctx, sess, dir, env, err, false, nil)
	}

	var fallback map[string]any
	if opts.ValidatedID != "" {
		env, err := p.sf(ctx, sess, dir, "project", "deploy", "quick", "--job-id", opts.ValidatedID, "--wait", wait, "--target-org", sess.alias)
		var sfErr *SFError
		if !errors.As(err, &sfErr) || sfErr.Name != cannotQuickDeploy {
			return p.readDeploy(ctx, sess, dir, env, err, true, nil)
		}
		refused := fmt.Sprintf("quick deploy of validation %s refused: %s", opts.ValidatedID, oneLine(sfErr.Message))
		if opts.QuickOnly {
			return providers.DeployResult{
				Problems: []string{refused + " A quick deploy is required, so no full deploy was run."},
				Facts: map[string]any{
					"salesforce.quick":         false,
					"salesforce.quick_refused": true,
				},
			}, nil
		}
		fallback = map[string]any{
			"salesforce.quick_refused":  true,
			"salesforce.quick_fallback": refused,
		}
	}

	args := []string{"project", "deploy", "start",
		"--manifest", filepath.Join(dir, filepath.FromSlash(cs.Manifest)),
		"--target-org", sess.alias,
		"--wait", wait,
	}
	for _, d := range cs.Destructive {
		flag := "--post-destructive-changes"
		if path.Base(d) == preDestructiveName {
			flag = "--pre-destructive-changes"
		}
		args = append(args, flag, filepath.Join(dir, filepath.FromSlash(d)))
	}
	if opts.TestLevel != "" {
		args = append(args, "--test-level", opts.TestLevel)
		if opts.TestLevel == "RunSpecifiedTests" {
			for _, t := range opts.Tests {
				args = append(args, "--tests", t)
			}
		}
	}
	scratch, err := p.isScratch(ctx, sess)
	if err != nil {
		return providers.DeployResult{}, err
	}
	if scratch {
		// A scratch org tracks source, and a tracking conflict stops deploy start. The
		// project is the source of truth for that org, so a conflict is no reason to stop.
		args = append(args, "--ignore-conflicts")
	}
	env, err := p.sf(ctx, sess, dir, args...)
	return p.readDeploy(ctx, sess, dir, env, err, false, fallback)
}

// readDeploy turns the envelope of deploy start, quick or resume into a result. sf puts the
// full body in result on success and on a component or test failure, with a non-zero status
// for the failure. A failure envelope with no body but a deploy Id in data is read again
// with deploy report. Anything else is an error.
func (p *Provider) readDeploy(ctx context.Context, sess *session, dir string, env envelope, runErr error, quick bool, extra map[string]any) (providers.DeployResult, error) {
	body, ok := parseDeployBody(env.Result)
	if !ok {
		id := failureDeployID(runErr)
		if id == "" {
			if runErr != nil {
				return providers.DeployResult{}, runErr
			}
			return providers.DeployResult{}, errors.New("salesforce: deploy reported no result body")
		}
		rep, err := p.sf(ctx, sess, dir, "project", "deploy", "report", "--job-id", id, "--target-org", sess.alias)
		if err != nil {
			return providers.DeployResult{}, fmt.Errorf("salesforce: read deploy result %s: %w", id, err)
		}
		if body, ok = parseDeployBody(rep.Result); !ok {
			return providers.DeployResult{}, fmt.Errorf("salesforce: deploy report for %s has no result body", id)
		}
	}
	if body.CheckOnly != nil && *body.CheckOnly {
		return providers.DeployResult{}, ErrCheckOnlyDeploy
	}
	res := body.toDeployResult(quick)
	for k, v := range extra {
		res.Facts[k] = v
	}
	return res, nil
}

// parseDeployBody reads a deploy body. ok is false when raw holds no body with an Id and a
// status, such as an error envelope.
func parseDeployBody(raw json.RawMessage) (deployResultBody, bool) {
	var b deployResultBody
	if len(raw) == 0 || json.Unmarshal(raw, &b) != nil {
		return b, false
	}
	return b, b.ID != "" && b.Status != ""
}

// failureDeployID returns the deploy Id an sf failure carries in data, or "".
func failureDeployID(err error) string {
	var sfErr *SFError
	if !errors.As(err, &sfErr) || len(sfErr.Data) == 0 {
		return ""
	}
	var data struct {
		DeployID string `json:"deployId"`
		ID       string `json:"id"`
	}
	if json.Unmarshal(sfErr.Data, &data) != nil {
		return ""
	}
	if data.DeployID != "" {
		return data.DeployID
	}
	return data.ID
}

// finished reports whether a deploy status is final.
func finished(status string) bool {
	switch status {
	case "Succeeded", "SucceededPartial", "Failed", "Canceled":
		return true
	}
	return false
}

// toDeployResult builds the neutral result from a deploy body. The problem lines and the test
// and coverage facts come from the same reading validate uses.
func (b deployResultBody) toDeployResult(quick bool) providers.DeployResult {
	v := b.deployBody.toResult(b.ID)
	res := providers.DeployResult{
		ID:                 b.ID,
		Quick:              quick,
		ComponentsDeployed: int(b.NumberComponentsDeployed),
		Facts:              v.Facts,
	}
	res.Facts["salesforce.check_only"] = false
	res.Facts["salesforce.quick"] = quick
	res.Facts["salesforce.components_deployed"] = res.ComponentsDeployed

	if !finished(b.Status) || (b.Done != nil && !*b.Done) {
		res.TimedOut = true
		res.Facts["salesforce.timed_out"] = true
		res.Problems = []string{fmt.Sprintf("deploy %s has not finished, status %s", b.ID, b.Status)}
		return res
	}

	// Read componentFailures, never numberComponentErrors alone: a destructive "not found"
	// failure comes with the count at 0.
	failures := len(b.Details.ComponentFailures) + len(b.Details.RunTestResult.Failures)
	if failures > 0 {
		res.Problems = v.Problems
	}
	testsFailed, _ := res.Facts["salesforce.tests_failed"].(int)
	switch {
	case b.Status == "SucceededPartial":
		res.Problems = append(res.Problems, "deploy only partly succeeded, status SucceededPartial")
	case b.Status != "Succeeded":
		if failures == 0 {
			res.Problems = append(res.Problems, "deploy did not succeed, status "+b.Status)
		}
	case b.CheckOnly == nil:
		res.Problems = append(res.Problems, "deploy result does not say whether it was check-only")
	case res.ComponentsDeployed <= 0:
		res.Problems = append(res.Problems, "deploy reported Succeeded but deployed no components")
	case failures > 0 || int(b.NumberComponentErrors) > 0 || testsFailed > 0:
		// A Succeeded body with failures should not happen. It is read as a failure.
		if failures == 0 {
			res.Problems = append(res.Problems, "deploy reported Succeeded with component or test errors")
		}
	default:
		res.Passed = true
	}
	return res
}

// checkPlanFiles checks that the manifests the deploy sends sit under the plan directory and
// that the files there still hash to the change set's plan_sha256, the value a PERMIT binds.
func checkPlanFiles(dir string, cs providers.ChangeSet) error {
	if cs.Dir == "" || cs.Manifest == "" || cs.SHA256 == "" {
		return errors.New("salesforce: the change set has no plan dir, manifest or plan_sha256")
	}
	planDir := path.Clean(cs.Dir)
	if path.IsAbs(planDir) || filepath.IsAbs(cs.Dir) || planDir == "." || planDir == ".." || strings.HasPrefix(planDir, "../") {
		return fmt.Errorf("salesforce: plan dir %q is not inside the project", cs.Dir)
	}
	for _, f := range append([]string{cs.Manifest}, cs.Destructive...) {
		if !strings.HasPrefix(path.Clean(f), planDir+"/") {
			return fmt.Errorf("salesforce: manifest %q is not inside the plan dir %s", f, cs.Dir)
		}
	}
	sum, err := planDigest(dir, filepath.Join(dir, filepath.FromSlash(planDir)))
	if err != nil {
		return fmt.Errorf("salesforce: hash the plan files: %w", err)
	}
	if sum != cs.SHA256 {
		return ErrPlanChanged
	}
	return nil
}

// isScratch reads whether the session's org is a scratch org, the same way Describe does.
func (p *Provider) isScratch(ctx context.Context, sess *session) (bool, error) {
	env, err := p.sf(ctx, sess, "", "data", "query", "--query", orgQuery, "--target-org", sess.alias)
	if err != nil {
		return false, describeErr("query the organization", err)
	}
	var q struct {
		Records []orgRecord `json:"records"`
	}
	if err := json.Unmarshal(env.Result, &q); err != nil || len(q.Records) != 1 || q.Records[0].IsSandbox == nil {
		return false, errors.New("salesforce: could not read the organization before the deploy")
	}
	rec := q.Records[0]
	return *rec.IsSandbox && rec.TrialExpirationDate != nil && *rec.TrialExpirationDate != "", nil
}

// waitMinutes is d in whole minutes, rounded up and at least 1. Zero means the default.
func waitMinutes(d time.Duration) string {
	if d <= 0 {
		return fmt.Sprint(deployWaitMinutes)
	}
	m := int((d + time.Minute - 1) / time.Minute)
	return fmt.Sprint(max(m, 1))
}
