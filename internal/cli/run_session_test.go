package cli

import (
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/glassops-platform/glassops/internal/core/exitcode"
	"github.com/glassops-platform/glassops/internal/core/permit"
	"github.com/glassops-platform/glassops/internal/core/policy"
	"github.com/glassops-platform/glassops/internal/core/run"
	"github.com/glassops-platform/glassops/internal/providers"
)

const sessionCommand = "node scripts/api-reserve.js"

// sessionToken is token shaped, so the step's output must never show it.
// It is built from parts so the source never holds a token shaped literal.
var sessionToken = "00Dxx0000001gPL" + "!" + "AQ4AQFakeTokenValueThatIsLongEnough.abc"

// sessionConfig is deployConfig with a pipeline whose one exec step runs logged in to qa.
const sessionConfig = deployConfig + `pipelines:
  scripts:
    - step: exec
      run: ` + sessionCommand + `
      target: qa
`

// sessionProvider is the deploy harness's fake provider with a session environment.
type sessionProvider struct{ *fakeProvider }

func (sessionProvider) SessionEnv(providers.Session) ([]string, error) {
	return []string{"HOME=/fake-session-home", "SF_TARGET_ORG=glassops-target"}, nil
}

type sessionRun struct {
	calls    int
	env      []string
	commands []string
}

// fakeSessionShell replaces sessionShellRunner. The command prints a token shaped value to
// both streams and exits code. shellRunner must not be used for a step with a target.
func fakeSessionShell(t *testing.T, code int) *sessionRun {
	t.Helper()
	sr := &sessionRun{}
	saved, savedPlain := sessionShellRunner, shellRunner
	t.Cleanup(func() { sessionShellRunner, shellRunner = saved, savedPlain })
	sessionShellRunner = func(_ context.Context, _ string, env []string, command string, stdout, stderr io.Writer) (int, error) {
		sr.calls++
		sr.env = env
		sr.commands = append(sr.commands, command)
		io.WriteString(stdout, "token "+sessionToken+" on stdout\n")
		io.WriteString(stderr, "token "+sessionToken+" on stderr")
		return code, nil
	}
	shellRunner = func(context.Context, string, []string, string, io.Writer, io.Writer) (int, error) {
		t.Fatal("an exec step with a target ran without the session runner")
		return -1, nil
	}
	return sr
}

func newSessionHarness(t *testing.T, cfg string, edit func(*permit.Permit)) *deployHarness {
	t.Helper()
	h := newDeployHarness(t, cfg, edit)
	h.app.Providers = providers.Registry{sessionProvider{h.fp}}
	return h
}

func (h *deployHarness) runScripts() int {
	return Execute(h.app, []string{"run", "scripts", "--dir", h.dir, "--ci", "github"})
}

func bindSessionCommand(p *permit.Permit) {
	p.ExecCommandsSHA256 = permit.ExecCommandDigests([]string{sessionCommand})
}

func TestExecSessionRunsWithABoundPermit(t *testing.T) {
	h := newSessionHarness(t, sessionConfig, bindSessionCommand)
	// An exec step inherits the process environment, so the key must be there to show it is
	// dropped.
	t.Setenv(permit.KeyEnv, deployKey)
	sr := fakeSessionShell(t, 0)
	if code := h.runScripts(); code != exitcode.OK {
		t.Fatalf("exit %d, output:\n%s", code, h.out)
	}
	if sr.calls != 1 || h.authCalls != 1 || !h.fp.allClosed() {
		t.Fatalf("command runs %d, logins %d, sessions closed %v", sr.calls, h.authCalls, h.fp.allClosed())
	}
	if !slices.Contains(sr.env, "SF_TARGET_ORG=glassops-target") || !slices.Contains(sr.env, "HOME=/fake-session-home") {
		t.Fatalf("the command did not get the session environment: %v", sr.env)
	}
	if sr.env[len(sr.env)-1] != "SF_TARGET_ORG=glassops-target" {
		t.Fatal("the session variables must come last, so they win over the job's own")
	}
	for _, kv := range sr.env {
		if strings.HasPrefix(kv, permit.KeyEnv+"=") {
			t.Fatal("a command holding a session got the PERMIT key")
		}
	}
	if out := h.out.String(); strings.Contains(out, sessionToken) || !strings.Contains(out, "token *** on stdout") {
		t.Fatalf("the token was not masked in the output:\n%s", out)
	}
	r := h.result(t, "exec.1")
	if r.Status != run.StatusPass || r.Target != "qa" || r.Facts["exec.session"] != true || r.Facts["exec.permit_id"] != "a1b2c3d4" {
		t.Fatalf("result = %+v", r)
	}
	for _, c := range []string{"permit", "tree", "guard", "command"} {
		if checkStatus(r, c) != run.StatusPass {
			t.Fatalf("check %s = %s, checks %+v", c, checkStatus(r, c), r.Checks)
		}
	}
}

