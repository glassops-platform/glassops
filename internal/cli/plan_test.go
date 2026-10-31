package cli

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/glassops-platform/glassops/internal/core/exitcode"
	"github.com/glassops-platform/glassops/internal/core/run"
	"github.com/glassops-platform/glassops/internal/providers"
)

const (
	planHead = "1111111111111111111111111111111111111111"
	planBase = "3333333333333333333333333333333333333333"
)

// planGit answers the git calls plan makes, by joined arguments. Anything else fails.
func planGit(answers map[string]string, calls *[]string) Runner {
	return func(_ string, name string, args ...string) ([]byte, error) {
		key := strings.Join(args, " ")
		*calls = append(*calls, key)
		if name == "git" {
			if out, ok := answers[key]; ok {
				return []byte(out + "\n"), nil
			}
		}
		return nil, errors.New("exit status 1")
	}
}

func gitAnswers(extra map[string]string) map[string]string {
	m := map[string]string{
		"rev-parse --is-shallow-repository":        "false",
		"rev-parse --verify --quiet HEAD^{commit}": planHead,
		"rev-parse HEAD":                           planHead,
	}
	for k, v := range extra {
		m[k] = v
	}
	return m
}

// planHarness returns a harness with a fake provider that records the plan request.
func planHarness(t *testing.T, cs providers.ChangeSet, planErr error, answers map[string]string) (*harness, *providers.PlanRequest, *[]string) {
	t.Helper()
	h := newHarness(t)
	var req providers.PlanRequest
	var calls []string
	fp := &fakeProvider{detect: true, plan: func(r providers.PlanRequest) (providers.ChangeSet, error) {
		req = r
		cs.From, cs.To, cs.Full = r.From, r.To, r.Full
		return cs, planErr
	}}
	h.app.Providers = providers.Registry{fp}
	h.app.Run = planGit(answers, &calls)
	return h, &req, &calls
}

func deltaSet() providers.ChangeSet {
	return providers.ChangeSet{
		Schema:      providers.PlanSchema,
		Provider:    "salesforce",
		Dir:         ".glassops/plan",
		Manifest:    ".glassops/plan/package/package.xml",
		Destructive: []string{".glassops/plan/destructiveChanges/destructiveChanges.xml"},
		SHA256:      "abc",
		Facts:       map[string]any{"salesforce.apex_classes": []string{"Greeter"}},
	}
}

func TestPlanExplicitRefs(t *testing.T) {
	h, req, _ := planHarness(t, deltaSet(), nil, gitAnswers(map[string]string{
		"rev-parse --verify --quiet HEAD~1^{commit}": planBase,
	}))
	if code := h.exec("plan", "--from", "HEAD~1", "--to", "HEAD"); code != exitcode.OK {
		t.Fatalf("exit %d, output:\n%s", code, h.out)
	}
	if req.From != planBase || req.To != planHead || req.Full {
		t.Fatalf("request = %+v", req)
	}
	if req.OutDir != filepath.Join(h.dir, ".glassops", "plan") {
		t.Fatalf("OutDir = %s", req.OutDir)
	}
	r := h.result(t, "plan")
	if r.Status != run.StatusPass || r.Provider != "salesforce" {
		t.Fatalf("result = %+v", r)
	}
	if r.Facts["plan.sha256"] != "abc" || r.Facts["plan.from"] != planBase || r.Facts["salesforce.apex_classes"] == nil {
		t.Fatalf("facts = %v", r.Facts)
	}
	cs, err := providers.ReadChangeSet(filepath.Join(h.dir, ".glassops", "plan.json"))
	if err != nil {
		t.Fatal(err)
	}
	if cs.Schema != providers.PlanSchema || cs.Provider != "salesforce" || cs.SHA256 != "abc" {
		t.Fatalf("plan.json = %+v", cs)
	}
}

func TestPlanEmptyIsStatusEmpty(t *testing.T) {
	cs := deltaSet()
	cs.Empty, cs.Destructive = true, []string{}
	h, _, _ := planHarness(t, cs, nil, gitAnswers(map[string]string{"rev-parse --verify --quiet HEAD~1^{commit}": planBase}))
	if code := h.exec("plan", "--from", "HEAD~1"); code != exitcode.OK {
		t.Fatalf("exit %d, want 0 for an empty plan", code)
	}
	if r := h.result(t, "plan"); r.Status != run.StatusEmpty {
		t.Fatalf("status = %s, want empty", r.Status)
	}
	if _, err := os.Stat(filepath.Join(h.dir, ".glassops", "plan.json")); err != nil {
		t.Fatalf("an empty plan still writes plan.json: %v", err)
	}
}

