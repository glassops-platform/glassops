---
type: Documentation
domain: runtime-ts
origin: packages/runtime-ts/src/integration/cli.integration.test.ts
last_modified: 2026-01-31
generated: true
source: packages/runtime-ts/src/integration/cli.integration.test.ts
generated_at: 2026-01-31T09:12:05.516686
hash: 7bdd3344b00f64cd730b3fc47b16f07a62fb9c929c986afde8dc397872fe94f8
---

## Runtime Environment CLI Integration Test Documentation

This document details the integration tests for the Runtime Environment Command Line Interface (CLI). These tests confirm the correct operation of the `RuntimeEnvironment` class, specifically focusing on module loading – including dynamic imports – while simulating external input/output operations.

**Purpose**

The primary goal of these tests is to ensure reliable CLI functionality, including installation and plugin management, within the broader runtime environment.

**Functionality Tested**

The tests cover the following key areas:

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

*   **governance.enabled:** A boolean indicating whether governance features are enabled.
*   **governance.plugin\_whitelist:** An array of strings representing allowed plugins and their optional version constraints (e.g., "sfdx-hardis@^6.0.0").
*   **runtime.cli\_version:** The desired CLI version.
*   **runtime.node\_version:** The required Node.js version.

**Behavior Notes**

*   The tests mock external interactions with the operating system and other tools (like `npm` and `sf`) to provide a controlled testing environment.
*   Installation commands are constructed differently based on the operating system (Windows vs. others) to ensure compatibility.
*   Version constraints are enforced during plugin installation when a whitelist is configured.
*   Error handling is implemented to gracefully manage installation failures and provide informative error messages.
*   Informational messages are logged to indicate the progress of installation and verification steps.

**User Instructions**

You do not need to directly interact with these tests. They are part of the internal build and validation process. However, understanding the tested functionality provides insight into the reliability and security features of the runtime environment.