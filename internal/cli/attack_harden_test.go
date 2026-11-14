package cli

// Red team tests added with the fixes for the deploy path findings. Each one drives the real
// gate, deploy and permit verify commands, as attack_test.go does, and passes when the attack
// is refused. The controls show the same run passes without the attack.

import (
	"errors"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/glassops-platform/glassops/internal/core/exitcode"
	"github.com/glassops-platform/glassops/internal/core/permit"
	"github.com/glassops-platform/glassops/internal/core/policy"
	"github.com/glassops-platform/glassops/internal/core/run"
	"github.com/glassops-platform/glassops/internal/providers"
)

// F1: a PERMIT from a gate that did not bind the change set is refused with a clear message.
func TestRTPermitWithoutChangeSetHashRefused(t *testing.T) {
	e := rtNew(t, rtConfig)
	e.gate(t, exitcode.OK, "--target", "qa")
	e.editPermitJSON(t, func(m map[string]any) { delete(m, "change_set_sha256") })
	e.refuse(t, true, "--target", "qa")
	if !strings.Contains(e.refusal(t), "older gate") {
		t.Fatalf("refusal does not say why: %s", e.refusal(t))
	}
	e.verifyRefused(t, "--target", "qa")
}

// F1: every change set field the deploy reads is bound, not only the plan files.
func TestRTChangeSetFieldsAreBound(t *testing.T) {
	cases := []struct {
		name string
		edit func(cs *providers.ChangeSet)
	}{
		{"empty flipped", func(cs *providers.ChangeSet) { cs.Empty = true }},
		{"full flipped", func(cs *providers.ChangeSet) { cs.Full = !cs.Full }},
		{"provider changed", func(cs *providers.ChangeSet) { cs.Provider = "exec" }},
		{"destructive list emptied", func(cs *providers.ChangeSet) { cs.Destructive = []string{} }},
		{"destructive entry renamed", func(cs *providers.ChangeSet) {
			cs.Destructive = []string{".glassops/plan/destructive/destructiveChangesPre.xml"}
		}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			e := rtNew(t, rtConfig)
			rtWritePlan(t, e.harness, ".glassops/plan", rtManifest, map[string]string{
				"destructiveChanges.xml": "<Package><types><members>Bar</members><name>ApexClass</name></types></Package>\n",
			})
			e.gate(t, exitcode.OK, "--target", "qa")
			rtEditPlan(t, e, tc.edit)
			e.refuse(t, true, "--target", "qa")
			e.verifyRefused(t, "--target", "qa")
		})
	}
}

// F2: deploy refuses unless HEAD is the PERMIT's commit and no tracked file is modified. An
// untracked file is not looked at, so the control with one passes.
func TestRTTreeMustBeTheCommit(t *testing.T) {
	t.Run("control: an untracked file", func(t *testing.T) {
		e := rtNew(t, rtConfig)
		e.gate(t, exitcode.OK, "--target", "qa")
		e.write(t, "notes.txt", "untracked\n")
		if code := e.cmd("deploy", "--target", "qa"); code != exitcode.OK {
			t.Fatalf("exit %d. output:\n%s", code, e.out)
		}
	})
	t.Run("a modified tracked file is named, its content never shown", func(t *testing.T) {
		e := rtNew(t, rtConfig)
		e.gate(t, exitcode.OK, "--target", "qa")
		e.write(t, rtSourcePath, "public class Foo { /* rt-secret-content */ }\n")
		if code := e.refuse(t, true, "--target", "qa"); code != exitcode.Error {
			t.Fatalf("exit %d, want 2", code)
		}
		why := e.refusal(t)
		if !strings.HasPrefix(why, "tree: ") || !strings.Contains(why, rtSourcePath) || !strings.Contains(why, rtCommit[:8]) {
			t.Fatalf("refusal = %s", why)
		}
		if strings.Contains(e.out.String(), "rt-secret-content") {
			t.Fatal("the output shows file content")
		}
	})
	t.Run("a staged change", func(t *testing.T) {
		e := rtNew(t, rtConfig)
		e.gate(t, exitcode.OK, "--target", "qa")
		e.write(t, rtSourcePath, "public class Foo { /* staged */ }\n")
		gitIn(t, e.dir, "add", rtSourcePath)
		e.refuse(t, true, "--target", "qa")
	})
	t.Run("a deleted tracked file", func(t *testing.T) {
		e := rtNew(t, rtConfig)
		e.gate(t, exitcode.OK, "--target", "qa")
		if err := os.Remove(e.dir + "/" + rtSourcePath); err != nil {
			t.Fatal(err)
		}
		e.refuse(t, true, "--target", "qa")
	})
	t.Run("HEAD is another commit", func(t *testing.T) {
		e := rtNew(t, rtConfig)
		e.gate(t, exitcode.OK, "--target", "qa")
		e.tools["git rev-parse HEAD"] = rtOtherCommit + "\n"
		e.refuse(t, true, "--target", "qa")
		if why := e.refusal(t); !strings.Contains(why, rtOtherCommit[:8]) || !strings.Contains(why, rtCommit[:8]) {
			t.Fatalf("refusal = %s", why)
		}
	})
	t.Run("git cannot read HEAD", func(t *testing.T) {
		e := rtNew(t, rtConfig)
		e.gate(t, exitcode.OK, "--target", "qa")
		delete(e.tools, "git rev-parse HEAD")
		if code := e.refuse(t, true, "--target", "qa"); code != exitcode.Error {
			t.Fatalf("exit %d, want 2", code)
		}
	})
	t.Run("git status fails, as outside a repository", func(t *testing.T) {
		e := rtNew(t, rtConfig)
		e.gate(t, exitcode.OK, "--target", "qa")
		if err := os.RemoveAll(e.dir + "/.git"); err != nil {
			t.Fatal(err)
		}
		if code := e.refuse(t, true, "--target", "qa"); code != exitcode.Error {
			t.Fatalf("exit %d, want 2", code)
		}
	})
	t.Run("git is missing", func(t *testing.T) {
		e := rtNew(t, rtConfig)
		e.gate(t, exitcode.OK, "--target", "qa")
		e.app.Run = func(string, string, ...string) ([]byte, error) { return nil, errors.New("exec: git: not found") }
		if code := e.refuse(t, true, "--target", "qa"); code != exitcode.Error {
			t.Fatalf("exit %d, want 2", code)
		}
	})
}