const planTargetConfig = "version: 1\ntargets:\n  qa:\n    id: $QA_ID\n    tier: qa\n    branch: develop\n    auth: { method: sfdx-url, url: $QA_URL }\n"

func TestPlanTargetUsesDeployedTag(t *testing.T) {
	// $QA_ID is unset on purpose: plan needs no Id and must not read it.
	tagged, req, _ := planHarness(t, deltaSet(), nil, gitAnswers(map[string]string{
		"rev-parse --verify --quiet refs/tags/glassops/deployed/qa^{commit}": planBase,
	}))
	tagged.write(t, "glassops.yaml", planTargetConfig)
	if code := tagged.exec("plan", "--target", "qa"); code != exitcode.OK {
		t.Fatalf("exit %d, output:\n%s", code, tagged.out)
	}
	if req.From != planBase || req.Full {
		t.Fatalf("request = %+v, want a delta from the tag", req)
	}
	r := tagged.result(t, "plan")
	if r.Target != "qa" || r.Facts["plan.from_source"] != "tag" || r.Facts["plan.first_deploy"] != false {
		t.Fatalf("result = %+v", r)
	}
}

// First deploy rule: a target with no deployed tag plans the whole source and says why,
// even when a merge base with its branch exists, because that delta would be empty or partial.
func TestPlanTargetWithoutTagPlansFullSource(t *testing.T) {
	cs := deltaSet()
	cs.Destructive = []string{}
	h, req, calls := planHarness(t, cs, nil, gitAnswers(map[string]string{
		"merge-base HEAD origin/develop": planBase,
	}))
	h.write(t, "glassops.yaml", planTargetConfig)
	if code := h.exec("plan", "--target", "qa"); code != exitcode.OK {
		t.Fatalf("exit %d, output:\n%s", code, h.out)
	}
	if !req.Full || req.From != "" || req.To != planHead {
		t.Fatalf("request = %+v, want a full plan at HEAD", req)
	}
	for _, c := range *calls {
		if strings.Contains(c, "merge-base") {
			t.Fatalf("a first deploy still ran git %q", c)
		}
	}
	r := h.result(t, "plan")
	if r.Status != run.StatusPass || r.Facts["plan.first_deploy"] != true || r.Facts["plan.from_source"] != "first-deploy" || r.Facts["plan.full"] != true {
		t.Fatalf("result = %+v", r)
	}
	if checkStatus(r, "refs") != run.StatusSkipped {
		t.Fatalf("checks = %+v", r.Checks)
	}
	if out := h.out.String(); !strings.Contains(out, "notice: plan: target qa has no tag glassops/deployed/qa") || !strings.Contains(out, "whole source") {
		t.Fatalf("no first deploy notice in output:\n%s", out)
	}
}

func TestPlanTargetExplicitFromIgnoresTag(t *testing.T) {
	h, req, calls := planHarness(t, deltaSet(), nil, gitAnswers(map[string]string{
		"rev-parse --verify --quiet HEAD~1^{commit}": planBase,
	}))
	h.write(t, "glassops.yaml", planTargetConfig)
	if code := h.exec("plan", "--target", "qa", "--from", "HEAD~1"); code != exitcode.OK {
		t.Fatalf("exit %d, output:\n%s", code, h.out)
	}
	if req.Full || req.From != planBase {
		t.Fatalf("request = %+v, want the explicit delta", req)
	}
	for _, c := range *calls {
		if strings.Contains(c, "refs/tags") {
			t.Fatalf("an explicit --from still looked up the tag: %q", c)
		}
	}
}

func TestPlanTargetShallowCloneIsExit2(t *testing.T) {
	h, _, _ := planHarness(t, deltaSet(), nil, map[string]string{"rev-parse --is-shallow-repository": "true"})
	h.write(t, "glassops.yaml", planTargetConfig)
	if code := h.exec("plan", "--target", "qa"); code != exitcode.Error {
		t.Fatalf("exit %d, want 2: a shallow clone must not read as a first deploy. output:\n%s", code, h.out)
	}
	if _, err := os.Stat(filepath.Join(h.dir, ".glassops", "plan.json")); err == nil {
		t.Fatal("plan.json written for a shallow clone")
	}
}

