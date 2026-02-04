package integration

import (
	"testing"

	"github.com/glassops-platform/glassops/packages/runtime/internal/contract"
)

func TestContractIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	t.Run("creates valid contract with defaults", func(t *testing.T) {
		c := contract.New()

		if c.SchemaVersion != "1.0" {
			t.Errorf("expected schema version 1.0, got %s", c.SchemaVersion)
		}
		if c.Status != "Succeeded" {
			t.Errorf("expected status Succeeded, got %s", c.Status)
		}
		if c.Meta.Engine != "native" {
			t.Errorf("expected engine native, got %s", c.Meta.Engine)
		}
		if c.Meta.Timestamp == "" {
			t.Error("expected timestamp to be set")
		}
	})

	t.Run("validates contract status", func(t *testing.T) {
		testCases := []struct {
			status    string
			expectErr bool
		}{
			{"Succeeded", false},
			{"Failed", false},
			{"Blocked", false},
			{"Invalid", true},
			{"", true},
		}

		for _, tc := range testCases {
			t.Run(tc.status, func(t *testing.T) {
				c := contract.New()
				c.Status = tc.status

				err := c.Validate()
				if tc.expectErr && err == nil {
					t.Errorf("expected error for status %q", tc.status)
				}
				if !tc.expectErr && err != nil {
					t.Errorf("unexpected error for status %q: %v", tc.status, err)
				}
			})
		}
	})

	t.Run("validates engine types", func(t *testing.T) {
		testCases := []struct {
			engine    string
			expectErr bool
		}{
			{"native", false},
			{"hardis", false},
			{"custom", false},
			{"invalid", true},
			{"", true},
		}

		for _, tc := range testCases {
			t.Run(tc.engine, func(t *testing.T) {
				c := contract.New()
				c.Meta.Engine = tc.engine

				err := c.Validate()
				if tc.expectErr && err == nil {
					t.Errorf("expected error for engine %q", tc.engine)
				}
				if !tc.expectErr && err != nil {
					t.Errorf("unexpected error for engine %q: %v", tc.engine, err)
				}
			})
		}
	})

	t.Run("validates coverage bounds", func(t *testing.T) {
		testCases := []struct {
			name      string
			actual    float64
			required  float64
			expectErr bool
		}{
			{"valid coverage", 85, 80, false},
			{"zero coverage", 0, 0, false},
			{"100% coverage", 100, 100, false},
			{"negative actual", -1, 80, true},
			{"over 100 actual", 101, 80, true},
			{"negative required", 80, -1, true},
			{"over 100 required", 80, 101, true},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				c := contract.New()
				c.Quality.Coverage.Actual = tc.actual
				c.Quality.Coverage.Required = tc.required

				err := c.Validate()
				if tc.expectErr && err == nil {
					t.Errorf("expected error for %s", tc.name)
				}
				if !tc.expectErr && err != nil {
					t.Errorf("unexpected error for %s: %v", tc.name, err)
				}
			})
		}
	})

	t.Run("serializes to JSON", func(t *testing.T) {
		c := contract.New()
		c.Status = "Succeeded"
		c.Quality.Coverage.Actual = 92
		c.Quality.Coverage.Required = 80
		c.Quality.Coverage.Met = true
		c.Quality.Tests.Total = 100
		c.Quality.Tests.Passed = 95
		c.Quality.Tests.Failed = 5
		c.Audit.TriggeredBy = "test-user"
		c.Audit.OrgID = "00D123456789012345"
		c.Audit.Repository = "test-org/test-repo"
		c.Audit.Commit = "abc123"

		jsonData, err := c.ToJSON()
		if err != nil {
			t.Fatalf("failed to serialize: %v", err)
		}

		if len(jsonData) == 0 {
			t.Error("expected non-empty JSON output")
		}
	})

	t.Run("contract with test data", func(t *testing.T) {
		// Use test data helpers
		c := contract.New()
		c.Quality.Coverage.Actual = TestData.CoverageData.Good.Actual
		c.Quality.Coverage.Required = TestData.CoverageData.Good.Required
		c.Quality.Coverage.Met = c.Quality.Coverage.Actual >= c.Quality.Coverage.Required

		c.Quality.Tests.Total = TestData.TestResults.Valid.Total
		c.Quality.Tests.Passed = TestData.TestResults.Valid.Passed
		c.Quality.Tests.Failed = TestData.TestResults.Valid.Failed

		err := c.Validate()
		if err != nil {
			t.Errorf("unexpected validation error: %v", err)
		}

		if !c.Quality.Coverage.Met {
			t.Error("expected coverage to be met")
		}
	})

	t.Run("failing coverage scenario", func(t *testing.T) {
		c := contract.New()
		c.Quality.Coverage.Actual = TestData.CoverageData.Failing.Actual
		c.Quality.Coverage.Required = TestData.CoverageData.Failing.Required
		c.Quality.Coverage.Met = c.Quality.Coverage.Actual >= c.Quality.Coverage.Required

		if c.Quality.Coverage.Met {
			t.Error("expected coverage NOT to be met")
		}
	})
}
