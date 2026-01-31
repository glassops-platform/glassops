---
type: Documentation
domain: runtime-ts
origin: packages/runtime-ts/src/integration/cli.integration.test.ts
last_modified: 2026-01-31
generated: true
source: packages/runtime-ts/src/integration/cli.integration.test.ts
generated_at: 2026-01-31T10:08:19.016598
hash: 7bdd3344b00f64cd730b3fc47b16f07a62fb9c929c986afde8dc397872fe94f8
---

## Runtime Environment Integration Test Documentation

This document details the integration tests for the Runtime Environment, focusing on CLI installation and plugin management. These tests ensure the environment functions correctly with dynamic module loading while isolating external I/O operations through mocking.

### Overview

The Runtime Environment provides functionality for installing the Salesforce CLI and managing its plugins.  The tests verify correct behavior in scenarios including: CLI presence, installation success/failure, plugin whitelisting, version constraints, and overall workflow execution.

### Key Components

*   **RuntimeEnvironment:** The primary class under test, responsible for CLI and plugin management.
*   **ProtocolConfig:**  A configuration object defining runtime settings, including CLI version, Node version, and plugin whitelist.
*   **Mocked Actions:**  External I/O operations are mocked using `@actions/exec`, `@actions/io`, and `@actions/core` modules.

### CLI Installation Tests

These tests verify the installation process for the Salesforce CLI.

*   **CLI Already Present:** If the CLI is detected in the environment, installation is skipped.
*   **CLI Not Present:** If the CLI is not found, it is installed using `npm install -g @salesforce/cli@<version>`.
*   **Installation Failure:**  Errors during installation are handled gracefully, and an error is thrown.

### Plugin Installation Tests

These tests focus on installing and verifying plugins based on a configured whitelist.

*   **No Whitelist:** When no whitelist is configured, plugins are installed without validation, and a warning is issued.
*   **Whitelisted Plugin with Version Constraint:** Plugins specified in the whitelist with version constraints (e.g., `sfdx-hardis@^6.0.0`) are installed with the specified constraint.
*   **Whitelisted Plugin without Version Constraint:** Plugins specified in the whitelist without a version constraint (e.g., `sfdx-hardis`) are installed using the latest available version.
*   **Scoped Package Installation:** Installation of scoped packages (e.g., `@salesforce/plugin-deploy-retrieve@^3.0.0`) with version constraints is verified.
*   **Non-Whitelisted Plugin:** Attempts to install plugins not present in the whitelist are rejected with an error.
*   **Multiple Plugins:**  Multiple whitelisted plugins can be installed in a single operation.
*   **No Plugins Specified:** If no plugins are provided for installation, a message is logged, and no installation attempt is made.
*   **Plugin Verification:** After installation, the environment verifies the plugin is installed correctly by querying the CLI.
*   **Verification Failure:** If plugin verification fails (e.g., plugin not found), an error is thrown.
*   **Installation Error Handling:** Errors during plugin installation are handled, and an appropriate error message is displayed.

### End-to-End Workflow Test

This test validates the complete plugin installation workflow, from start to finish, including validation and verification steps. It confirms the correct sequence of operations and logging messages.

### Platform Considerations

The tests adapt to the operating system (Windows or other) when constructing shell commands, using `cmd` and `/c` for Windows and `sh` and `-c` for other systems.