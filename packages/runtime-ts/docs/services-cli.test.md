---
type: Documentation
domain: runtime-ts
origin: packages/runtime-ts/src/services/cli.test.ts
last_modified: 2026-01-31
generated: true
source: packages/runtime-ts/src/services/cli.test.ts
generated_at: 2026-01-31T10:11:59.739031
hash: b3234a4c5d9ab2fe336c4578cdcdb047c8240534be58cb563e50a98b33158e5d
---

## Runtime Environment Documentation

This document details the functionality of the Runtime Environment service, responsible for bootstrapping the Salesforce CLI (sf CLI) and its plugins. It is intended for both technical and non-technical users.

**Overview**

The Runtime Environment service ensures the necessary tools are available for subsequent operations. It handles installation and verification of the sf CLI and specified plugins, adhering to configured governance policies.

**Functionality**

The service provides two primary functions:

1.  **Installation (install)**: This function manages the installation of the sf CLI.
    *   **Detection**: It first checks if the sf CLI is already present in the environment. If found, installation is skipped.
    *   **Installation Process**: If the sf CLI is not found, it installs the specified version (or the latest version if none is provided) using npm.
    *   **Verification**: After installation, it verifies the installation by running `sf version`.
    *   **Error Handling**:  The function includes retry logic for npm installation failures and throws an error if bootstrapping fails. It also handles failures during the version check.
    *   **Logging**:  The installation process is logged with informative messages indicating the steps taken.

2.  **Plugin Installation (installPlugins)**: This function manages the installation of Salesforce CLI plugins.
    *   **Configuration**: It utilizes a `ProtocolConfig` object to determine which plugins to install and whether a plugin whitelist is enforced.
    *   **Whitelist Enforcement**: If a whitelist is configured, only plugins present on the list are installed.  Installation of non-whitelisted plugins is rejected.
    *   **No Whitelist Handling**: If no whitelist is configured, a warning is issued, and plugins are installed without validation.
    *   **Installation Process**: Plugins are installed using the `sf plugins install` command.
    *   **Verification**: After installation, the service verifies the installation by querying installed plugins using `sf plugins --json`.
    *   **Error Handling**: The function handles errors during plugin installation and verification, providing informative error messages.
    *   **Version Constraints**: Supports installing plugins with version constraints specified in the whitelist (e.g., `sfdx-hardis@^6.0.0`).
    *   **Shell Selection**: Uses the appropriate shell (`cmd` on Windows, `sh` on other platforms) for executing commands.

**ProtocolConfig**

The `ProtocolConfig` object contains configuration settings for the runtime environment. Relevant properties include:

*   `governance.enabled`: A boolean indicating whether governance features (like plugin whitelisting) are enabled.
*   `governance.plugin_whitelist`: An array of strings representing the allowed plugins.  Can be undefined to disable the whitelist.
*   `runtime.cli_version`: The desired version of the sf CLI (e.g., "latest", "2.50.0").
*   `runtime.node_version`: The required Node.js version.

**Error Handling**

The service provides robust error handling, throwing exceptions with descriptive messages when issues occur during installation or verification. These messages aid in troubleshooting and identifying potential problems with the environment.

**Platform Considerations**

The service adapts to the operating system, using the appropriate shell and command syntax for plugin installation.