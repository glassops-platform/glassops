---
type: Documentation
domain: runtime-ts
origin: packages/runtime-ts/src/services/cli.ts
last_modified: 2026-01-31
generated: true
source: packages/runtime-ts/src/services/cli.ts
generated_at: 2026-01-31T09:15:55.895477
hash: 973b422680240f152da76f14a615c66ecf48871b05f31f1e52fb30e52ddb2199
---

## GlassOps Runtime Environment Documentation

This document details the functionality of the GlassOps Runtime Environment, a component responsible for managing the Salesforce CLI (sf) and its associated plugins. It provides a consistent and reliable environment for executing GlassOps operations.

**Overview**

The Runtime Environment handles the installation and verification of the Salesforce CLI and specified plugins. It incorporates retry mechanisms for network instability and enforces plugin whitelisting based on configuration. This ensures a secure and predictable execution environment.

**Key Features**

*   **Salesforce CLI Installation:** Automatically installs the Salesforce CLI if it is not already present in the environment. Supports specifying a version during installation.
*   **Plugin Management:** Installs and verifies Salesforce CLI plugins.
*   **Plugin Whitelisting:**  Enforces a configurable whitelist of allowed plugins, enhancing security.  If no whitelist is configured, plugins are installed without validation.
*   **Version Constraints:** Supports specifying version constraints for plugins via the configuration.
*   **Retry Logic:** Includes retry mechanisms for installation processes to handle transient network errors.
*   **Automated Confirmation:** Handles prompts requiring confirmation during plugin installation.

**Usage**

The `RuntimeEnvironment` class provides the following methods:

*   **`install(version?: string)`:**
    *   Installs the Salesforce CLI if it is not already installed.
    *   The optional `version` parameter allows you to specify a specific version of the CLI to install (e.g., "7.100.0"). If no version is provided, the latest version is installed.
    *   You should call this method before attempting to install plugins.
*   **`installPlugins(config: ProtocolConfig, plugins: string[])`:**
    *   Installs a list of Salesforce CLI plugins.
    *   The `config` parameter provides configuration details, including plugin whitelist settings and version constraints.
    *   The `plugins` parameter is an array of plugin names to install.
    *   This method validates plugins against the configured whitelist (if present) and installs them with any specified version constraints.

**Configuration**

The behavior of the Runtime Environment is influenced by the `ProtocolConfig` object, specifically the `governance` property.

*   **`governance.plugin_whitelist`:** An array of strings representing the allowed plugin names. If this array is empty or not defined, plugin installation proceeds without whitelisting.

**Error Handling**

The Runtime Environment includes robust error handling.  Installation failures will result in exceptions with descriptive error messages.  These messages will assist in diagnosing and resolving issues.

**Platform Support**

This Runtime Environment is designed to function across multiple platforms, including Windows, macOS, and Linux. It adapts its execution commands based on the detected operating system.