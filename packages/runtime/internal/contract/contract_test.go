package contract

import (
	"encoding/json"
	"testing"
)

func TestNewContract(t *testing.T) {
	c := New()

	if c.SchemaVersion != "1.0" {
		t.Errorf("expected schemaVersion '1.0', got '%s'", c.SchemaVersion)
	}

	if c.Meta.Adapter != "native" {
		t.Errorf("expected adapter 'native', got '%s'", c.Meta.Adapter)
	}

	if c.Meta.Engine != "native" {
		t.Errorf("expected engine 'native', got '%s'", c.Meta.Engine)
	}

	if c.Status != "Succeeded" {
		t.Errorf("expected status 'Succeeded', got '%s'", c.Status)
	}

	if c.Quality.Coverage.Required != 80 {
		t.Errorf("expected coverage.required 80, got %f", c.Quality.Coverage.Required)
	}
}

func TestContractToJSON(t *testing.T) {
	c := New()
	c.Audit.TriggeredBy = "test-user"
	c.Audit.OrgID = "00D000000000001"
	c.Audit.Repository = "test/repo"
	c.Audit.Commit = "abc123"

	jsonBytes, err := c.ToJSON()
	if err != nil {
		t.Fatalf("ToJSON() failed: %v", err)
	}

	// Verify it's valid JSON
	var parsed map[string]interface{}
	if err := json.Unmarshal(jsonBytes, &parsed); err != nil {
		t.Fatalf("produced invalid JSON: %v", err)
	}

	// Verify key fields
	if parsed["schemaVersion"] != "1.0" {
		t.Error("expected schemaVersion in JSON")
	}

	audit, ok := parsed["audit"].(map[string]interface{})
	if !ok {
		t.Fatal("expected audit object in JSON")
	}

	if audit["triggeredBy"] != "test-user" {
		t.Errorf("expected triggeredBy 'test-user', got '%v'", audit["triggeredBy"])
	}
}

func TestContractValidateStatus(t *testing.T) {
	validStatuses := []string{"Succeeded", "Failed", "Blocked"}

	for _, status := range validStatuses {
		c := New()
		c.Status = status
		if err := c.Validate(); err != nil {
			t.Errorf("expected status '%s' to be valid, got error: %v", status, err)
		}
	}

	c := New()
	c.Status = "Invalid"
	if err := c.Validate(); err == nil {
		t.Error("expected error for invalid status")
	}
}

func TestContractValidateEngine(t *testing.T) {
	validEngines := []string{"native", "hardis", "custom"}

	for _, engine := range validEngines {
		c := New()
		c.Meta.Engine = engine
		if err := c.Validate(); err != nil {
			t.Errorf("expected engine '%s' to be valid, got error: %v", engine, err)
		}
	}

	c := New()
	c.Meta.Engine = "invalid"
	if err := c.Validate(); err == nil {
		t.Error("expected error for invalid engine")
	}
}

func TestContractValidateCoverage(t *testing.T) {
	// Valid coverage
	c := New()
	c.Quality.Coverage.Actual = 85.5
	c.Quality.Coverage.Required = 80
	if err := c.Validate(); err != nil {
		t.Errorf("expected valid coverage to pass, got error: %v", err)
	}

	// Invalid actual (negative)
	c = New()
	c.Quality.Coverage.Actual = -1
	if err := c.Validate(); err == nil {
		t.Error("expected error for negative coverage.actual")
	}

	// Invalid actual (over 100)
	c = New()
	c.Quality.Coverage.Actual = 101
	if err := c.Validate(); err == nil {
		t.Error("expected error for coverage.actual > 100")
	}

	// Invalid required
	c = New()
	c.Quality.Coverage.Required = -1
	if err := c.Validate(); err == nil {
		t.Error("expected error for negative coverage.required")
	}
}

func TestContractCoverageMet(t *testing.T) {
	c := New()
	c.Quality.Coverage.Actual = 85
	c.Quality.Coverage.Required = 80
	c.Quality.Coverage.Met = c.Quality.Coverage.Actual >= c.Quality.Coverage.Required

	if !c.Quality.Coverage.Met {
		t.Error("expected coverage.met to be true when actual >= required")
	}

	c.Quality.Coverage.Actual = 75
	c.Quality.Coverage.Met = c.Quality.Coverage.Actual >= c.Quality.Coverage.Required

	if c.Quality.Coverage.Met {
		t.Error("expected coverage.met to be false when actual < required")
	}
}
