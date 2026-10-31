package sarif

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// otherToolLog is a synthetic log from another tool, with fields glassops does not model.
const otherToolLog = `{
  "$schema": "https://json.schemastore.org/sarif-2.1.0.json",
  "version": "2.1.0",
  "runs": [{
    "tool": {"driver": {"name": "Code Analyzer", "version": "5.0.0"}},
    "invocations": [{"executionSuccessful": true}],
    "results": [{"ruleId": "pmd:X", "level": "warning", "message": {"text": "m"}}]
  }]
}`

func sample() *Log {
	l := New()
	l.Runs = append(l.Runs, Run{
		Tool: Tool{Driver: Driver{Name: "glassops", Rules: []Rule{{ID: "r1"}}}},
		Results: []Result{{
			RuleID:  "r1",
			Level:   LevelError,
			Message: Text{Text: "bad"},
			Locations: []Location{{PhysicalLocation: PhysicalLocation{
				ArtifactLocation: ArtifactLocation{URI: "force-app/A.cls"},
				Region:           &Region{StartLine: 3},
			}}},
		}},
	})
	return l
}

func TestMarshalIsValid(t *testing.T) {
	data, err := Marshal(sample())
	if err != nil {
		t.Fatal(err)
	}
	if err := Validate(data); err != nil {
		t.Fatalf("own output fails validation: %v\n%s", err, data)
	}
}

func TestEmptyRunStillHasResultsArray(t *testing.T) {
	l := New()
	l.Runs = append(l.Runs, Run{Tool: Tool{Driver: Driver{Name: "glassops"}}})
	data, _ := Marshal(l)
	if !strings.Contains(string(data), `"results": []`) {
		t.Fatalf("results should encode as an empty array:\n%s", data)
	}
	if err := Validate(data); err != nil {
		t.Fatal(err)
	}
}

func TestValidateRejects(t *testing.T) {
	cases := map[string]string{
		"version":   `{"$schema":"s","version":"2.0.0","runs":[]}`,
		"schema":    `{"version":"2.1.0","runs":[]}`,
		"runs":      `{"$schema":"s","version":"2.1.0"}`,
		"driver":    `{"$schema":"s","version":"2.1.0","runs":[{"tool":{"driver":{}},"results":[]}]}`,
		"ruleId":    `{"$schema":"s","version":"2.1.0","runs":[{"tool":{"driver":{"name":"x"}},"results":[{"level":"error","message":{"text":"m"}}]}]}`,
		"level":     `{"$schema":"s","version":"2.1.0","runs":[{"tool":{"driver":{"name":"x"}},"results":[{"ruleId":"r","level":"fatal","message":{"text":"m"}}]}]}`,
		"message":   `{"$schema":"s","version":"2.1.0","runs":[{"tool":{"driver":{"name":"x"}},"results":[{"ruleId":"r","level":"error","message":{}}]}]}`,
		"startLine": `{"$schema":"s","version":"2.1.0","runs":[{"tool":{"driver":{"name":"x"}},"results":[{"ruleId":"r","level":"error","message":{"text":"m"},"locations":[{"physicalLocation":{"artifactLocation":{"uri":"a"},"region":{"startLine":0}}}]}]}]}`,
		"not json":  `{`,
	}
	for name, doc := range cases {
		t.Run(name, func(t *testing.T) {
			if err := Validate([]byte(doc)); err == nil {
				t.Fatalf("%s should be rejected", name)
			}
		})
	}
}

func TestMergeKeepsEachRunsTool(t *testing.T) {
	path := filepath.Join(t.TempDir(), "other.sarif")
	if err := os.WriteFile(path, []byte(otherToolLog), 0o644); err != nil {
		t.Fatal(err)
	}
	other, err := Read(path)
	if err != nil {
		t.Fatal(err)
	}
	merged := Merge(sample(), nil, other)
	if len(merged.Runs) != 2 {
		t.Fatalf("runs = %d, want 2", len(merged.Runs))
	}
	if merged.Runs[0].Tool.Driver.Name != "glassops" || merged.Runs[1].Tool.Driver.Name != "Code Analyzer" {
		t.Fatalf("tools not kept in order: %+v", merged.Runs)
	}
	data, err := Marshal(merged)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), "executionSuccessful") {
		t.Fatalf("a merged run lost fields it had:\n%s", data)
	}
	if err := Validate(data); err != nil {
		t.Fatalf("merged log invalid: %v", err)
	}
}

func TestReadRejectsWrongVersion(t *testing.T) {
	path := filepath.Join(t.TempDir(), "x.sarif")
	_ = os.WriteFile(path, []byte(`{"version":"2.0.0","runs":[]}`), 0o644)
	if _, err := Read(path); err == nil {
		t.Fatal("version 2.0.0 should be refused")
	}
}
