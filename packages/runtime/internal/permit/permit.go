package permit

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/glassops-platform/glassops/packages/runtime/internal/gha"
	"github.com/glassops-platform/glassops/packages/runtime/internal/policy"
)

// Permit represents the context handoff contract.
type Permit struct {
	RuntimeID string            `json:"runtime_id"`
	Timestamp string            `json:"timestamp"`
	OrgID     string            `json:"org_id"`
	Policy    *policy.Config    `json:"policy"`
	Context   map[string]string `json:"context"`
	Inputs    map[string]string `json:"inputs"`
}

// Generate creates and writes the glassops-permit.json to the workspace.
func Generate(runtimeID string, orgID string, config *policy.Config, instanceURL string) (string, error) {
	workspace := os.Getenv("GITHUB_WORKSPACE")
	if workspace == "" {
		workspace = "."
	}

	p := Permit{
		RuntimeID: runtimeID,
		Timestamp: time.Now().Format(time.RFC3339),
		OrgID:     orgID,
		Policy:    config,
		Context: map[string]string{
			"actor":      os.Getenv("GITHUB_ACTOR"),
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

	permitPath := filepath.Join(workspace, "glassops-permit.json")
	permitJSON, err := json.MarshalIndent(p, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal permit: %w", err)
	}

	if err := os.WriteFile(permitPath, permitJSON, 0644); err != nil {
		return "", fmt.Errorf("failed to write permit: %w", err)
	}

	return permitPath, nil
}