// F6: --resume works for the job the last deploy timed out on, under the same PERMIT, and for
// nothing else.
func TestRTResumeOnlyTheTimedOutJob(t *testing.T) {
	const job = "0Af00000000RT05AAA"
	timeOut := func(e *rtEnv) {
		e.fp.deploy = func(_ providers.Session, cs providers.ChangeSet, o providers.DeployOptions) (providers.DeployResult, error) {
			e.deployed = append(e.deployed, cs)
			e.opts = append(e.opts, o)
			return providers.DeployResult{ID: job, TimedOut: true, Facts: map[string]any{}}, nil
		}
	}
	pass := func(e *rtEnv) {
		e.deployed, e.opts = nil, nil
		e.fp.deploy = func(_ providers.Session, cs providers.ChangeSet, o providers.DeployOptions) (providers.DeployResult, error) {
			e.deployed = append(e.deployed, cs)
			e.opts = append(e.opts, o)
			return providers.DeployResult{Passed: true, ID: o.ResumeID, ComponentsDeployed: 1, Facts: map[string]any{}}, nil
		}
	}
	setup := func(t *testing.T) *rtEnv {
		e := rtNew(t, rtConfig)
		e.gate(t, exitcode.OK, "--target", "qa")
		timeOut(e)
		if code := e.cmd("deploy", "--target", "qa"); code != exitcode.Error {
			t.Fatalf("timed out deploy exit %d. output:\n%s", code, e.out)
		}
		pass(e)
		return e
	}
	t.Run("control: the timed out job under the same PERMIT", func(t *testing.T) {
		e := setup(t)
		if code := e.cmd("deploy", "--target", "qa", "--resume", job); code != exitcode.OK {
			t.Fatalf("exit %d. output:\n%s", code, e.out)
		}
		if e.opts[0].ResumeID != job {
			t.Fatalf("resumed %q", e.opts[0].ResumeID)
		}
	})
	t.Run("another job", func(t *testing.T) {
		e := setup(t)
		if code := e.refuse(t, false, "--target", "qa", "--resume", "0Af00000000RT99AAA"); code != exitcode.Error {
			t.Fatalf("exit %d, want 2", code)
		}
	})
	t.Run("under a new PERMIT", func(t *testing.T) {
		e := setup(t)
		e.gate(t, exitcode.OK, "--target", "qa")
		e.refuse(t, false, "--target", "qa", "--resume", job)
	})
	t.Run("the last deploy did not time out", func(t *testing.T) {
		e := rtNew(t, rtConfig)
		e.gate(t, exitcode.OK, "--target", "qa")
		if code := e.cmd("deploy", "--target", "qa"); code != exitcode.OK {
			t.Fatalf("exit %d. output:\n%s", code, e.out)
		}
		last := e.result(t, "deploy").Facts["deploy.id"].(string)
		pass(e)
		e.refuse(t, false, "--target", "qa", "--resume", last)
	})
	t.Run("no earlier deploy", func(t *testing.T) {
		e := rtNew(t, rtConfig)
		e.gate(t, exitcode.OK, "--target", "qa")
		e.refuse(t, false, "--target", "qa", "--resume", job)
	})
}

