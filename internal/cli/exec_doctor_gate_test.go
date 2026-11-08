package cli

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/glassops-platform/glassops/internal/core/exitcode"
	"github.com/glassops-platform/glassops/internal/core/permit"
	"github.com/glassops-platform/glassops/internal/core/run"
)

func TestDoctorPassesProviderExecFromConfig(t *testing.T) {
	withDescribe := strings.Replace(execConfig, "  deploy: ./deploy.sh\n", "  deploy: ./deploy.sh\n  describe: ./describe.sh\n", 1)
	p := newExecProject(t, withDescribe)
	p.sh.cmds["./describe.sh"] = func(map[string]string) int { return 0 }
	p.setDescribe(`{"id":"acct-prod","kind":"production"}`)
	code := p.exec("doctor", "--auth")
	r := p.result(t, "doctor")
	if checkStatus(r, "provider") != run.StatusPass || checkStatus(r, "auth.prod") != run.StatusPass {
		t.Fatalf("doctor exit %d, checks %+v: %s", code, r.Checks, p.out.String())
	}

	p2 := newExecProject(t, execConfig)
	code = p2.exec("doctor", "--auth", "--target", "qa")
	r = p2.result(t, "doctor")
	if checkStatus(r, "provider") != run.StatusPass || checkStatus(r, "auth.qa") != run.StatusSkipped || code == exitcode.Error {
		t.Fatalf("doctor exit %d, checks %+v: %s", code, r.Checks, p2.out.String())
	}
}

func TestGateExecBelowProdWithoutDescribeSkipsGuard(t *testing.T) {
	p := newExecProject(t, execConfig)
	p.plan(t)
	code := p.exec("gate", "--target", "qa")
	pm, err := permit.Read(filepath.Join(p.dir, ".glassops", "permit.json"))
	if err != nil || code != exitcode.OK || !pm.Allow {
		t.Fatalf("gate exit %d, err %v: %s", code, err, p.out.String())
	}
	p2 := newExecProject(t, execConfig)
	p2.plan(t)
	if code := p2.exec("gate", "--target", "prod"); code == exitcode.OK {
		t.Fatalf("gate allowed prod without describe: %s", p2.out.String())
	}
}
