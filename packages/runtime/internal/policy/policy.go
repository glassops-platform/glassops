// Package policy implements the governance policy engine.
package policy

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/glassops-platform/glassops/packages/runtime/internal/gha"
)

// Config represents the full governance configuration.
type Config struct {
	Governance GovernanceConfig `json:"governance"`
	Runtime    RuntimeConfig    `json:"runtime"`
}

// GovernanceConfig contains governance-specific settings.
type GovernanceConfig struct {
	Enabled         bool            `json:"enabled"`
	FreezeWindows   []FreezeWindow  `json:"freeze_windows,omitempty"`
	PluginWhitelist []string        `json:"plugin_whitelist,omitempty"`
	Analyzer        *AnalyzerConfig `json:"analyzer,omitempty"`
}

// FreezeWindow defines a time window during which deployments are blocked.
type FreezeWindow struct {
	Day   string `json:"day"`
	Start string `json:"start"`
	End   string `json:"end"`
}

// AnalyzerConfig contains static analysis settings.
type AnalyzerConfig struct {
	Enabled           bool     `json:"enabled"`
	SeverityThreshold int      `json:"severity_threshold"`
	Rulesets          []string `json:"rulesets,omitempty"`
	Opinionated       bool     `json:"opinionated"`
}

// RuntimeConfig contains runtime environment settings.
type RuntimeConfig struct {
	CLIVersion  string `json:"cli_version"`
	NodeVersion string `json:"node_version"`
}

// Engine manages policy loading and enforcement.
type Engine struct {
	configPath string
}

// New creates a new policy engine.
func New() *Engine {
	configPathInput := os.Getenv("GLASSOPS_CONFIG_PATH")
	if configPathInput == "" {
		configPathInput = "config/devops-config.json"
	}

	var configPath string
	if filepath.IsAbs(configPathInput) {
		configPath = configPathInput
	} else {
		workspace := os.Getenv("GITHUB_WORKSPACE")
		if workspace == "" {
			workspace = "."
		}
		configPath = filepath.Join(workspace, configPathInput)
	}

	return &Engine{
		configPath: configPath,
	}
}

// Load reads and parses the governance configuration.
func (e *Engine) Load() (*Config, error) {
	if _, err := os.Stat(e.configPath); os.IsNotExist(err) {
		gha.Warning("No devops-config.json found. Using default unsafe policy.")
		return &Config{
			Governance: GovernanceConfig{Enabled: false},
			Runtime: RuntimeConfig{
				CLIVersion:  "latest",
				NodeVersion: "20",
			},
		}, nil
	}

	data, err := os.ReadFile(e.configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config: %w", err)
	}

	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("Invalid Governance Policy: %w", err)
	}

	// Apply defaults
	if config.Runtime.CLIVersion == "" {
		config.Runtime.CLIVersion = "latest"
	}
	if config.Runtime.NodeVersion == "" {
		config.Runtime.NodeVersion = "20"
	}
	if config.Governance.Analyzer != nil {
		if config.Governance.Analyzer.SeverityThreshold == 0 {
			config.Governance.Analyzer.SeverityThreshold = 1
		}
	}

	// Validate freeze windows
	timeRegex := regexp.MustCompile(`^\d{2}:\d{2}$`)
	validDays := map[string]bool{
		"Monday": true, "Tuesday": true, "Wednesday": true,
		"Thursday": true, "Friday": true, "Saturday": true, "Sunday": true,
	}
	for _, fw := range config.Governance.FreezeWindows {
		if !validDays[fw.Day] {
			return nil, fmt.Errorf("invalid freeze window day: %s", fw.Day)
		}
		if !timeRegex.MatchString(fw.Start) || !timeRegex.MatchString(fw.End) {
			return nil, fmt.Errorf("invalid freeze window time format (expected HH:MM): %s-%s", fw.Start, fw.End)
		}
	}

	return &config, nil
}

// CheckFreeze validates that the current time is not within a freeze window.
func (e *Engine) CheckFreeze(config *Config) error {
	if len(config.Governance.FreezeWindows) == 0 {
		return nil
	}

	now := time.Now().UTC()
	days := []string{"Sunday", "Monday", "Tuesday", "Wednesday", "Thursday", "Friday", "Saturday"}
	currentDay := days[now.Weekday()]
	currentTime := fmt.Sprintf("%02d:%02d", now.Hour(), now.Minute())

	for _, window := range config.Governance.FreezeWindows {
		if window.Day == currentDay &&
			currentTime >= window.Start &&
			currentTime <= window.End {
			return fmt.Errorf("FROZEN: Deployment blocked by governance window (%s %s-%s)",
				window.Day, window.Start, window.End)
		}
	}

	return nil
}

// ValidatePluginWhitelist checks if a plugin is allowed by the whitelist.
func (e *Engine) ValidatePluginWhitelist(config *Config, pluginName string) bool {
	if len(config.Governance.PluginWhitelist) == 0 {
		return true // No whitelist = allow all
	}

	for _, whitelisted := range config.Governance.PluginWhitelist {
		extracted := extractPluginName(whitelisted)
		if pluginName == extracted {
			return true
		}
	}
	return false
}

// GetPluginVersionConstraint returns the version constraint for a whitelisted plugin.
func (e *Engine) GetPluginVersionConstraint(config *Config, pluginName string) string {
	if len(config.Governance.PluginWhitelist) == 0 {
		return ""
	}

	for _, whitelisted := range config.Governance.PluginWhitelist {
		extracted := extractPluginName(whitelisted)
		if pluginName == extracted {
			return extractVersionConstraint(whitelisted)
		}
	}
	return ""
}

// extractPluginName extracts the package name from a potentially versioned string.
// Examples: "@scope/package@1.0.0" -> "@scope/package", "package@1.0.0" -> "package"
func extractPluginName(entry string) string {
	if strings.HasPrefix(entry, "@") {
		// Scoped package
		lastAt := strings.LastIndex(entry, "@")
		if lastAt > 0 {
			return entry[:lastAt]
		}
		return entry
	}

	// Regular package
	atIdx := strings.Index(entry, "@")
	if atIdx > 0 {
		return entry[:atIdx]
	}
	return entry
}

// extractVersionConstraint extracts the version from a potentially versioned string.
func extractVersionConstraint(entry string) string {
	if strings.HasPrefix(entry, "@") {
		lastAt := strings.LastIndex(entry, "@")
		if lastAt > 0 {
			return entry[lastAt+1:]
		}
		return ""
	}

	atIdx := strings.Index(entry, "@")
	if atIdx > 0 {
		return entry[atIdx+1:]
	}
	return ""
}
