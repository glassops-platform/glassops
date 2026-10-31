package cli

import (
	"context"
	"strings"
	"testing"

	"github.com/glassops-platform/glassops/internal/core/config"
	"github.com/glassops-platform/glassops/internal/core/exitcode"
	"github.com/glassops-platform/glassops/internal/core/run"
)

func TestEnvTargetDefaultsToEphemeral(t *testing.T) {
	root := NewRootCommand(newHarness(t).app)
	for _, sub := range []string{"create", "delete", "sweep"} {
		c, _, _ := root.Find([]string{"env", sub})
		if got := c.Flags().Lookup("target").DefValue; got != "ephemeral" {
			t.Fatalf("env %s --target default = %q", sub, got)
		}
	}
}

func TestTestAndRunRefuseMissingArgs(t *testing.T) {
	h := newHarness(t)
	if code := h.exec("test"); code != exitcode.Error {
		t.Fatalf("test without --target exit %d", code)
	}
	if code := h.exec("run"); code != exitcode.Error {
		t.Fatalf("run without a pipeline exit %d", code)
	}
}

func TestAppExecuteRunsInProcess(t *testing.T) {
	h := newHarness(t)
	h.write(t, "sfdx-project.json", "{}")
	if code := h.app.Execute(context.Background(), []string{"doctor", "--dir", h.dir, "--ci", "local"}); code != exitcode.OK {
		t.Fatalf("exit %d, output:\n%s", code, h.out)
	}
	if code := h.app.Execute(context.Background(), []string{"gate", "--target", "prod", "--dir", h.dir, "--ci", "local"}); code != exitcode.Error {
		t.Fatalf("stub exit %d", code)
	}
}

func TestDoctorExtras(t *testing.T) {
	saved := doctorExtras
	t.Cleanup(func() { doctorExtras = saved })
	calls := 0
	doctorExtras = []func(*App, *config.Config) []run.Check{
		func(*App, *config.Config) []run.Check {
			calls++
			return []run.Check{{Name: "extra.a", Status: run.StatusPass}, {Name: "extra.b", Status: run.StatusSkipped}}
		},
		func(*App, *config.Config) []run.Check {
			calls++
			return []run.Check{{Name: "extra.c", Status: run.StatusFail, Detail: "no"}}
		},
	}

	h := newHarness(t)
	h.write(t, "sfdx-project.json", "{}")
	if code := h.exec("doctor"); code != exitcode.Fail {
		t.Fatalf("exit %d, output:\n%s", code, h.out)
	}
	r := h.result(t, "doctor")
	var names []string
	for _, c := range r.Checks {
		if strings.HasPrefix(c.Name, "extra.") {
			names = append(names, c.Name)
		}
	}
	if strings.Join(names, ",") != "extra.a,extra.b,extra.c" || calls != 2 {
		t.Fatalf("extras %v, calls %d", names, calls)
	}
	// The extras come after the tool checks.
	lastTool, firstExtra := -1, -1
	for i, c := range r.Checks {
		if strings.HasPrefix(c.Name, "tool.") {
			lastTool = i
		}
		if c.Name == "extra.a" {
			firstExtra = i
		}
	}
	if lastTool < 0 || firstExtra < lastTool {
		t.Fatalf("extras are not after the tool checks: %+v", r.Checks)
	}

	calls = 0
	bad := newHarness(t)
	bad.write(t, "sfdx-project.json", "{}")
	bad.write(t, "glassops.yaml", "version: 1\nfoo: bar\n")
	bad.exec("doctor")
	if calls != 0 {
		t.Fatalf("extras ran without a loaded config, %d calls", calls)
	}
}