// Every refusal on a protected tier comes before the login, so the command never runs.
func TestExecSessionRefusedBeforeLogin(t *testing.T) {
	cases := []struct {
		name  string
		edit  func(*permit.Permit)
		setup func(h *deployHarness)
		check string
	}{
		{"no command bound", nil, nil, "permit"},
		{"another command bound", func(p *permit.Permit) {
			p.ExecCommandsSHA256 = permit.ExecCommandDigests([]string{"node scripts/other.js"})
		}, nil, "permit"},
		{"PERMIT denies", func(p *permit.Permit) { bindSessionCommand(p); p.Allow = false }, nil, "permit"},
		{"no PERMIT", bindSessionCommand, func(h *deployHarness) {
			_ = os.Remove(filepath.Join(h.dir, ".glassops", "permit.json"))
		}, "permit"},
		{"modified tree", bindSessionCommand, func(h *deployHarness) {
			h.tools[deployStatusCmd] = " M scripts/api-reserve.js\n"
		}, "tree"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			h := newSessionHarness(t, sessionConfig, tc.edit)
			if tc.setup != nil {
				tc.setup(h)
			}
			sr := fakeSessionShell(t, 0)
			if code := h.runScripts(); code != exitcode.Error {
				t.Fatalf("exit %d, want 2. output:\n%s", code, h.out)
			}
			if sr.calls != 0 || h.authCalls != 0 {
				t.Fatalf("command runs %d, logins %d, want neither", sr.calls, h.authCalls)
			}
			if r := h.result(t, "exec.1"); checkStatus(r, tc.check) != run.StatusError {
				t.Fatalf("check %s = %s, checks %+v", tc.check, checkStatus(r, tc.check), r.Checks)
			}
		})
	}
}

const sessionEphemeralConfig = `version: 1
targets:
  qa:
    id: 00D000000000002AAA
    tier: ephemeral
    auth: { method: sfdx-url, url: $SF_QA_AUTH_URL }
pipelines:
  scripts:
    - step: exec
      run: ` + sessionCommand + `
      target: qa
`

// An ephemeral target needs only the guard: no PERMIT, no tree check.
func TestExecSessionEphemeralNeedsOnlyTheGuard(t *testing.T) {
	h := newSessionHarness(t, sessionEphemeralConfig, nil)
	_ = os.Remove(filepath.Join(h.dir, ".glassops", "permit.json"))
	h.tools[deployStatusCmd] = " M anything\n"
	h.fp.describe = reportsAs(providers.KindScratch)
	sr := fakeSessionShell(t, 3)
	if code := h.runScripts(); code != exitcode.Fail {
		t.Fatalf("exit %d, want 1 from the command's exit 3. output:\n%s", code, h.out)
	}
	if sr.calls != 1 || h.authCalls != 1 || !h.fp.allClosed() {
		t.Fatalf("command runs %d, logins %d, closed %v", sr.calls, h.authCalls, h.fp.allClosed())
	}
	r := h.result(t, "exec.1")
	if checkStatus(r, "permit") != run.StatusSkipped || checkStatus(r, "guard") != run.StatusPass || checkStatus(r, "command") != run.StatusFail {
		t.Fatalf("checks = %+v", r.Checks)
	}
}

// The guard still applies on an ephemeral target: a login that lands on another kind of org
// never runs the command.
func TestExecSessionGuardMismatchRunsNothing(t *testing.T) {
	h := newSessionHarness(t, sessionEphemeralConfig, nil)
	h.fp.describe = reportsAs(providers.KindProduction)
	sr := fakeSessionShell(t, 0)
	if code := h.runScripts(); code != exitcode.Error {
		t.Fatalf("exit %d, want 2. output:\n%s", code, h.out)
	}
	if sr.calls != 0 || !h.fp.allClosed() {
		t.Fatalf("command runs %d, closed %v", sr.calls, h.fp.allClosed())
	}
}

// A provider that cannot hand out a session refuses the step before any login.
func TestExecSessionNeedsAProviderThatHandsOutSessions(t *testing.T) {
	h := newDeployHarness(t, sessionEphemeralConfig, nil)
	sr := fakeSessionShell(t, 0)
	if code := h.runScripts(); code != exitcode.Error {
		t.Fatalf("exit %d, want 2. output:\n%s", code, h.out)
	}
	if sr.calls != 0 || h.authCalls != 0 {
		t.Fatalf("command runs %d, logins %d", sr.calls, h.authCalls)
	}
	if r := h.result(t, "exec.1"); !strings.Contains(checkDetail(r, "session"), "cannot hand a session") {
		t.Fatalf("checks = %+v", r.Checks)
	}
}