const rtWindowToOne = `version: 1
targets:
  qa:
    id: 00D000000000002AAA
    tier: qa
    auth: { method: sfdx-url, url: $SF_QA_AUTH_URL }
policy:
  windows:
    qa:
      - { days: [mon], from: "09:00", to: "13:00", tz: UTC }
`

// F7: a PERMIT issued at 12:50 in a window ending 13:00 expires at 13:00, signed, so deploy at
// 13:01 is refused and deploy at 12:59 is not.
func TestRTPermitExpiresWhenTheWindowCloses(t *testing.T) {
	at := func(h, m int) func() time.Time {
		return func() time.Time { return time.Date(2026, 10, 26, h, m, 0, 0, time.UTC) }
	}
	e := rtNew(t, rtWindowToOne)
	rtUseRules(t, policy.Default)
	e.app.Now = at(12, 50)
	e.gate(t, exitcode.OK, "--target", "qa")
	if got := e.readPermit(t).ExpiresAt; !got.Equal(at(13, 0)()) {
		t.Fatalf("expires_at %s, want 13:00", got.UTC().Format(time.RFC3339))
	}
	if r := e.result(t, "gate"); r.Facts["gate.expires_at_window_end"] != true {
		t.Fatalf("facts = %v", r.Facts)
	}
	e.app.Now = at(13, 1)
	e.refuse(t, true, "--target", "qa")
	e.verifyRefused(t, "--target", "qa")
	e.app.Now = at(12, 59)
	if code := e.cmd("deploy", "--target", "qa"); code != exitcode.OK {
		t.Fatalf("deploy inside the window exit %d. output:\n%s", code, e.out)
	}
}

// F7: a target with no window config keeps the configured TTL.
func TestRTPermitWithoutAWindowKeepsTheTTL(t *testing.T) {
	e := rtNew(t, rtConfig)
	e.gate(t, exitcode.OK, "--target", "qa")
	p := e.readPermit(t)
	if got := p.ExpiresAt.Sub(p.IssuedAt); got != 2*time.Hour {
		t.Fatalf("lifetime %s, want the default TTL", got)
	}
}

// WS23 note: the quick deploy Id comes from a validate result only when that result validated
// the signed plan, and deploy uses the Id the PERMIT signed, never one edited into the result
// afterwards.
func TestRTQuickDeployIDIsBoundToTheSignedPlan(t *testing.T) {
	e := rtNew(t, rtConfig)
	cs, err := providers.ReadChangeSet(e.dir + "/.glassops/plan.json")
	if err != nil {
		t.Fatal(err)
	}
	write := func(id string) {
		r := run.New("validate", nil)
		r.Target = "qa"
		r.Facts["salesforce.deploy_id"] = id
		r.Facts["salesforce.check_only"] = true
		r.Facts["validate.plan_sha256"] = cs.SHA256
		r.Facts["validate.change_set_sha256"] = permit.ChangeSetDigest(cs)
		writeResult(t, e.harness, r)
	}
	write("0Af00000000RT02AAA")
	e.gate(t, exitcode.OK, "--target", "qa")
	if v := e.readPermit(t).Validation; v == nil || v.DeployID != "0Af00000000RT02AAA" {
		t.Fatalf("validation = %+v", v)
	}
	write("0Af00000000RT66AAA")
	if code := e.cmd("deploy", "--target", "qa"); code != exitcode.OK {
		t.Fatalf("exit %d. output:\n%s", code, e.out)
	}
	if got := e.opts[0].ValidatedID; got != "0Af00000000RT02AAA" {
		t.Fatalf("quick deployed %q, not the signed validation", got)
	}
	// A validate result of the same Id for another plan is never bound.
	e.deployed, e.opts = nil, nil
	rtEditPlan(t, e, func(c *providers.ChangeSet) { c.SHA256 = strings.Repeat("b", 64) })
	e.gate(t, exitcode.Error, "--target", "qa")
	if p := e.readPermit(t); p.Allow || p.Validation != nil {
		t.Fatalf("PERMIT = %+v", p)
	}
}

// F3, F12: an outcome gate does not know is a failing check in the result, as it is a deny in
// the PERMIT.
func TestRTUnknownOutcomeIsAFailingCheck(t *testing.T) {
	e := rtNew(t, rtConfig)
	rtUseRules(t, rtFixed(policy.Outcome("FAIL")))
	e.gate(t, exitcode.Fail, "--target", "qa")
	if checkStatus(e.result(t, "gate"), "rt_rule") != run.StatusFail {
		t.Fatalf("checks = %+v", e.result(t, "gate").Checks)
	}
	if e.readPermit(t).Allow {
		t.Fatal("allowed")
	}
}
