//go:build redteam_open

package cli

// Red team attacks that currently succeed. Each test asserts the behaviour the deploy path
// should have, so it fails until the finding is fixed. They are behind the redteam_open build
// tag so the normal gate stays green. Reproduce with:
//
//	go test -tags redteam_open -run TestRTOpen ./internal/cli/

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/glassops-platform/glassops/internal/core/exitcode"
	"github.com/glassops-platform/glassops/internal/core/policy"
	"github.com/glassops-platform/glassops/internal/core/run"
	"github.com/glassops-platform/glassops/internal/providers"
)

// rtEditPlan rewrites .glassops/plan.json without touching the plan files, so plan_sha256
// and the files still agree.
func rtEditPlan(t *testing.T, e *rtEnv, edit func(cs *providers.ChangeSet)) {
	t.Helper()
	path := filepath.Join(e.dir, ".glassops", "plan.json")
	cs, err := providers.ReadChangeSet(path)
	if err != nil {
		t.Fatal(err)
	}
	edit(&cs)
	if err := providers.WriteChangeSet(path, cs); err != nil {
		t.Fatal(err)
	}
}

// Finding: plan.json fields outside the plan files are not bound by plan_sha256, so the
// destructive list can be changed after gate under a PERMIT that still verifies. Here the
// deploy manifest itself is listed as a post-destructive manifest.
func TestRTOpenDestructiveAddedToPlanAfterGate(t *testing.T) {
	e := rtNew(t, rtConfig)
	e.gate(t, exitcode.OK, "--target", "qa")
	rtEditPlan(t, e, func(cs *providers.ChangeSet) { cs.Destructive = []string{rtPlanRel} })
	e.refuse(t, true, "--target", "qa")
}

// Finding: the source the manifest names is not bound by the PERMIT. Only the manifests are
// hashed, and the commit is read from the CI, not from the tree, so a source file changed
// between gate and deploy is deployed.
func TestRTOpenSourceChangedAfterGate(t *testing.T) {
	e := rtNew(t, rtConfig)
	e.gate(t, exitcode.OK, "--target", "qa")
	e.write(t, rtSourcePath, "public class Foo { /* rt: changed after gate */ }\n")
	e.refuse(t, true, "--target", "qa")
}

const rtSoDConfig = `version: 1
targets:
  qa:
    id: 00D000000000002AAA
    tier: qa
    auth: { method: sfdx-url, url: $SF_QA_AUTH_URL }
policy:
  rules:
    separation_of_duties:
      tiers: [qa]
`

// Finding: gate copies facts from results/plan.json, results/validate.json and the facts in
// plan.json, then sets the ci facts only when it has a value for them. A ci fact it has no
// value for keeps whatever an unsigned file said. On a push with no pull request found, or
// with no API (Bitbucket, GitLab, no token), an injected ci.pr_author or ci.approvals
// satisfies separation of duties for an author deploying their own change.
func TestRTOpenInjectedCIFactsSatisfySoD(t *testing.T) {
	cases := []struct {
		name   string
		cfg    string
		setup  func(t *testing.T, e *rtEnv)
		inject func(t *testing.T, e *rtEnv)
	}{
		{"ci.pr_author in results/plan.json", rtSoDConfig, nil, func(t *testing.T, e *rtEnv) {
			r := run.New("plan", nil)
			r.Facts["ci.pr_author"] = "someone-else"
			writeResult(t, e.harness, r)
		}},
		{"ci.pr_author in results/plan.json, GitHub API on, a push with no pull request", rtSoDConfig,
			func(t *testing.T, e *rtEnv) {
				e.env["GITHUB_SHA"] = unknownSHA
				delete(e.env, "GITHUB_RUN_ID")
				e.useAPI(newGitHubServer(t, defaultRoutes()))
			},
			func(t *testing.T, e *rtEnv) {
				r := run.New("plan", nil)
				r.Facts["ci.pr_author"] = "someone-else"
				writeResult(t, e.harness, r)
			}},
		{"ci.pr_author in the facts of plan.json", rtSoDConfig, nil, func(t *testing.T, e *rtEnv) {
			rtEditPlan(t, e, func(cs *providers.ChangeSet) { cs.Facts["ci.pr_author"] = "someone-else" })
		}},
		{"ci.approvals in results/validate.json", strings.Replace(rtSoDConfig, "tiers: [qa]", "tiers: [qa]\n      approval: { environment: glassops-deploy }", 1),
			nil, func(t *testing.T, e *rtEnv) {
				r := run.New("validate", nil)
				r.Status = run.StatusFail // not bound for a quick deploy, its facts are still read
				r.Facts["ci.approvals"] = []any{map[string]any{
					"mechanism": "environment", "environment": "glassops-deploy", "state": "approved", "approver": "someone-else",
				}}
				writeResult(t, e.harness, r)
			}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			e := rtNew(t, tc.cfg)
			rtUseRules(t, policy.Default)
			if tc.setup != nil {
				tc.setup(t, e)
			}
			// The control: without the injected fact the author deploying their own push is
			// denied.
			e.gate(t, exitcode.Fail, "--target", "qa")
			tc.inject(t, e)
			code := e.cmd("gate", "--target", "qa")
			if p := e.readPermit(t); p.Allow || code == exitcode.OK {
				t.Fatalf("gate exit %d, allow %v: an unsigned file supplied separation of duties evidence. output:\n%s", code, p.Allow, e.out)
			}
		})
	}
}

