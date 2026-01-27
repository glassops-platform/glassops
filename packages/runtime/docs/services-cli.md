---
type: Documentation
domain: runtime
origin: packages/runtime/src/services/cli.ts
last_modified: 2026-01-26
generated: true
source: packages/runtime/src/services/cli.ts
generated_at: 2026-01-26T14:20:43.787Z
hash: 83433df2f19b6c647e461c4f9b34d1190a77edb6a64f71bf0a4260b4ce58908c
---

## GlassOps Runtime Environment Documentation

This document details the functionality of the GlassOps Runtime Environment, a service responsible for managing the Salesforce CLI (sf) and its associated plugins. It ensures the necessary tools are present and configured correctly for subsequent operations.

**Overview**

The Runtime Environment provides methods for installing the Salesforce CLI and managing Salesforce CLI plugins. It handles versioning, whitelisting, and verification to maintain a secure and consistent environment.  This service is designed to operate within automated workflows, such as CI/CD pipelines, and provides detailed logging through the Actions core library.

**Key Features**

*   **Salesforce CLI Installation:** Automatically installs the Salesforce CLI if it is not already present on the system.  You can specify a version; otherwise, the latest version is installed.
*   **Plugin Management:** Installs, validates, and verifies Salesforce CLI plugins.
*   **Plugin Whitelisting:** Supports a plugin whitelist to enforce security policies and prevent the installation of unauthorized plugins.
*   **Version Control:**  Allows specifying version constraints for plugins, ensuring compatibility.
*   **Automated Confirmation:** Handles prompts requiring confirmation during installation processes.
*   **Platform Compatibility:**  Adapts installation commands based on the operating system (Windows, macOS, Linux).



**Functionality**

**1. Installation of Salesforce CLI**

The `install()` method checks for an existing Salesforce CLI installation. If not found, it installs the specified version (or the latest version if none is provided) using npm.  A verification step confirms the installation was successful.

**2. Plugin Installation**

The `installPlugins()` method manages the installation of Salesforce CLI plugins.  The process includes:

*   **Validation:** Checks if the plugin is permitted based on a configured whitelist. If no whitelist is configured, plugins are installed without validation (with a warning).
*   **Version Constraints:**  Respects version constraints defined in the configuration, installing the appropriate plugin version.
*   **Installation:** Executes the `sf plugins install` command.
*   **Verification:** Confirms the plugin was installed correctly by querying the installed plugins list.

**3. Automated Confirmation**

The `execWithAutoConfirm()` method handles commands that require user confirmation. It automatically provides the 'y' response to proceed with the installation, enabling non-interactive execution.

**Configuration**

Plugin installation behavior is governed by a `ProtocolConfig` object. Specifically, the `governance.plugin_whitelist` property defines a list of allowed plugins.  The policy engine uses this list to validate plugin installations.

**Error Handling**

The Runtime Environment includes robust error handling.  Installation failures result in descriptive error messages, aiding in troubleshooting.  Errors during plugin installation will halt the process and provide details about the failure.

**Usage**

I am designed to be used programmatically within automated workflows. You interact with me through my methods: `install()` and `installPlugins()`.  The `ProtocolConfig` object provides the necessary configuration for plugin validation.