func TestExecSessionDryRunNamesTheTarget(t *testing.T) {
	h := newSessionHarness(t, sessionConfig, bindSessionCommand)
	sr := fakeSessionShell(t, 0)
	if code := Execute(h.app, []string{"run", "scripts", "--dry-run", "--dir", h.dir, "--ci", "github"}); code != exitcode.OK {
		t.Fatalf("exit %d, output:\n%s", code, h.out)
	}
	if sr.calls != 0 || h.authCalls != 0 || !strings.Contains(h.out.String(), "logged in to qa") {
		t.Fatalf("runs %d, logins %d, output:\n%s", sr.calls, h.authCalls, h.out)
	}
}

func TestGateBindsExecCommands(t *testing.T) {
	installGateFakes(t, fixedRule("plan_not_empty", policy.Pass))
	h, _ := gateHarness(t, gateConfig)
	if code := h.gate("--target", "prod", "--exec-command", "b.sh", "--exec-command", "a.sh", "--exec-command", "b.sh"); code != exitcode.OK {
		t.Fatalf("exit %d, output:\n%s", code, h.out)
	}
	p := readPermit(t, h)
	want := permit.ExecCommandDigests([]string{"a.sh", "b.sh"})
	if !slices.Equal(p.ExecCommandsSHA256, want) {
		t.Fatalf("exec_commands_sha256 = %v, want %v", p.ExecCommandsSHA256, want)
	}
	if strings.Contains(h.out.String(), "a.sh") {
		t.Fatal("the command text reached the output")
	}
}

func TestGateWithoutExecCommandsBindsNone(t *testing.T) {
	installGateFakes(t, fixedRule("plan_not_empty", policy.Pass))
	h, _ := gateHarness(t, gateConfig)
	if code := h.gate("--target", "prod"); code != exitcode.OK {
		t.Fatalf("exit %d, output:\n%s", code, h.out)
	}
	if p := readPermit(t, h); p.ExecCommandsSHA256 != nil {
		t.Fatalf("exec_commands_sha256 = %v, want none", p.ExecCommandsSHA256)
	}
	if r := h.result(t, "gate"); r.Facts["gate.exec_commands_sha256"] != nil {
		t.Fatalf("fact set with no command: %v", r.Facts["gate.exec_commands_sha256"])
	}
}

func TestGateRefusesAnEmptyExecCommand(t *testing.T) {
	installGateFakes(t, fixedRule("plan_not_empty", policy.Pass))
	h, _ := gateHarness(t, gateConfig)
	if code := h.gate("--target", "prod", "--exec-command", "  "); code != exitcode.Error {
		t.Fatalf("exit %d, want 2", code)
	}
	if _, err := os.Stat(filepath.Join(h.dir, ".glassops", "permit.json")); err == nil {
		t.Fatal("a PERMIT was written for an empty --exec-command")
	}
}

// revokingProvider is sessionProvider that also revokes, recording each call.
type revokingProvider struct {
	sessionProvider
	revokes *int
	err     error
}

func (r revokingProvider) RevokeSession(context.Context, providers.Session) error {
	*r.revokes++
	return r.err
}

func TestExecSessionRevokesTheTokenAfterTheCommand(t *testing.T) {
	for _, tc := range []struct {
		name    string
		err     error
		check   run.Status
		revoked bool
	}{
		{"revoked", nil, run.StatusPass, true},
		{"revoke fails, a warning only", errors.New("revoke: example.invalid answered HTTP 400"), run.StatusSkipped, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			h := newSessionHarness(t, sessionConfig, bindSessionCommand)
			revokes := 0
			h.app.Providers = providers.Registry{revokingProvider{sessionProvider{h.fp}, &revokes, tc.err}}
			sr := fakeSessionShell(t, 0)
			if code := h.runScripts(); code != exitcode.OK {
				t.Fatalf("exit %d, output:\n%s", code, h.out)
			}
			if sr.calls != 1 || revokes != 1 || !h.fp.allClosed() {
				t.Fatalf("runs %d, revokes %d, closed %v", sr.calls, revokes, h.fp.allClosed())
			}
			r := h.result(t, "exec.1")
			if r.Status != run.StatusPass || checkStatus(r, "revoke") != tc.check || r.Facts["exec.token_revoked"] != tc.revoked {
				t.Fatalf("result = %+v", r)
			}
			if warned := strings.Contains(h.out.String(), "::warning") && strings.Contains(h.out.String(), "not revoked"); warned == tc.revoked {
				t.Fatalf("warning shown = %v, revoked = %v:\n%s", warned, tc.revoked, h.out)
			}
		})
	}
}

// A refused step logs in to nothing, so there is no token to revoke.
func TestExecSessionRefusalRevokesNothing(t *testing.T) {
	h := newSessionHarness(t, sessionConfig, nil)
	revokes := 0
	h.app.Providers = providers.Registry{revokingProvider{sessionProvider{h.fp}, &revokes, nil}}
	fakeSessionShell(t, 0)
	if code := h.runScripts(); code != exitcode.Error {
		t.Fatalf("exit %d, want 2", code)
	}
	if revokes != 0 {
		t.Fatalf("revokes = %d after a refusal", revokes)
	}
}