const rtDestructiveConfig = `version: 1
targets:
  qa:
    id: 00D000000000002AAA
    tier: qa
    auth: { method: sfdx-url, url: $SF_QA_AUTH_URL }
policy:
  rules:
    destructive_needs_label:
      label: destructive-ok
      tiers: [qa]
`

// Finding: change.destructive_count comes from the provider's count in plan.json facts, which
// plan_sha256 does not cover, and those facts are merged over the facts gate derives. A plan
// that sends a destructive manifest can tell gate it has none.
func TestRTOpenDestructiveCountOverriddenByPlanFacts(t *testing.T) {
	for _, key := range []string{"salesforce.destructive_count", "change.destructive_count"} {
		t.Run(key, func(t *testing.T) {
			e := rtNew(t, rtDestructiveConfig)
			rtUseRules(t, policy.Default)
			rtWritePlan(t, e.harness, ".glassops/plan", rtManifest, map[string]string{
				"destructiveChanges.xml": "<Package><types><members>Bar</members><name>ApexClass</name></types></Package>\n",
			})
			// The control: a destructive plan with no pull request label is denied.
			e.gate(t, exitcode.Fail, "--target", "qa")
			rtEditPlan(t, e, func(cs *providers.ChangeSet) { cs.Facts[key] = 0 })
			code := e.cmd("gate", "--target", "qa")
			if p := e.readPermit(t); p.Allow || code == exitcode.OK {
				t.Fatalf("gate exit %d, allow %v for a plan with a destructive manifest and no label. output:\n%s", code, p.Allow, e.out)
			}
		})
	}
}

// Finding: finish prints every check detail to stdout as it is, and GitHub reads any stdout
// line that starts with :: as a workflow command. A target name, or any provider message,
// holding a newline injects commands such as ::warning::, ::add-mask:: or ::stop-commands::.
// The step summary gets the same text and its table is broken the same way.
func TestRTOpenWorkflowCommandInjectionThroughStdout(t *testing.T) {
	e := rtNew(t, rtConfig)
	summary := filepath.Join(e.dir, "rt-summary.md")
	e.env["GITHUB_STEP_SUMMARY"] = summary
	e.gate(t, exitcode.OK, "--target", "qa")
	e.refuse(t, true, "--target", "qa\n::warning title=rt::rt-forged\n# rt-forged heading")
	for _, line := range strings.Split(e.out.String(), "\n") {
		if strings.HasPrefix(strings.TrimSpace(line), "::warning title=rt::") {
			t.Fatalf("an input became a workflow command on its own stdout line: %q", line)
		}
	}
	data, _ := os.ReadFile(summary)
	for _, line := range strings.Split(string(data), "\n") {
		if strings.HasPrefix(line, "# rt-forged") {
			t.Fatalf("an input broke out of the summary table: %q", line)
		}
	}
}

// Finding: --resume takes any job Id. The PERMIT verifies for this plan, then deploy reports
// on a job the PERMIT never named, and when that job succeeded it moves the deployed tag to
// this commit. The next plan then starts from a commit whose changes may never have been
// deployed.
func TestRTOpenResumeOfAJobThePermitNeverNamed(t *testing.T) {
	e := rtNew(t, rtConfig)
	e.gate(t, exitcode.OK, "--target", "qa")
	e.refuse(t, true, "--target", "qa", "--resume", "0Af00000000RT99AAA")
}

const rtWindowConfig = `version: 1
targets:
  qa:
    id: 00D000000000002AAA
    tier: qa
    auth: { method: sfdx-url, url: $SF_QA_AUTH_URL }
policy:
  windows:
    qa:
      - { days: [mon], from: "09:00", to: "12:30", tz: UTC }
`

// Finding: deploy honours a PERMIT for its whole lifetime and never looks at the clock rules
// again. A PERMIT issued inside the deploy window deploys after the window closes, for up to
// policy.permit_ttl (24 hours at most).
func TestRTOpenDeployAfterTheWindowClosed(t *testing.T) {
	e := rtNew(t, rtWindowConfig)
	rtUseRules(t, policy.Default)
	e.gate(t, exitcode.OK, "--target", "qa") // Monday 12:00 UTC, inside the window
	e.app.Now = func() time.Time { return rtGateTime.Add(time.Hour) }
	// The control: gate itself refuses at 13:00.
	saved, err := os.ReadFile(e.permitPath())
	if err != nil {
		t.Fatal(err)
	}
	e.gate(t, exitcode.Fail, "--target", "qa")
	if err := os.WriteFile(e.permitPath(), saved, 0o644); err != nil {
		t.Fatal(err)
	}
	e.refuse(t, true, "--target", "qa")
}

// Finding: a PERMIT from the future verifies. Verify checks expires_at only, so a PERMIT whose
// issued_at is after the deploy clock, from a gate with a skewed clock, is honoured for its
// whole lifetime plus the skew.
func TestRTOpenPermitFromTheFuture(t *testing.T) {
	e := rtNew(t, rtConfig)
	e.app.Now = func() time.Time { return rtGateTime.Add(6 * time.Hour) }
	e.gate(t, exitcode.OK, "--target", "qa")
	e.app.Now = func() time.Time { return rtGateTime }
	e.refuse(t, true, "--target", "qa")
}
