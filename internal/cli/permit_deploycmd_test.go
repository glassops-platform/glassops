package cli

import (
	"strings"
	"testing"

	"github.com/glassops-platform/glassops/internal/core/exitcode"
	"github.com/glassops-platform/glassops/internal/core/run"
)

const boundCommand = `["./other.sh","two words","plain"]`

func gateExec(t *testing.T, p *execProject, extra ...string) {
	t.Helper()
	if code := p.exec(append([]string{"gate", "--target", "qa", "--no-connect"}, extra...)...); code != exitcode.OK {
		t.Fatalf("gate exit %d: %s", code, p.out.String())
	}
}

// permit verify --deploy-command checks the PERMIT against the command about to run, with
// gate's own parsing and hashing.
func TestPermitVerifyDeployCommand(t *testing.T) {
	cases := []struct {
		name      string
		gateArgs  []string
		verifyArg []string
		want      int
		detail    string
	}{
		{"same command", []string{"--deploy-command", boundCommand}, []string{"--deploy-command", boundCommand}, exitcode.OK, ""},
		{"same words with other JSON spacing", []string{"--deploy-command", boundCommand}, []string{"--deploy-command", `[ "./other.sh" , "two words" , "plain" ]`}, exitcode.OK, ""},
		{"another argument", []string{"--deploy-command", boundCommand}, []string{"--deploy-command", `["./other.sh","two words","x"]`}, exitcode.Fail, "differs"},
		{"same words split another way", []string{"--deploy-command", boundCommand}, []string{"--deploy-command", `["./other.sh","two","words","plain"]`}, exitcode.Fail, "differs"},
		{"the PERMIT binds none", nil, []string{"--deploy-command", boundCommand}, exitcode.Fail, "binds no deploy command"},
		{"flag not given, command bound", []string{"--deploy-command", boundCommand}, nil, exitcode.OK, ""},
		{"flag not given, none bound", nil, nil, exitcode.OK, ""},
		{"not JSON", []string{"--deploy-command", boundCommand}, []string{"--deploy-command", "./other.sh plain"}, exitcode.Error, "JSON array"},
		{"empty list", []string{"--deploy-command", boundCommand}, []string{"--deploy-command", "[]"}, exitcode.Error, "names no command"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			p := newExecProject(t, execConfig)
			p.plan(t)
			gateExec(t, p, tc.gateArgs...)
			args := append([]string{"permit", "verify", "--target", "qa"}, tc.verifyArg...)
			code := p.exec(args...)
			if code != tc.want {
				t.Fatalf("exit %d, want %d: %s", code, tc.want, p.out.String())
			}
			if tc.detail != "" && !strings.Contains(p.out.String(), tc.detail) {
				t.Fatalf("output lacks %q: %s", tc.detail, p.out.String())
			}
			if code == exitcode.Fail {
				if r := p.result(t, "permit-verify"); checkStatus(r, "permit") != run.StatusFail {
					t.Fatalf("checks %+v", r.Checks)
				}
			}
		})
	}
}

// Without provider exec the flag is refused the way gate refuses it.
func TestPermitVerifyDeployCommandNeedsExec(t *testing.T) {
	h := permitHarness(t)
	writePermit(t, h, 1, nil)
	if code := h.execGitHub("permit", "verify", "--target", "qa", "--deploy-command", `["./x.sh"]`); code != exitcode.Error {
		t.Fatalf("exit %d: %s", code, h.out)
	}
	if !strings.Contains(h.out.String(), "needs provider exec") {
		t.Fatalf("output: %s", h.out)
	}
}
