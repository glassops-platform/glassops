---
type: Documentation
domain: runtime-ts
origin: packages/runtime-ts/src/integration/cli.integration.test.ts
last_modified: 2026-01-29
generated: true
source: packages/runtime-ts/src/integration/cli.integration.test.ts
generated_at: 2026-01-29T20:55:13.614677
hash: 7bdd3344b00f64cd730b3fc47b16f07a62fb9c929c986afde8dc397872fe94f8
---

## Runtime Environment CLI Integration Test Documentation

This document details the integration tests for the Runtime Environment Command Line Interface (CLI). These tests confirm the correct operation of the `RuntimeEnvironment` class, specifically focusing on module loading – including dynamic imports – while simulating external input/output operations.

**Purpose**

The primary goal of these tests is to ensure reliable CLI functionality, including installation and plugin management, within the broader runtime environment.

**Functionality Tested**

The integration tests cover the following key areas:

*   **CLI Installation:**
    *   Skipping installation if the CLI is already present.
    *   Installing the CLI when it is not found.
    *   Handling installation failures.
*   **Plugin Installation with Whitelist:**
    *   Installing plugins without a whitelist configuration.
    *   Installing whitelisted plugins with version constraints.
    *   Installing scoped packages with version constraints.
    *   Installing whitelisted plugins without version constraints.
    *   Rejecting non-whitelisted plugins.
    *   Installing multiple whitelisted plugins.
    *   Skipping installation when no plugins are specified.
    *   Verifying plugin installation.
    *   Handling plugin verification failures.
    *   Handling plugin installation errors.
*   **End-to-End Plugin Flow:**
    *   Validating the complete plugin installation workflow from start to finish.

**Configuration**

Plugin installation is governed by a `ProtocolConfig` object, which includes:

*   `governance.enabled`: A boolean indicating whether governance features are enabled.
*   `governance.plugin_whitelist`: An array of strings representing allowed plugins and their optional version constraints (e.g., `"sfdx-hardis@^6.0.0"`).
*   `runtime.cli_version`: The desired CLI version.
*   `runtime.node_version`: The required Node.js version.

**Behavior Notes**

*   The tests mock external interactions with the operating system and other tools (like `npm` and `sf`) to provide a controlled testing environment.
*   The tests adapt to the operating system (Windows or others) when constructing shell commands.
*   Version constraints (e.g., `^6.0.0`) are enforced during plugin installation when a whitelist is configured.
*   If a plugin is not found in the whitelist, installation is prevented, and an error is thrown.
*   Informational messages and warnings are logged to provide feedback on the installation process.
*   Error handling is implemented to gracefully manage installation failures and provide informative error messages.

**User Instructions**

You do not need to directly interact with these tests. They are part of the internal build and validation process. However, understanding the tested functionality provides insight into the reliability and security features of the runtime environment.