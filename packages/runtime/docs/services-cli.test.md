---
type: Documentation
domain: runtime
origin: packages/runtime/src/services/cli.test.ts
last_modified: 2026-01-26
generated: true
source: packages/runtime/src/services/cli.test.ts
generated_at: 2026-01-26T14:20:22.603Z
hash: 0a1e43ffbd83c287cb055a1c5a61a2af7a8a488b81211e0f0e24dd4bb4519af7
---

## Runtime Environment Documentation

This document details the functionality of the Runtime Environment service, responsible for managing the Salesforce CLI (sf CLI) and its plugins. It outlines installation, verification, and governance features.

**Overview**

The Runtime Environment provides methods to ensure the necessary sf CLI and plugins are present and correctly configured for operation. It handles installation, version checks, and plugin whitelisting to maintain a secure and predictable environment.

**Functionality**

**1. Installation ( `runtime.install()` )**

This function manages the installation of the sf CLI.

*   **Detection:** First, it checks if the sf CLI is already available in the system’s PATH. If found, installation is skipped.
*   **Installation Process:** If the sf CLI is not found, it installs the latest version (or a specified version) using npm.  After installation, it verifies the installation by running `sf version`.
*   **Error Handling:**  The function throws an error if either the npm installation or the sf CLI version check fails, indicating potential issues with the npm registry or the CLI itself.
*   **Logging:**  Installation attempts and skips are logged for transparency.  A start and end group is used to delineate the installation process.

**Parameters:**

*   `version` (optional):  A string specifying the desired sf CLI version (e.g., "2.50.0"). If omitted, the latest version is installed.

**2. Plugin Management ( `runtime.installPlugins()` )**

This function manages the installation of Salesforce CLI plugins.

*   **Plugin Whitelisting:**  It enforces a plugin whitelist defined in the `ProtocolConfig`. Only plugins present in the whitelist are installed. If no whitelist is configured, a warning is issued, and all plugins are permitted.
*   **Installation Process:** For each whitelisted plugin, it executes the `sf plugins install` command.
*   **Verification:** After installation, it verifies the installation by querying installed plugins using `sf plugins --json`.
*   **Error Handling:**  The function throws errors in the following scenarios:
    *   Attempting to install a plugin not on the whitelist.
    *   Failure during plugin installation.
    *   Unexpected output format from the `sf plugins --json` command.
    *   Plugin verification failing (plugin not found after installation).
*   **Logging:**  Plugin validation, installation attempts, and any errors are logged.
*   **Platform Considerations:** Uses the appropriate shell (`sh` on Linux/macOS, `cmd` on Windows) and command flags for plugin installation.

**Parameters:**

*   `config`: A `ProtocolConfig` object containing governance settings, including the plugin whitelist.
*   `plugins`: An array of strings, where each string is the name of a plugin to install.

**Configuration ( `ProtocolConfig` )**

The `ProtocolConfig` object governs the behavior of the Runtime Environment. Relevant properties include:

*   `governance.enabled`: A boolean indicating whether governance features (like plugin whitelisting) are enabled.
*   `governance.plugin_whitelist`: An array of strings representing the allowed plugins.  Can be undefined to allow all plugins.
*   `runtime.cli_version`: The desired CLI version.
*   `runtime.node_version`: The required Node.js version.

**Dependencies**

*   `@actions/exec`: Used for executing shell commands.
*   `@actions/io`: Used for checking if a command is available in the system’s PATH.
*   `@actions/core`: Used for logging and managing action groups.