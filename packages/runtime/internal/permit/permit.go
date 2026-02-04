package permit

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/glassops-platform/glassops/packages/runtime/internal/gha"
)

// Identity represents a canonical subject across substrates.
type Identity struct {
	Subject    string `json:"subject"`
	Provider   string `json:"provider"`
	ProviderID string `json:"provider_id"`
	Verified   bool   `json:"verified"`
}

// PolicyEvaluation tracks the result of governance checks.
type PolicyEvaluation struct {
	Allowed    bool     `json:"allowed"`
	Evaluated  []string `json:"evaluated"`
	Violations []string `json:"violations,omitempty"`
}

// Permit represents the context handoff contract (Roadmap v1).
type Permit struct {
	Version   string            `json:"version"`
	PermitID  string            `json:"permit_id"`
	Timestamp string            `json:"timestamp"`
	Actor     Identity          `json:"actor"`
	Policies  PolicyEvaluation  `json:"policies"`
	Context   map[string]string `json:"context"`
	Inputs    map[string]string `json:"inputs"`
}

// Generate creates and writes the glassops-permit.json to the workspace.
func Generate(permitID string, actor Identity, evaluation PolicyEvaluation, instanceURL string) (string, error) {
	workspace := os.Getenv("GITHUB_WORKSPACE")
	if workspace == "" {
		workspace = "."
	}

	p := Permit{
		Version:   "1.0",
		PermitID:  permitID,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Actor:     actor,
		Policies:  evaluation,
		Context: map[string]string{
			"repository": os.Getenv("GITHUB_REPOSITORY"),
			"commit":     os.Getenv("GITHUB_SHA"),
			"workspace":  workspace,
		},
		Inputs: map[string]string{
			"instance_url": instanceURL,
			"username":     gha.GetInput("username"),
			"client_id":    gha.GetInput("client_id"),
		},
	}

	// Ensure .glassops directory exists
	glassopsDir := filepath.Join(workspace, ".glassops")
	if err := os.MkdirAll(glassopsDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create .glassops directory: %w", err)
	}

	permitPath := filepath.Join(glassopsDir, "glassops-permit.json")
	permitJSON, err := json.MarshalIndent(p, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal permit: %w", err)
	}

	if err := os.WriteFile(permitPath, permitJSON, 0644); err != nil {
		return "", fmt.Errorf("failed to write permit: %w", err)
	}

	return permitPath, nil
}
