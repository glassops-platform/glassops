---
type: Documentation
domain: runtime-ts
origin: packages/runtime-ts/src/services/cli.ts
last_modified: 2026-01-29
generated: true
source: packages/runtime-ts/src/services/cli.ts
generated_at: 2026-01-29T20:58:54.200415
hash: 973b422680240f152da76f14a615c66ecf48871b05f31f1e52fb30e52ddb2199
---

## GlassOps Runtime Environment Documentation

This document details the functionality of the GlassOps Runtime Environment, a component responsible for managing the Salesforce CLI (sf) and its associated plugins. It provides a consistent and reliable environment for executing GlassOps operations.

**Overview**

The Runtime Environment handles the installation and verification of the Salesforce CLI and plugins required for specific operations. It incorporates retry mechanisms for network instability and enforces plugin whitelisting based on configured policies. This ensures a secure and predictable execution environment.

**Key Features**

*   **Salesforce CLI Installation:** Automatically installs the Salesforce CLI if it is not already present in the environment. You can specify a version during installation; otherwise, the latest version is installed.
*   **Plugin Management:** Installs and verifies Salesforce CLI plugins. Supports installation of multiple plugins.
*   **Plugin Whitelisting:**  Enforces a configurable whitelist of allowed plugins, enhancing security and governance.  If no whitelist is configured, plugins are installed without validation.
*   **Version Constraints:** Supports version constraints for plugins as defined in the configuration policy.
*   **Automated Confirmation:** Handles prompts requiring confirmation during plugin installation.
*   **Error Handling:** Provides detailed error messages for installation failures, aiding in troubleshooting.
*   **Retry Logic:** Implements retry mechanisms for Salesforce CLI installation to mitigate transient network issues.

**Usage**

The `RuntimeEnvironment` class provides the following methods:

*   **`install(version?: string): Promise<void>`**
    Installs the Salesforce CLI.
    *   `version` (optional): The version of the Salesforce CLI to install. Defaults to "latest".
*   **`installPlugins(config: ProtocolConfig, plugins: string[]): Promise<void>`**
    Installs a list of Salesforce CLI plugins.
    *   `config`:  The configuration object containing governance policies, including the plugin whitelist.
    *   `plugins`: An array of plugin names to install.

**Configuration**

The behavior of the Runtime Environment is influenced by the `ProtocolConfig` object, specifically the `governance` property.

*   **`governance.plugin_whitelist`**: An array of strings representing the allowed plugin names. If this array is empty or not defined, plugin installation proceeds without whitelisting.

**Verification**

After plugin installation, the Runtime Environment verifies the installation by querying the Salesforce CLI for a list of installed plugins.  It confirms that the requested plugin is present in the list.

**Platform Support**

The Runtime Environment adapts to the operating system (Windows, macOS, Linux) to handle automated confirmation prompts appropriately.