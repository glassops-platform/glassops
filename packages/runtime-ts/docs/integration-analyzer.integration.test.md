---
type: Documentation
domain: runtime-ts
origin: packages/runtime-ts/src/integration/analyzer.integration.test.ts
last_modified: 2026-01-31
generated: true
source: packages/runtime-ts/src/integration/analyzer.integration.test.ts
generated_at: 2026-01-31T10:07:46.254162
hash: 0d473ef7c4d829aa4b6073496362792104c26afac17c56c11692582fff9706d5
---

## Analyzer Integration Test Documentation

**Purpose:** This document details the integration tests for the Analyzer service, verifying its interaction with the `sf code-analyzer` tool. These tests ensure the Analyzer correctly constructs commands, parses results, and handles potential errors.

**Scope:** This document covers the functionality and behavior validated by the integration tests. It is intended for developers, testers, and anyone interested in understanding how the Analyzer service integrates with external code analysis tools.

**Overview:**

The Analyzer service provides a consistent interface for running code analysis against a project. It prioritizes the use of `sf code-analyzer` for code quality checks. The integration tests focus on verifying this interaction, ensuring accurate command construction, result parsing, and error handling.

**Functionality Tested:**

*   **Command Construction:** The tests confirm that the Analyzer builds the correct `sf code-analyzer` command based on provided parameters (target directory and ruleset). The expected command includes the necessary arguments for specifying the target source code and the desired ruleset.
*   **Result Parsing:** The Analyzer parses the JSON output from `sf code-analyzer` to identify code violations. Tests verify that violations are correctly extracted, including rule name, message, severity, and line number.
*   **Error Handling:** The tests validate that the Analyzer gracefully handles failures during the execution of the `sf code-analyzer` command. Specifically, it checks that errors are caught and re-thrown with informative messages.

**Test Cases:**

1.  **Valid Command Construction:**
    *   **Input:** Target directory: `src`, Ruleset: `Complexity`.
    *   **Expected Outcome:** The `sf code-analyzer` command is invoked with the correct arguments: `sf code-analyzer run --target src --ruleset Complexity`.  The test mocks a successful execution of the command, returning an empty array (no violations).
2.  **Violation Parsing:**
    *   **Input:** Target directory: `src`, Mock JSON output containing a single violation (eslint, rule: `no-any`, file: `src/bad.ts`, line: 10).
    *   **Expected Outcome:** The Analyzer correctly parses the JSON output and identifies one violation. The parsed violation’s rule is verified to be `no-any`. The test mocks a non-zero exit code to simulate a violation being found.
3.  **Command Failure Handling:**
    *   **Input:** Target directory: `src`, Mocked `sf code-analyzer` command throws an error ("Command not found: sf").
    *   **Expected Outcome:** The Analyzer catches the error and re-throws it with the same message ("Command not found"). This ensures that failures in the external tool are propagated to the calling code.

**Dependencies:**

*   `@actions/exec`: This package is used to execute shell commands. The tests mock this dependency to control the execution environment.
*   `sf code-analyzer`: The external code analysis tool that the Analyzer integrates with.

**Usage:**

You do not directly interact with these integration tests. They are part of the internal development process and are executed automatically as part of the build pipeline.

**Maintainance:**

I will update these tests as the `sf code-analyzer` tool evolves or as new features are added to the Analyzer service. We aim to keep these tests comprehensive and reliable to ensure the continued quality of the integration.