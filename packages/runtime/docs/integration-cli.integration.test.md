---
type: Documentation
domain: runtime
origin: packages/runtime/src/integration/cli.integration.test.ts
last_modified: 2026-01-26
generated: true
source: packages/runtime/src/integration/cli.integration.test.ts
generated_at: 2026-01-26T14:14:33.353Z
hash: 23661b48e1ad5b57732587612650c8478219d2ca8d2811ef6b0c753ae5457306
---

## Runtime Environment CLI Integration Tests Documentation

This document details the integration tests for the Runtime Environment Command Line Interface (CLI). These tests ensure the Runtime Environment functions correctly with module loading, specifically dynamic imports, while simulating external input/output operations.

**Purpose**

The primary goal of these tests is to validate the installation and management of the Salesforce CLI and its plugins within the runtime environment. This includes verifying correct installation procedures, adherence to plugin whitelists, and graceful handling of potential errors.

**Key Components**

*   **RuntimeEnvironment:** The core class responsible for managing the CLI lifecycle, including installation and plugin management.
*   **ProtocolConfig:**  A configuration object that defines runtime settings, including CLI version, Node.js version, and plugin governance policies (e.g., whitelists).
*   **Mocking:** External dependencies (GitHub Actions modules for execution, I/O, and logging) are mocked to isolate the Runtime Environment and ensure predictable test results.

**Functionality Tested**

**1. CLI Installation**

*   **Existing CLI Detection:** The system correctly identifies when the Salesforce CLI is already installed and skips the installation process.
*   **CLI Installation:** When the CLI is not present, the system installs it using npm.
*   **Installation Failure Handling:** The system gracefully handles errors during the CLI installation process.

**2. Plugin Installation with Whitelist**

*   **No Whitelist Configuration:** When no plugin whitelist is defined, plugins are installed without validation, accompanied by a warning message.
*   **Whitelisted Plugin Installation:** Plugins listed in the whitelist are installed with specified version constraints (e.g., `^6.0.0`).
*   **Scoped Package Installation:** Installation of scoped packages (e.g., `@salesforce/plugin-deploy-retrieve`) with version constraints is validated.
*   **Non-Whitelisted Plugin Rejection:**  Plugins not present in the whitelist are rejected, preventing their installation.
*   **Multiple Plugin Installation:** The system can install multiple whitelisted plugins simultaneously.
*   **No Plugin Specified:** The system handles the case where no plugins are specified for installation.
*   **Plugin Verification:** After installation, the system verifies the plugin is installed correctly by querying the CLI.
*   **Plugin Verification Failure:** The system handles failures during plugin verification.
*   **Plugin Installation Error Handling:** The system gracefully handles errors during plugin installation.

**3. End-to-End Plugin Flow**

*   **Complete Workflow Validation:**  Tests the entire plugin installation workflow, from validation to installation and verification, ensuring all steps function correctly.

**Configuration**

The `ProtocolConfig` object is central to controlling the runtime environment. Key settings include:

*   `governance.enabled`: Enables or disables plugin governance features.
*   `governance.plugin_whitelist`: A list of allowed plugins, potentially including version constraints.
*   `runtime.cli_version`: Specifies the desired Salesforce CLI version.
*   `runtime.node_version`: Specifies the required Node.js version.

**Platform Considerations**

The tests account for differences between operating systems (Windows and others) when constructing shell commands. Specifically, the tests adjust the shell executable and command flags accordingly.