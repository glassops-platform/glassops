---
type: Documentation
domain: runtime-ts
origin: packages/runtime-ts/src/services/cli.test.ts
last_modified: 2026-01-31
generated: true
source: packages/runtime-ts/src/services/cli.test.ts
generated_at: 2026-01-31T09:15:37.541231
hash: b3234a4c5d9ab2fe336c4578cdcdb047c8240534be58cb563e50a98b33158e5d
---

## Runtime Environment Documentation

This document details the functionality of the Runtime Environment service, responsible for managing the Salesforce CLI (sf CLI) and its plugins. It is intended for both technical and non-technical users.

**Overview**

The Runtime Environment service ensures the necessary Salesforce CLI and plugins are installed and configured for operation. It handles installation, version management, and validation, providing a consistent and reliable environment.

**Functionality**

The service provides the following core functions:

*   **CLI Installation:**  Automatically installs the Salesforce CLI if it is not already present on the system. It supports installing the latest version or a specific version as requested.
*   **Plugin Management:** Installs and validates Salesforce CLI plugins based on a provided configuration.  A whitelist mechanism controls which plugins are permitted, enhancing security.
*   **Environment Validation:** Checks for existing CLI installations to avoid redundant installations.
*   **Error Handling:**  Provides robust error handling, including retries for installation failures and informative error messages.

**Installation Process**

1.  **Check for Existing CLI:** The service first verifies if the Salesforce CLI is already installed.
2.  **Install if Necessary:** If the CLI is not found, it is installed using npm (Node Package Manager).
3.  **Version Verification:** After installation, the service verifies the CLI version.
4.  **Plugin Installation:**  If plugins are specified, the service proceeds to install them according to the configured policy.

**Plugin Installation Details**

*   **Whitelist Enforcement:** Plugin installation is governed by a whitelist. Only plugins listed in the whitelist are installed.  If no whitelist is configured, a warning is issued, and plugins are installed without validation.
*   **Verification:** After installation, each plugin is verified to ensure it was installed correctly.
*   **Version Constraints:** The service supports installing plugins with specific version constraints (e.g., `sfdx-hardis@^6.0.0`).
*   **Platform Considerations:** The service adapts to the operating system, using the appropriate shell (cmd on Windows, sh on other platforms) for plugin installation.

**Configuration**

The service relies on a `ProtocolConfig` object to determine its behavior. Key configuration elements include:

*   `governance.enabled`:  A boolean indicating whether governance features (like the plugin whitelist) are enabled.
*   `governance.plugin_whitelist`: An array of strings representing the allowed plugins.
*   `runtime.cli_version`: Specifies the desired Salesforce CLI version ("latest" or a specific version number).
*   `runtime.node_version`: Specifies the required Node.js version.

**Error Handling and Reporting**

The service provides detailed error messages to assist in troubleshooting. Common error scenarios include:

*   **NPM Installation Failure:** Indicates a problem with the npm registry or network connectivity.
*   **CLI Version Check Failure:**  Indicates an issue verifying the installed CLI version.
*   **Plugin Installation Failure:** Indicates a problem during plugin installation.
*   **Plugin Verification Failure:** Indicates that a plugin could not be verified after installation.
*   **Non-Whitelisted Plugin:**  An error is thrown if an attempt is made to install a plugin not present in the whitelist.

**Usage Notes**

You can trigger the installation and plugin management functions programmatically. The service provides clear error messages to guide you in resolving any issues.