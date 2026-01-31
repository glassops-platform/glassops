---
type: Documentation
domain: runtime
origin: packages/runtime/internal/services/cli.go
last_modified: 2026-01-31
generated: true
source: packages/runtime/internal/services/cli.go
generated_at: 2026-01-31T09:07:58.341380
hash: bdffb7a1ab67d92a28eb7709a10ddc33e93f9484d200010ddb9f2d4eb003dcb0
---

## GlassOps Runtime Environment Service Documentation

This document details the `RuntimeEnvironment` service, responsible for managing the Salesforce CLI and its associated plugins. It provides an overview of the service’s purpose, key components, and functionality.

**Package Purpose:**

The `services` package contains implementations of runtime services for GlassOps. This specific service, `RuntimeEnvironment`, focuses on ensuring the Salesforce CLI is installed and configured correctly, along with the necessary plugins defined by policy.

**Key Types and Interfaces:**

*   **`RuntimeEnvironment` struct:** This structure encapsulates the runtime environment’s state, currently storing the operating system platform. It provides methods for installing the CLI and managing plugins.
*   **`policy.Config` struct:** (From the `policy` package) Represents the configuration settings related to governance, including plugin whitelists and version constraints.
*   **`policy.PolicyEngine` interface:** (From the `policy` package) Provides methods for validating plugins against configured policies, such as checking the whitelist and retrieving version constraints.

**Important Functions and Behavior:**

*   **`NewRuntimeEnvironment()`:**  This function creates and returns a new `RuntimeEnvironment` instance, initializing it with the operating system detected from the `GOOS` environment variable.
*   **`Install(version string)`:** This function installs the Salesforce CLI (`sf`) if it is not already present on the system. It accepts an optional version string; if none is provided, it defaults to installing the latest version.  The installation is performed using `npm install -g @salesforce/cli@version`. It includes retry logic with exponential backoff to handle transient network issues during the installation process.  After installation, it verifies the installation by running `sf version`.
*   **`InstallPlugins(config *policy.Config, plugins []string)`:** This function installs a list of Salesforce CLI plugins. It iterates through the provided plugin list, performing the following steps for each plugin:
    *   **Policy Validation:** Checks if a plugin whitelist is configured. If so, it validates the plugin against the whitelist using the `policy.PolicyEngine`. If a version constraint is defined in the policy, it appends it to the plugin name during installation. If no whitelist is configured, a warning is logged.
    *   **Plugin Installation:** Installs the plugin using `sf plugins install plugin`.
    *   **Installation Verification:**  Verifies the installation by running `sf plugins --json` and parsing the output to confirm the plugin is listed as installed. It handles potential variations in the JSON output format.
*   **`execWithAutoConfirm(command string, args []string)`:** This helper function executes a shell command with automatic confirmation for prompts. It handles differences between Windows and other operating systems by using `echo y|` on non-Windows systems and `echo y|` within `cmd /c` on Windows. This is used to bypass prompts during plugin installation.

**Error Handling:**

The service employs robust error handling:

*   Functions return explicit `error` values to indicate failure.
*   Errors are wrapped using `fmt.Errorf("%w", err)` to preserve the original error context.
*   The `Install` function includes retry logic for transient network errors.
*   Plugin installation failures result in informative error messages including the plugin name and the underlying error.
*   JSON parsing errors during plugin verification are handled with fallback logic to accommodate different output formats.

**Concurrency:**

This service does not currently employ explicit concurrency patterns like goroutines or channels. All operations are performed sequentially within the calling function.

**Design Decisions:**

*   **npm as Installation Method:** The Salesforce CLI is installed using `npm` because it is a common package manager and simplifies the installation process.
*   **Automatic Confirmation:** The `execWithAutoConfirm` function is used to automatically answer prompts during plugin installation, enabling unattended operation.
*   **Plugin Verification:**  The service verifies plugin installations by parsing the output of `sf plugins --json` to ensure reliability.
*   **Policy Integration:** The service integrates with a policy engine to enforce governance rules regarding plugin usage.