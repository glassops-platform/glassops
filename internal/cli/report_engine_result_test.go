package cli

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/glassops-platform/glassops/internal/core/report"
	"github.com/glassops-platform/glassops/internal/core/run"
	execprovider "github.com/glassops-platform/glassops/internal/providers/exec"
)

func TestReportAndExecAgreeOnTheEngineResultSchema(t *testing.T) {
	if report.EngineResultSchema != execprovider.ResultSchema {
		t.Fatalf("report skips %q, exec writes %q", report.EngineResultSchema, execprovider.ResultSchema)
	}
}

// An exec command's result file next to the step results is not unreadable evidence.
func TestReportSkipsAnEngineResultFile(t *testing.T) {
	dir := t.TempDir()
	step := run.New("deploy", nil)
	step.AddCheck("deploy", run.StatusPass, "deployed")
	data, err := json.Marshal(step)
	if err != nil {
		t.Fatal(err)
	}
	files := map[string]string{
		"deploy.json":      string(data),
		"exec-deploy.json": `{"schema":"glassops.engine-result/v1","step":"deploy","passed":true,"id":"rel-1"}`,
	}
	for name, body := range files {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	b, err := report.Load(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(b.Problems) != 0 {
		t.Fatalf("problems: %v", b.Problems)
	}
	if len(b.Results) != 1 {
		t.Fatalf("want the one step result, got %d", len(b.Results))
	}
}
