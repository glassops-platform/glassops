---
type: Documentation
domain: runtime-ts
origin: packages/runtime-ts/src/integration/cli-platform.integration.test.ts
last_modified: 2026-01-31
generated: true
source: packages/runtime-ts/src/integration/cli-platform.integration.test.ts
generated_at: 2026-01-31T10:08:01.679233
hash: c0fe62e004405747ec5aed3472ced22489880f70c7e4fec6ba9821c76fff59c7
---

## CLI Platform Integration Test Documentation

This document details integration tests designed to verify the correct operation of command-line interface (CLI) functions across different operating systems – Windows, macOS, and Linux. These tests ensure consistent behavior regardless of the underlying platform.

**Purpose**

The primary goal of these tests is to validate platform-specific CLI operations, including plugin installation and command execution. This ensures the system functions as expected in diverse environments.

**Scope**

The tests cover the following key areas:

*   **Plugin Installation:** Verifies that plugins are installed correctly on each platform, utilizing the appropriate shell commands (e.g., `cmd` on Windows, `sh` on Unix-based systems).  The tests confirm correct handling of both successful installations and failure scenarios.
*   **CLI Installation:** Validates the installation process of the CLI itself, including scenarios where the CLI is already present on the system and cases where installation fails.
*   **Cross-Platform Command Execution:** Confirms that commands are executed using the correct shell for the current operating system.

**Testing Methodology**

The tests employ mocking of external dependencies, specifically modules from the `@actions` suite (e.g., `@actions/exec`, `@actions/io`). This allows for controlled testing without relying on actual system commands.  The `process.platform` property is temporarily modified within each test to simulate different operating system environments.  After each test, the original platform is restored.

**Configuration**

The tests utilize a `ProtocolConfig` object to define runtime settings, including:

*   `governance.enabled`:  A boolean indicating whether governance features are enabled.
*   `governance.plugin_whitelist`: An array of strings specifying allowed plugins.
*   `runtime.cli_version`: The desired CLI version.
*   `runtime.node_version`: The required Node.js version.

**Key Components**

*   **RuntimeEnvironment:** This class encapsulates the CLI interaction logic and is the primary component under test.
*   **ProtocolConfig:**  Defines the configuration parameters for the runtime environment.
*   **Mocked Modules:**  The tests rely on mocked versions of `@actions/exec` and `@actions/io` to control the execution environment.

**Test Cases**

The tests include scenarios to verify:

*   Correct shell command prefix for plugin installation on Windows, macOS, and Linux.
*   Handling of plugin installation failures.
*   CLI installation when the CLI is not already present.
*   Handling of CLI installation failures.
*   Skipping CLI installation when the CLI is already installed.
*   Correct shell usage for command execution on Windows and Unix-based systems.

**Dependencies**

*   TypeScript/JavaScript runtime
*   Jest testing framework
*   `@actions/exec` and `@actions/io` modules (mocked during testing)

**Usage**

You can run these tests using the standard Jest command: `jest`.  Ensure that the necessary dependencies are installed.