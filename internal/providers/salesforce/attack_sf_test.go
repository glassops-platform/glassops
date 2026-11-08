package salesforce

// Red team tests for the plan checks the Salesforce deploy runs before it calls sf. A test here
// passes when the attack is refused. Attacks that currently succeed are in
// attack_sf_open_test.go, behind the redteam_open build tag.

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/glassops-platform/glassops/internal/providers"
)

func TestRTPlanFilesOutsideThePlanDirAreRefused(t *testing.T) {
	cases := []struct {
		name string
		edit func(cs *providers.ChangeSet)
	}{
		{"manifest in the source tree", func(cs *providers.ChangeSet) { cs.Manifest = "force-app/package.xml" }},
		{"manifest climbing out", func(cs *providers.ChangeSet) { cs.Manifest = ".glassops/plan/../../package.xml" }},
		{"manifest in a sibling with the same prefix", func(cs *providers.ChangeSet) { cs.Manifest = ".glassops/planx/package.xml" }},
		{"manifest absolute", func(cs *providers.ChangeSet) { cs.Manifest = "/etc/package.xml" }},
		{"destructive in the source tree", func(cs *providers.ChangeSet) { cs.Destructive = []string{"force-app/destructiveChanges.xml"} }},
		{"destructive climbing out", func(cs *providers.ChangeSet) {
			cs.Destructive = []string{".glassops/plan/destructiveChanges/../../../x.xml"}
		}},
		{"plan dir the project root", func(cs *providers.ChangeSet) { cs.Dir = "." }},
		{"plan dir above the project", func(cs *providers.ChangeSet) { cs.Dir = "../plan" }},
		{"plan dir absolute", func(cs *providers.ChangeSet) { cs.Dir = "/tmp/plan" }},
		{"no plan_sha256", func(cs *providers.ChangeSet) { cs.SHA256 = "" }},
		{"another plan_sha256", func(cs *providers.ChangeSet) { cs.SHA256 = strings.Repeat("0", 64) }},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			p, f, s, dir, cs := deploySetup(t, nil)
			f.replies = map[string]fakeReply{queryPrefix: sandboxOrg(t), startPrefix: {Stdout: fixture(t, "deploy-start-pass.json")}}
			tc.edit(&cs)
			if err := checkPlanFiles(dir, cs); err == nil {
				t.Fatal("checkPlanFiles accepted it")
			}
			if _, err := p.Deploy(context.Background(), s, cs, providers.DeployOptions{Dir: dir}); err == nil {
				t.Fatal("Deploy accepted it")
			}
			for _, c := range f.commands() {
				if strings.HasPrefix(c, startPrefix) || strings.HasPrefix(c, quickPrefix) {
					t.Fatalf("sf ran a deploy: %s", c)
				}
			}
		})
	}
}

func TestRTPlanFilesEditedAfterGateAreRefusedEvenForResume(t *testing.T) {
	p, f, s, dir, cs := deploySetup(t, nil)
	cs.SHA256 = strings.Repeat("0", 64)
	_, err := p.Deploy(context.Background(), s, cs, providers.DeployOptions{Dir: dir, ResumeID: "0Af00000000RT99AAA"})
	if !errors.Is(err, ErrPlanChanged) {
		t.Fatalf("got %v", err)
	}
	if len(f.commands()) != 0 {
		t.Fatalf("sf ran: %v", f.commands())
	}
}
