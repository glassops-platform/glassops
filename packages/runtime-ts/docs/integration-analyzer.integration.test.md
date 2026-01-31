---
type: Documentation
domain: runtime-ts
origin: packages/runtime-ts/src/integration/analyzer.integration.test.ts
last_modified: 2026-01-31
generated: true
source: packages/runtime-ts/src/integration/analyzer.integration.test.ts
generated_at: 2026-01-31T09:11:33.280049
hash: 0d473ef7c4d829aa4b6073496362792104c26afac17c56c11692582fff9706d5
---

## Analyzer Integration Test Documentation

**Purpose:** This document details the integration tests for the Analyzer service, verifying its interaction with the `sf code-analyzer` command-line tool. These tests ensure the Analyzer correctly constructs commands, parses results, and handles potential errors. This work satisfies requirement BR-003.

**Overview:**

The Analyzer service is designed to integrate with a static analysis tool – specifically, `sf code-analyzer` – to identify code quality issues. It provides a consistent interface for scanning code and reporting violations, prioritizing the use of `sf code-analyzer` over alternative tools like `sf scanner`.

**Functionality Tested:**

The integration tests cover the following key aspects:

*   **Command Construction:** Verification that the Analyzer builds the correct `sf code-analyzer` command with the appropriate arguments for target directories and rulesets.
*   **Result Parsing:** Confirmation that the Analyzer accurately parses JSON output from `sf code-analyzer`, extracting violation details such as rule name, message, severity, and location.
*   **Error Handling:**  Testing the Analyzer’s ability to gracefully handle failures during command execution, such as when the `sf` command is not found.

**Test Scenarios:**

1.  **Valid Command Construction:** This test confirms that when `analyzer.scan()` is called with a target directory and ruleset, the `sf code-analyzer` command is invoked with the expected parameters.  The test mocks successful execution of the command and verifies the arguments passed to `exec.exec`.

2.  **Violation Parsing:** This test simulates a scenario where `sf code-analyzer` returns a JSON payload containing violation data. The Analyzer is expected to parse this data and populate a `violations` array with the extracted information. The test asserts that the parsed violations have the correct properties.

3.  **Command Failure Handling:** This test verifies that if `sf code-analyzer` fails to execute (e.g., due to a missing command), the Analyzer correctly propagates the error to the calling code. The test mocks a rejected promise from `exec.exec` and asserts that `analyzer.scan()` throws an exception with the expected error message.

**Usage:**

You interact with the Analyzer through its `scan()` method.  You provide an array of target directories and a ruleset name. The `scan()` method executes the `sf code-analyzer` command and returns a result object containing any detected violations.

**Dependencies:**

*   `@actions/exec`: This package is used to execute shell commands. The tests mock this dependency to isolate the Analyzer’s logic.
*   `sf code-analyzer`: The external command-line tool that performs the static analysis.

**Output:**

The `scan()` method returns a promise that resolves to an object containing a `violations` array. Each element in the `violations` array represents a detected code quality issue and includes the following properties:

*   `rule`: The name of the violated rule.
*   `message`: A description of the violation.
*   `severity`: The severity level of the violation.
*   `line`: The line number where the violation occurred.

In case of an error during command execution, the `scan()` method rejects with an error object.