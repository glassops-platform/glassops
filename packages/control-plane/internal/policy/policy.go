package policy

import (
	"encoding/json"
	"fmt"
	"os"
)

type ProtocolConfig struct {
	Governance Governance `json:"governance"`
}

type Governance struct {
	Enabled         bool            `json:"enabled"`
	MinCoverage     float64         `json:"min_coverage"` // GitHub Floor
	StaticAnalysis  StaticAnalysis  `json:"static_analysis"`
	PluginWhitelist []string        `json:"plugin_whitelist"`
}

type StaticAnalysis struct {
	Enabled bool     `json:"enabled"`
	BlockOn []string `json:"block_on"` // e.g., ["critical", "high"]
}

// ResolvePolicy implements the "Highest Value Wins" merge
// 
func ResolvePolicy(localPath string, githubFloor float64) (ProtocolConfig, error) {
	var config ProtocolConfig

	// 1. Load local devops-config.json
	data, err := os.ReadFile(localPath)
	if err != nil {
		// Fallback to absolute floor if config missing
		return ProtocolConfig{
			Governance: Governance{MinCoverage: githubFloor},
		}, nil
	}

	if err := json.Unmarshal(data, &config); err != nil {
		return config, fmt.Errorf("failed to parse policy: %w", err)
	}

	// 2. Additive Merge: GitHub Floor cannot be lowered
	if config.Governance.MinCoverage < githubFloor {
		config.Governance.MinCoverage = githubFloor
	}

	return config, nil
}
