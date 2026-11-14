package salesforce

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"strings"
	"testing"

	"github.com/glassops-platform/glassops/internal/providers"
)

// The provider must satisfy the interface the core uses.
var _ providers.Provider = (*Provider)(nil)

func TestParseEnvelopeShapes(t *testing.T) {
	cases := []struct {
		file   string
		status int
		name   string
	}{
		{"deploy-validate-pass.json", 0, ""},
		{"deploy-validate-compile-fail.json", 1, "FailedValidationError"},
		{"deploy-report-test-fail.json", 0, ""},
		{"sgd-delta-bad-ref.json", 1, ""},
	}
	for _, c := range cases {
		t.Run(c.file, func(t *testing.T) {
			env, err := parseEnvelope(fixture(t, c.file))
			if err != nil {
				t.Fatal(err)
			}
			if env.Status != c.status || env.Name != c.name {
				t.Fatalf("status=%d name=%q", env.Status, env.Name)
			}
		})
	}
}

func TestParseEnvelopeSkipsLeadingText(t *testing.T) {
	// sf printed a masked secret prompt before the JSON when stdin was piped.
	out := []byte("? OAuth client secret [input is masked]\x1b[G\n{\"status\":0,\"result\":{}}")
	if _, err := parseEnvelope(out); err != nil {
		t.Fatal(err)
	}
	if _, err := parseEnvelope([]byte("no json here")); err == nil {
		t.Fatal("want an error for output with no JSON")
	}
}

func TestSFRunsWithSessionEnvAndJSON(t *testing.T) {
	f := newFakeExec(map[string]fakeReply{
		"sf project deploy validate": {Stdout: fixture(t, "deploy-validate-compile-fail.json"), ExitCode: 1},
	})
	p := newTestProvider(t, f, nil)
	s := newTestSession(t, providers.Target{Name: "ephemeral"})
	env, err := p.sf(context.Background(), s, "", "project", "deploy", "validate", "--target-org", s.alias)
	var sfErr *SFError
	if !errors.As(err, &sfErr) || sfErr.Name != "FailedValidationError" {
		t.Fatalf("err = %v", err)
	}
	var data struct {
		DeployID string `json:"deployId"`
	}
	if err := json.Unmarshal(env.Data, &data); err != nil || data.DeployID == "" {
		t.Fatalf("deployId from data: %q, %v", data.DeployID, err)
	}
	cmds := f.commands()
	if len(cmds) != 1 || !strings.HasSuffix(cmds[0], " --json") {
		t.Fatalf("commands = %v", cmds)
	}
	var home bool
	for _, kv := range f.calls[0].Env {
		if kv == "HOME="+s.home {
			home = true
		}
	}
	if !home {
		t.Fatalf("session HOME not passed: %v", f.calls[0].Env)
	}
}

func TestCanonicalID(t *testing.T) {
	p := New()
	if p.CanonicalID("00D000000000001AAA") != p.CanonicalID("00D000000000001") {
		t.Fatal("18 and 15 character forms differ")
	}
	if p.CanonicalID("00D000000000001") == p.CanonicalID("00d000000000001") {
		t.Fatal("Ids must compare case-sensitively")
	}
}

func TestMergeEnvOverrides(t *testing.T) {
	got := mergeEnv([]string{"HOME=/real", "PATH=/bin"}, []string{"HOME=/tmp/s"})
	if strings.Join(got, ",") != "PATH=/bin,HOME=/tmp/s" {
		t.Fatalf("got %v", got)
	}
}

func TestSessionCloseRemovesHome(t *testing.T) {
	s := newTestSession(t, providers.Target{})
	if err := s.Close(); err != nil {
		t.Fatal(err)
	}
	if err := s.Close(); err != nil {
		t.Fatal("second close:", err)
	}
}

func TestUnsetEnvRemovesKeys(t *testing.T) {
	got := unsetEnv([]string{"A=1", "TOKEN=x", "B=", "TOKEN_X=y"}, []string{"TOKEN", "B"})
	if strings.Join(got, ",") != "A=1,TOKEN_X=y" {
		t.Fatalf("got %v", got)
	}
}

// TestOsExecUnsetIsAbsentNotEmpty runs a real child and checks it does not see the variable at
// all, which a blank value would not satisfy.
func TestOsExecUnsetIsAbsentNotEmpty(t *testing.T) {
	if os.Getenv("GO_WANT_ENV_CHILD") == "1" {
		_, set := os.LookupEnv("GLASSOPS_TEST_UNSET_ME")
		if set {
			os.Stdout.WriteString("present")
		} else {
			os.Stdout.WriteString("absent")
		}
		os.Exit(0)
	}
	t.Setenv("GLASSOPS_TEST_UNSET_ME", "secret")
	run := func(unset []string) string {
		out, err := osExec(context.Background(), Cmd{
			Name:  os.Args[0],
			Args:  []string{"-test.run=TestOsExecUnsetIsAbsentNotEmpty"},
			Env:   []string{"GO_WANT_ENV_CHILD=1"},
			Unset: unset,
		})
		if err != nil {
			t.Fatal(err)
		}
		return string(out.Stdout)
	}
	if got := run(nil); !strings.Contains(got, "present") {
		t.Fatalf("control: the child should see the variable, got %q", got)
	}
	if got := run([]string{"GLASSOPS_TEST_UNSET_ME"}); !strings.Contains(got, "absent") {
		t.Fatalf("the child should not see the variable, got %q", got)
	}
}
