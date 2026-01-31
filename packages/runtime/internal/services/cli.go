// Package services contains runtime service implementations.
package services

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/glassops-platform/glassops/packages/runtime/internal/gha"
	"github.com/glassops-platform/glassops/packages/runtime/internal/policy"
)

// RuntimeEnvironment handles CLI installation and plugin management.
type RuntimeEnvironment struct {
	platform string
}

// NewRuntimeEnvironment creates a new RuntimeEnvironment.
func NewRuntimeEnvironment() *RuntimeEnvironment {
	return &RuntimeEnvironment{
		platform: os.Getenv("GOOS"),
	}
}

// Install installs the Salesforce CLI if not already present.
func (r *RuntimeEnvironment) Install(version string) error {
	gha.StartGroup("🔧 Bootstrapping GlassOps Runtime")
	defer gha.EndGroup()

	// Check if sf is already installed
	if _, err := exec.LookPath("sf"); err == nil {
		gha.Info("⚡ Salesforce CLI detected in environment. Skipping install.")
		return nil
	}

	if version == "" {
		version = "latest"
	}

	gha.Info(fmt.Sprintf("⬇️ Installing @salesforce/cli@%s...", version))

	// Retry npm install for transient network failures
	var lastErr error
	for attempt := 0; attempt < 3; attempt++ {
		cmd := exec.Command("npm", "install", "-g", fmt.Sprintf("@salesforce/cli@%s", version))
		if err := cmd.Run(); err != nil {
			lastErr = err
			time.Sleep(time.Duration(2000*(1<<attempt)) * time.Millisecond)
			continue
		}

		// Verify installation
		verifyCmd := exec.Command("sf", "version")
		if err := verifyCmd.Run(); err != nil {
			return fmt.Errorf("❌ Failed to verify CLI installation: %w", err)
		}

		return nil
	}

	return fmt.Errorf("❌ Failed to bootstrap runtime. NPM registry might be down: %w", lastErr)
}

// InstallPlugins installs and validates Salesforce CLI plugins.
func (r *RuntimeEnvironment) InstallPlugins(config *policy.Config, plugins []string) error {
	if len(plugins) == 0 {
		gha.Info("ℹ️ No plugins specified for installation.")
		return nil
	}

	gha.StartGroup("🔌 Installing Salesforce CLI Plugins")
	defer gha.EndGroup()

	policyEngine := policy.New()

	for _, plugin := range plugins {
		gha.Info(fmt.Sprintf("🔍 Validating plugin: %s", plugin))

		// Check whitelist
		if len(config.Governance.PluginWhitelist) > 0 {
			if !policyEngine.ValidatePluginWhitelist(config, plugin) {
				return fmt.Errorf("🚫 Plugin '%s' is not in the whitelist. Allowed: %s",
					plugin, strings.Join(config.Governance.PluginWhitelist, ", "))
			}

			// Get version constraint if specified
			constraint := policyEngine.GetPluginVersionConstraint(config, plugin)
			if constraint != "" {
				plugin = fmt.Sprintf("%s@%s", plugin, constraint)
			}
		} else {
			gha.Warning(fmt.Sprintf("⚠️ No plugin whitelist configured. Installing %s without validation.", plugin))
		}

		gha.Info(fmt.Sprintf("⬇️ Installing plugin: %s", plugin))
		if err := r.execWithAutoConfirm("sf", []string{"plugins", "install", plugin}); err != nil {
			return fmt.Errorf("❌ Failed to install plugin '%s': %w", plugin, err)
		}

		// Verify installation
		cmd := exec.Command("sf", "plugins", "--json")
		output, err := cmd.Output()
		if err != nil {
			return fmt.Errorf("failed to verify plugin installation: %w", err)
		}

		var installedPlugins []struct {
			Name string `json:"name"`
		}
		if err := json.Unmarshal(output, &installedPlugins); err != nil {
			// Try wrapped format
			var wrapped struct {
				Result []struct {
					Name string `json:"name"`
				} `json:"result"`
			}
			if err := json.Unmarshal(output, &wrapped); err != nil {
				return fmt.Errorf("unexpected output format from 'sf plugins --json'")
			}
			installedPlugins = wrapped.Result
		}

		found := false
		for _, p := range installedPlugins {
			if p.Name == plugin || strings.HasPrefix(plugin, p.Name+"@") {
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("plugin '%s' installation verification failed", plugin)
		}

		gha.Info(fmt.Sprintf("✅ Plugin '%s' installed and verified successfully", plugin))
	}

	return nil
}

func (r *RuntimeEnvironment) execWithAutoConfirm(command string, args []string) error {
	quotedArgs := make([]string, len(args))
	for i, arg := range args {
		quotedArgs[i] = fmt.Sprintf(`"%s"`, arg)
	}
	fullCommand := fmt.Sprintf("%s %s", command, strings.Join(quotedArgs, " "))

	var cmd *exec.Cmd
	if r.platform == "windows" {
		cmd = exec.Command("cmd", "/c", fmt.Sprintf("echo y|%s", fullCommand))
	} else {
		cmd = exec.Command("sh", "-c", fmt.Sprintf("echo y | %s", fullCommand))
	}

	return cmd.Run()
}
