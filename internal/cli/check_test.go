package cli

import (
	"errors"
	"strings"
	"testing"

	"github.com/glassops-platform/glassops/internal/core/exitcode"
	"github.com/glassops-platform/glassops/internal/core/run"
)

const checkConfig = `version: 1
check:
  scripts:
    - name: lint
      run: npm run lint
    - name: unit
      run: npm test
`

// scriptRunner answers "sh -c <script>" from a map of script to error. A missing key passes.
func scriptRunner(h *harness, results map[string]error, ran *[]string) {
	h.app.Run = func(_ string, name string, args ...string) ([]byte, error) {
		if name == "sh" && len(args) == 2 && args[0] == "-c" {
			*ran = append(*ran, args[1])
			return []byte("output of " + args[1] + "\n"), results[args[1]]
		}
		return h.tools.run(name, args...)
	}
}

func TestCheckRunsEveryScriptAndPasses(t *testing.T) {
	h := newHarness(t)
	h.write(t, "sfdx-project.json", "{}")
	h.write(t, "glassops.yaml", checkConfig)
	var ran []string
	scriptRunner(h, nil, &ran)
	if code := h.exec("check"); code != exitcode.OK {
		t.Fatalf("exit %d, output:\n%s", code, h.out)
	}
	if strings.Join(ran, "|") != "npm run lint|npm test" {
		t.Fatalf("ran %v", ran)
	}
	r := h.result(t, "check")
	if checkStatus(r, "script.lint") != run.StatusPass || checkStatus(r, "script.unit") != run.StatusPass {
		t.Fatalf("checks %+v", r.Checks)
	}
	if r.Facts["check.passed"] != float64(2) || r.Facts["check.failed"] != float64(0) {
		t.Fatalf("facts %v", r.Facts)
	}
}

func TestCheckFailingScriptExits1ButRunsTheRest(t *testing.T) {
	h := newHarness(t)
	h.write(t, "sfdx-project.json", "{}")
	h.write(t, "glassops.yaml", checkConfig)
	var ran []string
	scriptRunner(h, map[string]error{"npm run lint": errors.New("exit status 2")}, &ran)
	if code := h.exec("check"); code != exitcode.Fail {
		t.Fatalf("exit %d, output:\n%s", code, h.out)
	}
	if len(ran) != 2 {
		t.Fatalf("ran %v", ran)
	}
	r := h.result(t, "check")
	if checkStatus(r, "script.lint") != run.StatusFail || checkStatus(r, "script.unit") != run.StatusPass {
		t.Fatalf("checks %+v", r.Checks)
	}
	if strings.Contains(r.Checks[0].Detail, "output of") {
		t.Fatalf("script output leaked into the result: %q", r.Checks[0].Detail)
	}
}

func TestCheckStartFailureIsAnError(t *testing.T) {
	h := newHarness(t)
	h.write(t, "sfdx-project.json", "{}")
	h.write(t, "glassops.yaml", checkConfig)
	var ran []string
	scriptRunner(h, map[string]error{"npm test": errors.New("exec: sh: executable file not found in $PATH")}, &ran)
	if code := h.exec("check"); code != exitcode.Error {
		t.Fatalf("exit %d, output:\n%s", code, h.out)
	}
}

func TestCheckScriptFlag(t *testing.T) {
	h := newHarness(t)
	h.write(t, "sfdx-project.json", "{}")
	h.write(t, "glassops.yaml", checkConfig)
	var ran []string
	scriptRunner(h, nil, &ran)
	if code := h.exec("check", "--script", "unit"); code != exitcode.OK || strings.Join(ran, "|") != "npm test" {
		t.Fatalf("exit %d, ran %v", code, ran)
	}
	ran = nil
	if code := h.exec("check", "--script", "nope"); code != exitcode.Error || len(ran) != 0 {
		t.Fatalf("unknown script: exit %d, ran %v", code, ran)
	}
}

func TestCheckWithNoScriptsDoesNothing(t *testing.T) {
	h := newHarness(t)
	h.write(t, "sfdx-project.json", "{}")
	var ran []string
	scriptRunner(h, nil, &ran)
	if code := h.exec("check"); code != exitcode.OK || len(ran) != 0 {
		t.Fatalf("exit %d, ran %v, output:\n%s", code, ran, h.out)
	}
	if r := h.result(t, "check"); r.Status != run.StatusEmpty {
		t.Fatalf("status %s", r.Status)
	}
}
