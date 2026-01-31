// Package contract defines the deployment contract schema.
package contract

import (
	"encoding/json"
	"time"
)

// DeploymentContract represents the governance output contract.
type DeploymentContract struct {
	SchemaVersion string  `json:"schemaVersion"`
	Meta          Meta    `json:"meta"`
	Status        string  `json:"status"` // "Succeeded", "Failed", "Blocked"
	Quality       Quality `json:"quality"`
	Audit         Audit   `json:"audit"`
}

// Meta contains execution metadata.
type Meta struct {
	Adapter   string `json:"adapter"`
	Engine    string `json:"engine"` // "native", "hardis", "custom"
	Timestamp string `json:"timestamp"`
	Trigger   string `json:"trigger"`
}

// Quality contains code quality metrics.
type Quality struct {
	Coverage Coverage    `json:"coverage"`
	Tests    TestResults `json:"tests"`
}

// Coverage tracks code coverage requirements.
type Coverage struct {
	Actual   float64 `json:"actual"`
	Required float64 `json:"required"`
	Met      bool    `json:"met"`
}

// TestResults tracks test execution results.
type TestResults struct {
	Total  int `json:"total"`
	Passed int `json:"passed"`
	Failed int `json:"failed"`
}

// Audit contains audit trail information.
type Audit struct {
	TriggeredBy string `json:"triggeredBy"`
	OrgID       string `json:"orgId"`
	Repository  string `json:"repository"`
	Commit      string `json:"commit"`
}

// New creates a new deployment contract with defaults.
func New() *DeploymentContract {
	return &DeploymentContract{
		SchemaVersion: "1.0",
		Meta: Meta{
			Adapter:   "native",
			Engine:    "native",
			Timestamp: time.Now().UTC().Format(time.RFC3339),
		},
		Status: "Succeeded",
		Quality: Quality{
			Coverage: Coverage{Required: 80},
			Tests:    TestResults{},
		},
	}
}

// ToJSON serializes the contract to JSON.
func (c *DeploymentContract) ToJSON() ([]byte, error) {
	return json.MarshalIndent(c, "", "  ")
}

// Validate ensures all required fields are present and valid.
func (c *DeploymentContract) Validate() error {
	// Status validation
	validStatuses := map[string]bool{"Succeeded": true, "Failed": true, "Blocked": true}
	if !validStatuses[c.Status] {
		return &ValidationError{Field: "status", Message: "must be Succeeded, Failed, or Blocked"}
	}

	// Engine validation
	validEngines := map[string]bool{"native": true, "hardis": true, "custom": true}
	if !validEngines[c.Meta.Engine] {
		return &ValidationError{Field: "meta.engine", Message: "must be native, hardis, or custom"}
	}

	// Coverage validation
	if c.Quality.Coverage.Actual < 0 || c.Quality.Coverage.Actual > 100 {
		return &ValidationError{Field: "quality.coverage.actual", Message: "must be between 0 and 100"}
	}
	if c.Quality.Coverage.Required < 0 || c.Quality.Coverage.Required > 100 {
		return &ValidationError{Field: "quality.coverage.required", Message: "must be between 0 and 100"}
	}

	return nil
}

// ValidationError represents a contract validation error.
type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return "validation error: " + e.Field + " " + e.Message
}
