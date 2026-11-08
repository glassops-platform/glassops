//go:build redteam_open

package salesforce

// Red team attacks on the Salesforce deploy that currently succeed. Reproduce with:
//
//	go test -tags redteam_open -run TestRTOpen ./internal/providers/salesforce/

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/glassops-platform/glassops/internal/providers"
)

// Finding: checkPlanFiles checks that each manifest sits under the plan dir and that the plan
// files hash to plan_sha256, but not which file plays which part. The change set's destructive
// list is not hashed, so after gate it can name the deploy manifest itself, and sf is asked to
// delete every component the plan deploys.
func TestRTOpenDeployManifestListedAsDestructive(t *testing.T) {
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
