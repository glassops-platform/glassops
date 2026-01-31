package contract

import "time"

type DeploymentContract struct {
	SchemaVersion string    `json:"schemaVersion"`
	Meta          Meta      `json:"meta"`
	Status        string    `json:"status"` // Succeeded, Failed, Blocked
	Quality       Quality   `json:"quality"`
	Audit         Audit     `json:"audit"`
}

type Meta struct {
	Adapter   string    `json:"adapter"`
	Engine    string    `json:"engine"` // native, hardis, custom
	Timestamp time.Time `json:"timestamp"`
	Trigger   string    `json:"trigger"`
}

type Quality struct {
	Coverage       Coverage       `json:"coverage"`
	Tests          Tests          `json:"tests"`
	StaticAnalysis StaticAnalysis `json:"staticAnalysis,omitempty"` // Added for Phase 1.5
}

type Coverage struct {
	Actual   float64 `json:"actual"`
	Required float64 `json:"required"`
	Met      bool    `json:"met"`
}

type Tests struct {
	Total  int `json:"total"`
	Passed int `json:"passed"`
	Failed int `int:"failed"`
}

// StaticAnalysis represents findings from MegaLinter/Scanner Adapters
type StaticAnalysis struct {
	Tool               string   `json:"tool"`
	Met                bool     `json:"met"`
	CriticalViolations int      `json:"criticalViolations"`
	HighViolations     int      `json:"highViolations"`
	BlockingViolations []string `json:"blockingViolations"`
}

type Audit struct {
	TriggeredBy string `json:"triggeredBy"`
	OrgID       string `json:"orgId"`
	Repository  string `json:"repository"`
	Commit      string `json:"commit"`
}
