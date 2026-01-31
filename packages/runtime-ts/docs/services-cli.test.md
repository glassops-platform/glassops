---
type: Documentation
domain: runtime-ts
origin: packages/runtime-ts/src/services/cli.test.ts
last_modified: 2026-01-29
generated: true
source: packages/runtime-ts/src/services/cli.test.ts
generated_at: 2026-01-29T20:58:35.305482
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

**Plugin Configuration and Security**

Plugin installation is governed by a configuration that includes a plugin whitelist.

*   **Whitelist:**  Only plugins listed in the whitelist are installed. This prevents the installation of potentially harmful or unauthorized plugins.
*   **No Whitelist:** If no whitelist is configured, a warning is issued, and all requested plugins are installed without validation.
*   **Version Constraints:** The service supports installing plugins with specific version constraints (e.g., `sfdx-hardis@^6.0.0`).

**Error Scenarios**

The service handles several potential error scenarios:

*   **npm Installation Failure:** If the npm installation fails (e.g., due to network issues), the service retries the installation.  If retries fail, an error is thrown.
*   **CLI Version Check Failure:** If verifying the CLI version fails, an error is thrown.
*   **Plugin Installation Failure:** If a plugin installation fails, an error is thrown.
*   **Plugin Verification Failure:** If the installed plugins cannot be verified, an error is thrown.
*   **Non-Whitelisted Plugins:** Attempts to install plugins not on the whitelist result in an error.
*   **Unexpected Output Format:** Errors are thrown if the output from CLI commands is not in the expected format.

**Platform Considerations**

The service adapts to the operating system:

*   **Windows:** Uses `cmd` shell with the `/c` flag for executing commands.
*   **Non-Windows:** Uses `sh` shell with the `-c` flag.

**Usage Notes**

You can specify the CLI version to install. If no version is provided, the latest version is installed.  Plugin installation is controlled by the configuration file.  Ensure the configuration file is correctly set up to define the desired plugins and security policies.