func TestPlanPullRequestBase(t *testing.T) {
	h, req, _ := planHarness(t, deltaSet(), nil, gitAnswers(map[string]string{"merge-base HEAD origin/main": planBase}))
	h.env["GITHUB_BASE_REF"] = "main"
	h.env["GITHUB_ACTIONS"] = "true"
	if code := Execute(h.app, []string{"plan", "--dir", h.dir, "--ci", "github"}); code != exitcode.OK {
		t.Fatalf("exit %d, output:\n%s", code, h.out)
	}
	if req.From != planBase {
		t.Fatalf("From = %s, want the merge base with the PR base", req.From)
	}
}

func TestPlanRefErrorsAreExit2(t *testing.T) {
	cases := []struct {
		name    string
		answers map[string]string
		args    []string
		want    string
	}{
		{"auto with nothing to go on", gitAnswers(nil), []string{"plan"}, "Pass --from."},
		{"shallow clone", map[string]string{"rev-parse --is-shallow-repository": "true"}, []string{"plan", "--from", "HEAD~1"}, "fetch-depth: 0"},
		{"explicit ref that is not a commit", gitAnswers(nil), []string{"plan", "--from", "nosuchref"}, "nosuchref"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			h, _, _ := planHarness(t, deltaSet(), nil, tc.answers)
			if code := h.exec(tc.args...); code != exitcode.Error {
				t.Fatalf("exit %d, want 2. output:\n%s", code, h.out)
			}
			r := h.result(t, "plan")
			if r.Status != run.StatusError || len(r.Errors) == 0 || !strings.Contains(r.Errors[0], tc.want) {
				t.Fatalf("result = %+v, want an error containing %q", r, tc.want)
			}
			if _, err := os.Stat(filepath.Join(h.dir, ".glassops", "plan.json")); err == nil {
				t.Fatal("plan.json written after a ref error")
			}
		})
	}
}

func TestPlanUnknownTargetIsExit2(t *testing.T) {
	h, _, _ := planHarness(t, deltaSet(), nil, gitAnswers(nil))
	if code := h.exec("plan", "--target", "nope"); code != exitcode.Error {
		t.Fatalf("exit %d, want 2", code)
	}
}

func TestPlanProviderErrorIsExit2(t *testing.T) {
	h, _, _ := planHarness(t, providers.ChangeSet{}, errors.New("sfdx-git-delta: boom"), gitAnswers(map[string]string{
		"rev-parse --verify --quiet HEAD~1^{commit}": planBase,
	}))
	if code := h.exec("plan", "--from", "HEAD~1"); code != exitcode.Error {
		t.Fatalf("exit %d, want 2", code)
	}
	if r := h.result(t, "plan"); r.Status != run.StatusError || !strings.Contains(strings.Join(r.Errors, " "), "boom") {
		t.Fatalf("result = %+v", r)
	}
}

// Full mode with the defaults --from auto --to auto, no target and no pull request. This is
// how the end-to-end workflow runs it on a push to main.
func TestPlanFullModeNeedsNoRef(t *testing.T) {
	cs := deltaSet()
	cs.Full, cs.Destructive = true, []string{}
	h, req, calls := planHarness(t, cs, nil, map[string]string{
		// A shallow clone and no way to resolve auto: full mode must not care.
		"rev-parse --is-shallow-repository": "true",
		"rev-parse HEAD":                    planHead,
	})
	h.write(t, "glassops.yaml", "version: 1\nsalesforce:\n  delta:\n    enabled: false\n")
	if code := h.exec("plan"); code != exitcode.OK {
		t.Fatalf("exit %d, output:\n%s", code, h.out)
	}
	if !req.Full || req.From != "" || req.To != planHead {
		t.Fatalf("request = %+v", req)
	}
	for _, c := range *calls {
		if strings.Contains(c, "shallow") || strings.Contains(c, "merge-base") {
			t.Fatalf("full mode ran git %q", c)
		}
	}
	r := h.result(t, "plan")
	if r.Status != run.StatusPass || r.Facts["plan.full"] != true {
		t.Fatalf("result = %+v", r)
	}
	if got, err := providers.ReadChangeSet(filepath.Join(h.dir, ".glassops", "plan.json")); err != nil || !got.Full {
		t.Fatalf("plan.json = %+v, %v", got, err)
	}
}

func TestPlanFullModeWithoutGitStillPlans(t *testing.T) {
	cs := deltaSet()
	cs.Full = true
	h, req, _ := planHarness(t, cs, nil, map[string]string{})
	h.write(t, "glassops.yaml", "version: 1\nsalesforce:\n  delta:\n    enabled: false\n")
	if code := h.exec("plan"); code != exitcode.OK {
		t.Fatalf("exit %d, output:\n%s", code, h.out)
	}
	if req.To != "" {
		t.Fatalf("To = %q, want empty when HEAD cannot be read", req.To)
	}
}
