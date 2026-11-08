package salesforce

// Red team tests for the plan checks the Salesforce deploy runs before it calls sf. A test here
// passes when the attack is refused. An attack that succeeds goes in attack_sf_open_test.go,
// behind the redteam_open build tag, until it is fixed. None is open now.

import (
	"context"
	"errors"
	"path/filepath"
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
		{"destructive is the manifest", func(cs *providers.ChangeSet) { cs.Destructive = []string{cs.Manifest} }},
		{"destructive is the manifest written another way", func(cs *providers.ChangeSet) {
			cs.Destructive = []string{strings.Replace(cs.Manifest, "/package/", "/package/./", 1)}
		}},
		{"destructive not a destructiveChanges file", func(cs *providers.ChangeSet) {
			cs.Destructive = []string{cs.Dir + "/destructiveChanges/package.xml"}
		}},
		{"destructive not xml", func(cs *providers.ChangeSet) {
			cs.Destructive = []string{cs.Dir + "/destructiveChanges/destructiveChanges.txt"}
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

// Finding F1, now refused: checkPlanFiles checked that each manifest sits under the plan dir
// and that the plan files hash to plan_sha256, but not which file plays which part. The change set's destructive
// list is not hashed, so after gate it can name the deploy manifest itself, and sf is asked to
// delete every component the plan deploys.
func TestRTDeployManifestListedAsDestructive(t *testing.T) {
	p, f, s, dir, cs := deploySetup(t, nil)
	f.replies = map[string]fakeReply{queryPrefix: sandboxOrg(t), startPrefix: {Stdout: fixture(t, "deploy-start-pass.json")}}
	cs.Destructive = []string{cs.Manifest}
	manifest := filepath.Join(dir, filepath.FromSlash(cs.Manifest))
	_, err := p.Deploy(context.Background(), s, cs, providers.DeployOptions{Dir: dir})
	for _, c := range f.calls {
		for i := 0; i+1 < len(c.Args); i++ {
			if c.Args[i] == "--post-destructive-changes" && c.Args[i+1] == manifest {
				t.Fatalf("sf was asked to delete the components of the deploy manifest %s (err %v)", c.Args[i+1], err)
			}
		}
	}
	if err == nil {
		t.Fatal("a change set whose destructive list names the deploy manifest was deployed")
	}
}
