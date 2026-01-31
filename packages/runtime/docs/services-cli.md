---
type: Documentation
domain: runtime
origin: packages/runtime/internal/services/cli.go
last_modified: 2026-01-31
generated: true
source: packages/runtime/internal/services/cli.go
generated_at: 2026-01-31T10:03:57.856378
hash: bdffb7a1ab67d92a28eb7709a10ddc33e93f9484d200010ddb9f2d4eb003dcb0
---

## GlassOps Runtime Environment Service Documentation

This document details the `services` package, specifically the `RuntimeEnvironment` service, responsible for managing the Salesforce CLI and its plugins. This service ensures the necessary tools are present and configured correctly for GlassOps operations.

**Package Purpose:**

The `services` package provides implementations for runtime services within GlassOps. The `RuntimeEnvironment` within this package focuses on the installation and management of the Salesforce CLI (sf) and associated plugins.

**Key Types:**

*   **`RuntimeEnvironment`**: This struct encapsulates the runtime environment’s configuration. Currently, it stores the operating system (`platform`) detected from the `GOOS` environment variable. This allows for platform-specific command execution.

**Important Functions:**

*   **`NewRuntimeEnvironment()`**: This function creates and returns a new instance of the `RuntimeEnvironment` struct, initializing the `platform` field with the value of the `GOOS` environment variable.

*   **`Install(version string)`**: This function installs the Salesforce CLI if it is not already present on the system.
    *   It first checks if `sf` is in the system’s PATH. If found, it skips the installation.
    *   If a `version` is provided, it attempts to install that specific version. If no version is given, it defaults to installing the “latest” version.
    *   The installation is performed using `npm install -g @salesforce/cli@<version>`.
    *   The function includes retry logic with exponential backoff to handle transient network issues during the `npm install` process. It retries up to three times.
    *   After installation, it verifies the installation by running `sf version`.
    *   Errors are returned if the installation or verification fails.

*   **`InstallPlugins(config *policy.Config, plugins []string)`**: This function installs and validates a list of Salesforce CLI plugins.
    *   It iterates through the provided `plugins` slice.
    *   It checks a plugin whitelist defined in the provided `policy.Config`. If a whitelist is configured, it validates that each plugin is present in the whitelist using the `policyEngine.ValidatePluginWhitelist` method.  Version constraints from the policy are also applied.
    *   If no whitelist is configured, a warning is logged, and the plugin is installed without validation.
    *   Plugins are installed using `sf plugins install <plugin>`.
    *   After installation, the function verifies the installation by running `sf plugins --json` and parsing the output to confirm the plugin is listed as installed. It handles two possible JSON output formats.
    *   Errors are returned if plugin installation or verification fails.

*   **`execWithAutoConfirm(command string, args []string)`**: This helper function executes a shell command with automatic confirmation for prompts.
    *   It constructs the full command string, quoting each argument to handle spaces and special characters.
    *   It uses different shell commands based on the operating system (`platform`). On Windows, it uses `cmd /c "echo y| <command>"`. On other systems, it uses `sh -c "echo y | <command>"`. This automatically answers "yes" to any prompts during command execution.
    *   It returns an error if the command execution fails.

**Error Handling:**

The service employs standard Go error handling practices. Functions return an `error` value to indicate failure. Errors are often wrapped using `fmt.Errorf("%w", err)` to provide context and preserve the original error for debugging.  Retry logic is implemented in the `Install` function to handle transient errors.

**Concurrency:**

This service does not currently employ explicit concurrency patterns like goroutines or channels. All operations are performed sequentially within the calling function.

**Design Decisions:**

*   **Platform-Specific Execution:** The `execWithAutoConfirm` function uses different shell commands based on the operating system to ensure compatibility.
*   **Plugin Whitelisting:** The use of a plugin whitelist provides a governance mechanism to control which plugins can be installed, enhancing security and stability.
*   **Installation Verification:**  The service verifies plugin installations by parsing the output of `sf plugins --json` to ensure that the plugins are correctly installed and reported by the CLI.
*   **Retry Logic:** The `Install` function includes retry logic to handle intermittent network issues during the `npm install` process.