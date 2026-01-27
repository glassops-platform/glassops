---
type: Documentation
domain: runtime
origin: packages/runtime/src/integration/cli-platform.integration.test.ts
last_modified: 2026-01-26
generated: true
source: packages/runtime/src/integration/cli-platform.integration.test.ts
generated_at: 2026-01-26T14:14:14.871Z
hash: aca88f7feb985cd7992e45265428d1ce06de6852e0875ac9be8f5c0fb655b495
---

## CLI Platform Integration Test Documentation

This document details integration tests designed to verify the correct operation of the CLI across different operating systems – Windows, macOS, and Linux. These tests ensure platform-specific behaviors, particularly around command execution and plugin installation, function as expected.

**Purpose**

The primary goal of these tests is to confirm that the CLI operates consistently and reliably regardless of the underlying platform. This includes verifying the correct shell commands are used for each operating system and that plugin installation processes handle platform-specific requirements.

**Scope**

The tests cover the following key areas:

*   **Plugin Installation:** Validates that plugins are installed correctly on each platform, utilizing the appropriate command prefixes (e.g., `echo y|` for Windows, `echo y |` for Unix-based systems). Error handling during plugin installation is also tested.
*   **CLI Installation:** Confirms that the CLI is installed when not already present and skipped if it exists.  Tests also verify error handling during the CLI installation process.
*   **Cross-Platform Command Execution:** Ensures commands are executed using the correct shell (cmd on Windows, sh on Unix-like systems) for consistent behavior.

**Key Components**

*   **RuntimeEnvironment:** This class encapsulates the logic for interacting with the CLI, including installation and plugin management.
*   **ProtocolConfig:**  This configuration object defines runtime settings, including plugin whitelists and CLI/Node versions.
*   **Mocking:** The tests extensively use mocking of external GitHub Actions modules (`@actions/exec`, `@actions/io`) to isolate the CLI functionality and control the execution environment.

**Test Methodology**

The tests operate by:

1.  Temporarily modifying the `process.platform` property to simulate different operating systems.
2.  Creating an instance of the `RuntimeEnvironment` class.
3.  Calling methods on the `RuntimeEnvironment` instance to perform CLI operations (e.g., `installPlugins`, `install`).
4.  Asserting that the expected shell commands were called with the correct arguments using mocked functions.
5.  Restoring the original `process.platform` value after each test to avoid impacting other tests.

**Configuration**

You can configure the tests by modifying the `ProtocolConfig` object. Specifically, the `governance.plugin_whitelist` property controls which plugins are allowed to be installed.  The `runtime.cli_version` and `runtime.node_version` properties define the desired CLI and Node.js versions.

**Error Handling**

The tests include scenarios to verify that errors during plugin installation and CLI installation are handled gracefully and propagate appropriately.  Expectations are set to confirm that specific error messages are thrown when failures occur.