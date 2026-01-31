package analyzer

import (
	"testing"
)

func TestNewAnalyzer(t *testing.T) {
	a := New()
	if a == nil {
		t.Error("expected New() to return non-nil Analyzer")
	}
}

func TestEnsureCompatibility(t *testing.T) {
	a := New()
	err := a.EnsureCompatibility()
	if err != nil {
		t.Errorf("EnsureCompatibility() should not return error, got: %v", err)
	}
}

func TestParseOutputEmpty(t *testing.T) {
	a := New()
	result := a.parseOutput("", 0)

	if len(result.Violations) != 0 {
		t.Errorf("expected 0 violations for empty output, got %d", len(result.Violations))
	}

	if result.ExitCode != 0 {
		t.Errorf("expected exit code 0, got %d", result.ExitCode)
	}
}

func TestParseOutputNoJSON(t *testing.T) {
	a := New()
	result := a.parseOutput("No violations found.", 0)

	if len(result.Violations) != 0 {
		t.Errorf("expected 0 violations for non-JSON output, got %d", len(result.Violations))
	}
}

func TestParseOutputValidJSON(t *testing.T) {
	a := New()

	jsonOutput := `[
		{
			"fileName": "src/classes/MyClass.cls",
			"violations": [
				{
					"ruleName": "ApexDoc",
					"message": "Missing ApexDoc comment",
					"severity": 2,
					"line": 10
				},
				{
					"ruleName": "NamingConvention",
					"message": "Method name should be camelCase",
					"severity": 1,
					"line": 25
				}
			]
		}
	]`

	result := a.parseOutput(jsonOutput, 1)

	if len(result.Violations) != 2 {
		t.Fatalf("expected 2 violations, got %d", len(result.Violations))
	}

	if result.ExitCode != 1 {
		t.Errorf("expected exit code 1, got %d", result.ExitCode)
	}

	// Check first violation
	v := result.Violations[0]
	if v.Rule != "ApexDoc" {
		t.Errorf("expected rule 'ApexDoc', got '%s'", v.Rule)
	}
	if v.Severity != 2 {
		t.Errorf("expected severity 2, got %d", v.Severity)
	}
	if v.File != "src/classes/MyClass.cls" {
		t.Errorf("expected file 'src/classes/MyClass.cls', got '%s'", v.File)
	}
	if v.Line != 10 {
		t.Errorf("expected line 10, got %d", v.Line)
	}

	// Check second violation
	v = result.Violations[1]
	if v.Rule != "NamingConvention" {
		t.Errorf("expected rule 'NamingConvention', got '%s'", v.Rule)
	}
	if v.Severity != 1 {
		t.Errorf("expected severity 1, got %d", v.Severity)
	}
}

func TestParseOutputWithClutter(t *testing.T) {
	a := New()

	// JSON with leading/trailing non-JSON content
	jsonOutput := `Some preamble text...
[{"fileName": "test.cls", "violations": [{"ruleName": "Test", "message": "Test msg", "severity": 1, "line": 1}]}]
Some trailing text`

	result := a.parseOutput(jsonOutput, 0)

	if len(result.Violations) != 1 {
		t.Fatalf("expected 1 violation, got %d", len(result.Violations))
	}

	if result.Violations[0].Rule != "Test" {
		t.Errorf("expected rule 'Test', got '%s'", result.Violations[0].Rule)
	}
}

func TestParseOutputInvalidJSON(t *testing.T) {
	a := New()
	result := a.parseOutput("[invalid json", 0)

	// Should return empty violations, not panic
	if len(result.Violations) != 0 {
		t.Errorf("expected 0 violations for invalid JSON, got %d", len(result.Violations))
	}
}

func TestParseOutputMultipleFiles(t *testing.T) {
	a := New()

	jsonOutput := `[
		{
			"fileName": "file1.cls",
			"violations": [
				{"ruleName": "Rule1", "message": "Msg1", "severity": 1, "line": 1}
			]
		},
		{
			"fileName": "file2.cls",
			"violations": [
				{"ruleName": "Rule2", "message": "Msg2", "severity": 2, "line": 2},
				{"ruleName": "Rule3", "message": "Msg3", "severity": 3, "line": 3}
			]
		}
	]`

	result := a.parseOutput(jsonOutput, 0)

	if len(result.Violations) != 3 {
		t.Fatalf("expected 3 violations, got %d", len(result.Violations))
	}

	// Check files are properly associated
	if result.Violations[0].File != "file1.cls" {
		t.Errorf("expected first violation file 'file1.cls', got '%s'", result.Violations[0].File)
	}
	if result.Violations[1].File != "file2.cls" {
		t.Errorf("expected second violation file 'file2.cls', got '%s'", result.Violations[1].File)
	}
}
