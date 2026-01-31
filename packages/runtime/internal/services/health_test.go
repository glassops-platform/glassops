package services

import (
	"testing"
)

func TestHealthCheckResult(t *testing.T) {
	// Test healthy result
	result := HealthCheckResult{
		Healthy: true,
		Version: "2.0.0",
	}

	if !result.Healthy {
		t.Error("expected Healthy to be true")
	}

	if result.Version != "2.0.0" {
		t.Errorf("expected Version '2.0.0', got '%s'", result.Version)
	}

	// Test unhealthy result
	result = HealthCheckResult{
		Healthy: false,
		Error:   "sf not found",
	}

	if result.Healthy {
		t.Error("expected Healthy to be false")
	}

	if result.Error != "sf not found" {
		t.Errorf("expected Error 'sf not found', got '%s'", result.Error)
	}
}
