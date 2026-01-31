package services

import (
	"encoding/json"
	"os/exec"
)

// HealthCheckResult contains CLI health check results.
type HealthCheckResult struct {
	Healthy bool
	Version string
	Error   string
}

// HealthCheck verifies the Salesforce CLI is available.
func HealthCheck() HealthCheckResult {
	cmd := exec.Command("sf", "version", "--json")
	output, err := cmd.Output()
	if err != nil {
		errMsg := err.Error()
		if exitErr, ok := err.(*exec.ExitError); ok {
			errMsg = string(exitErr.Stderr)
		}
		return HealthCheckResult{
			Healthy: false,
			Error:   errMsg,
		}
	}

	var result struct {
		CLIVersion string `json:"cliVersion"`
		Result     struct {
			CLIVersion string `json:"cliVersion"`
		} `json:"result"`
	}

	if err := json.Unmarshal(output, &result); err != nil {
		return HealthCheckResult{
			Healthy: false,
			Error:   "failed to parse version output",
		}
	}

	version := result.CLIVersion
	if version == "" {
		version = result.Result.CLIVersion
	}
	if version == "" {
		version = "unknown"
	}

	return HealthCheckResult{
		Healthy: true,
		Version: version,
	}
}
