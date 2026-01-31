// Package integration provides shared test helpers for integration tests.
package integration

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// DefaultConfig is the default test governance configuration.
var DefaultConfig = map[string]interface{}{
	"governance": map[string]interface{}{
		"enabled": true,
	},
	"runtime": map[string]interface{}{
		"cli_version":  "latest",
		"node_version": "20",
	},
}

// TestEnvironment holds the test environment state.
type TestEnvironment struct {
	WorkspacePath string
	ConfigPath    string
	OriginalEnv   map[string]string
}

// SetupTestWorkspace creates a temporary test workspace with config.
func SetupTestWorkspace(config map[string]interface{}) (*TestEnvironment, error) {
	// Create temp directory
	workspace, err := os.MkdirTemp("", "glassops-test-*")
	if err != nil {
		return nil, err
	}

	// Create config directory
	configDir := filepath.Join(workspace, "config")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		os.RemoveAll(workspace)
		return nil, err
	}

	// Write config file
	configPath := filepath.Join(configDir, "devops-config.json")
	if config == nil {
		config = DefaultConfig
	}
	configData, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		os.RemoveAll(workspace)
		return nil, err
	}
	if err := os.WriteFile(configPath, configData, 0644); err != nil {
		os.RemoveAll(workspace)
		return nil, err
	}

	env := &TestEnvironment{
		WorkspacePath: workspace,
		ConfigPath:    configPath,
		OriginalEnv:   make(map[string]string),
	}

	return env, nil
}

// SetEnvironment sets up the test environment variables.
func (e *TestEnvironment) SetEnvironment(vars map[string]string) {
	// Default environment
	defaults := map[string]string{
		"GITHUB_WORKSPACE":     e.WorkspacePath,
		"GITHUB_ACTOR":         "test-actor",
		"GITHUB_REPOSITORY":    "test-org/test-repo",
		"GITHUB_SHA":           "abc123def456",
		"GITHUB_EVENT_NAME":    "push",
		"GITHUB_HEAD_REF":      "feature-branch",
		"GLASSOPS_CONFIG_PATH": "config/devops-config.json",
	}

	// Merge with overrides
	for k, v := range vars {
		defaults[k] = v
	}

	// Save original values and set new ones
	for k, v := range defaults {
		e.OriginalEnv[k] = os.Getenv(k)
		os.Setenv(k, v)
	}
}

// Cleanup removes the test workspace and restores environment.
func (e *TestEnvironment) Cleanup() {
	// Restore original environment
	for k, v := range e.OriginalEnv {
		if v == "" {
			os.Unsetenv(k)
		} else {
			os.Setenv(k, v)
		}
	}

	// Remove workspace
	if e.WorkspacePath != "" {
		os.RemoveAll(e.WorkspacePath)
	}
}

// WriteConfig updates the test configuration file.
func (e *TestEnvironment) WriteConfig(config map[string]interface{}) error {
	configData, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(e.ConfigPath, configData, 0644)
}

// TestData provides common test data factories.
var TestData = struct {
	FreezeWindows FreezeWindowData
	PluginConfigs PluginConfigData
	TestResults   TestResultsData
	CoverageData  CoverageTestData
}{
	FreezeWindows: FreezeWindowData{
		Weekend: []map[string]string{
			{"day": "Saturday", "start": "00:00", "end": "23:59"},
		},
		Weekday: []map[string]string{
			{"day": "Monday", "start": "09:00", "end": "17:00"},
		},
		Multiple: []map[string]string{
			{"day": "Friday", "start": "17:00", "end": "23:59"},
			{"day": "Saturday", "start": "00:00", "end": "23:59"},
		},
	},
	PluginConfigs: PluginConfigData{
		Whitelist:   []string{"sfdx-hardis@^4.0.0", "@salesforce/plugin-deploy-retrieve"},
		NoWhitelist: []string{},
		Versioned:   []string{"sfdx-hardis@^6.0.0", "sf-metadata-scanner@1.2.3"},
		Scoped:      []string{"@salesforce/plugin-deploy-retrieve@latest"},
	},
	TestResults: TestResultsData{
		Valid:     TestResults{Total: 100, Passed: 95, Failed: 5},
		Empty:     TestResults{Total: 0, Passed: 0, Failed: 0},
		AllPassed: TestResults{Total: 50, Passed: 50, Failed: 0},
		AllFailed: TestResults{Total: 20, Passed: 0, Failed: 20},
	},
	CoverageData: CoverageTestData{
		Good:       Coverage{Actual: 92, Required: 80},
		Borderline: Coverage{Actual: 80, Required: 80},
		Failing:    Coverage{Actual: 75, Required: 80},
		Perfect:    Coverage{Actual: 100, Required: 90},
	},
}

// FreezeWindowData contains freeze window test configurations.
type FreezeWindowData struct {
	Weekend  []map[string]string
	Weekday  []map[string]string
	Multiple []map[string]string
}

// PluginConfigData contains plugin whitelist test configurations.
type PluginConfigData struct {
	Whitelist   []string
	NoWhitelist []string
	Versioned   []string
	Scoped      []string
}

// TestResultsData contains test result test data.
type TestResultsData struct {
	Valid     TestResults
	Empty     TestResults
	AllPassed TestResults
	AllFailed TestResults
}

// TestResults represents test execution results.
type TestResults struct {
	Total  int
	Passed int
	Failed int
}

// CoverageTestData contains coverage test data.
type CoverageTestData struct {
	Good       Coverage
	Borderline Coverage
	Failing    Coverage
	Perfect    Coverage
}

// Coverage represents code coverage data.
type Coverage struct {
	Actual   float64
	Required float64
}
