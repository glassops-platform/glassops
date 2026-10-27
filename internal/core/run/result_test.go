package run

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/glassops-platform/glassops/internal/core/exitcode"
)

func fixedClock() func() time.Time {
	t := time.Date(2026, 10, 26, 12, 0, 0, 0, time.UTC)
	return func() time.Time { return t }
}

func TestStatusDegradesToWorst(t *testing.T) {
	cases := []struct {
		name   string
		checks []Status
		want   Status
	}{
		{"all pass", []Status{StatusPass, StatusPass}, StatusPass},
		{"skip does not lower pass", []Status{StatusSkipped}, StatusPass},
		{"fail beats pass", []Status{StatusPass, StatusFail}, StatusFail},
		{"error beats fail", []Status{StatusFail, StatusError, StatusPass}, StatusError},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			r := New("doctor", fixedClock())
			for i, s := range tc.checks {
				r.AddCheck("c"+string(rune('a'+i)), s, "")
			}
			if r.Status != tc.want {
				t.Fatalf("status = %s, want %s", r.Status, tc.want)
			}
		})
	}
}

func TestExitCodes(t *testing.T) {
	cases := map[Status]int{
		StatusPass:    exitcode.OK,
		StatusSkipped: exitcode.OK,
		StatusEmpty:   exitcode.OK,
		StatusFail:    exitcode.Fail,
		StatusError:   exitcode.Error,
	}
	for s, want := range cases {
		if got := s.ExitCode(); got != want {
			t.Errorf("%s.ExitCode() = %d, want %d", s, got, want)
		}
	}
}

func TestAddErrorMarksError(t *testing.T) {
	r := New("init", fixedClock())
	r.AddError(errors.New("boom"))
	if r.Status != StatusError || len(r.Errors) != 1 || r.Errors[0] != "boom" {
		t.Fatalf("unexpected result: %+v", r)
	}
}

func TestWriteProducesEnvelope(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "results")
	r := New("doctor", fixedClock())
	r.Provider = "salesforce"
	r.Facts["tools.git"] = "2.45.0"
	path, err := r.Write(dir, fixedClock())
	if err != nil {
		t.Fatal(err)
	}
	if filepath.Base(path) != "doctor.json" {
		t.Fatalf("path = %s", path)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var got map[string]any
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatal(err)
	}
	for _, key := range []string{"schema", "step", "status", "started", "ended", "facts", "artifacts", "errors"} {
		if _, ok := got[key]; !ok {
			t.Errorf("envelope missing %q", key)
		}
	}
	if got["schema"] != Schema {
		t.Errorf("schema = %v", got["schema"])
	}
}
