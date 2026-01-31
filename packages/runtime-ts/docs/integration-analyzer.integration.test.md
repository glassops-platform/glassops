---
type: Documentation
domain: runtime-ts
origin: packages/runtime-ts/src/integration/analyzer.integration.test.ts
last_modified: 2026-01-29
generated: true
source: packages/runtime-ts/src/integration/analyzer.integration.test.ts
generated_at: 2026-01-29T20:54:43.482993
hash: 0d473ef7c4d829aa4b6073496362792104c26afac17c56c11692582fff9706d5
---

## Analyzer Integration Test Documentation

**Purpose:** This document details the integration tests for the Analyzer service, verifying its interaction with the `sf code-analyzer` command-line tool. These tests ensure the Analyzer correctly constructs commands, parses results, and handles potential errors. This work satisfies requirement BR-003.

**Overview:**

The Analyzer service is designed to integrate with a static code analysis tool – specifically, `sf code-analyzer` – to identify potential issues within a codebase. It provides a consistent interface for scanning code and reporting violations, prioritizing the use of `sf code-analyzer` over alternative tools like `sf scanner`.

**Functionality Tested:**

The integration tests cover the following key aspects:

*   **Command Construction:** Verification that the Analyzer builds the correct `sf code-analyzer` command with the appropriate arguments, including target directories and rulesets.
*   **Result Parsing:** Confirmation that the Analyzer accurately parses JSON output from `sf code-analyzer`, extracting violation details such as rule name, message, severity, and location.
*   **Error Handling:**  Testing the Analyzer’s ability to gracefully handle failures during command execution, such as when the `sf code-analyzer` tool is not found.

**Test Scenarios:**

1.  **Valid Command Construction:** This test confirms that when `analyzer.scan()` is called with a target directory and ruleset, the `sf code-analyzer` command is invoked with the expected parameters.  The test mocks successful execution of the command and verifies the arguments passed to `exec`.

2.  **Violation Parsing:** This test simulates a scenario where `sf code-analyzer` identifies violations. It mocks the command’s output with a sample JSON payload containing violation data. The test then asserts that the `analyzer.scan()` function correctly parses this data and returns a list of violations with the expected properties.

3.  **Command Failure Handling:** This test verifies that the Analyzer handles errors during command execution. It mocks a scenario where the `sf code-analyzer` command fails (e.g., command not found) and confirms that `analyzer.scan()` throws an appropriate error.

**Usage:**

You interact with the Analyzer through its `scan()` method.  You provide an array of target directories and a ruleset name. The `scan()` method executes the `sf code-analyzer` command, parses the results, and returns an object containing any identified violations.

**Output:**

The `scan()` method returns a Promise that resolves to an object containing a `violations` array. Each element in the `violations` array represents a single code violation and includes the following properties:

*   `rule`: The name of the violated rule.
*   `message`: A descriptive message explaining the violation.
*   `severity`: An integer representing the severity of the violation.
*   `line`: The line number where the violation occurred.

**Dependencies:**

*   `@actions/exec`: This package is used to execute the `sf code-analyzer` command. The tests mock this dependency to isolate the Analyzer’s logic.
*   `sf code-analyzer`: The external command-line tool that performs the static code analysis.

**Maintainer Notes:**

I have implemented these tests to ensure the Analyzer’s reliability and maintainability.  We prioritize clear error handling and accurate result parsing.  Future development should continue to focus on expanding test coverage and improving the robustness of the integration with `sf code-analyzer`.