---
type: Documentation
domain: runtime
origin: packages/runtime/src/integration/analyzer.integration.test.ts
last_modified: 2026-01-26
generated: true
source: packages/runtime/src/integration/analyzer.integration.test.ts
generated_at: 2026-01-26T14:13:57.471Z
hash: db284f88c154c7cb43d17411616b41f99a32a2c893d9a0f78fe872a28c5e1443
---

## Analyzer Integration Test Documentation

**Purpose:** This document details the integration tests for the Code Analyzer service. These tests verify the correct operation of the analyzer, ensuring it interacts properly with the `sf code-analyzer` command-line tool and processes its output. This work satisfies requirement BR-003.

**Overview:**

The Analyzer service is designed to integrate with the `sf code-analyzer` tool to identify code quality issues. It executes the analyzer against specified targets and rulesets, parses the results, and provides a structured representation of any violations found.  I prioritize the use of `sf code-analyzer` over `sf scanner`.

**Functionality Tested:**

The integration tests cover the following key areas:

*   **Command Construction:** Verification that the correct `sf code-analyzer` command is constructed with the appropriate arguments (target directories and rulesets).
*   **Output Parsing:**  Confirmation that the service correctly parses the JSON output from `sf code-analyzer`, extracting violation details such as rule name, message, severity, and location.
*   **Error Handling:**  Ensuring the service handles failures during command execution gracefully, propagating errors to the calling code.

**Test Details:**

1.  **Valid Command Construction:**
    *   **Input:** A target directory ("src") and a ruleset ("Complexity").
    *   **Behavior:** The test mocks the execution of `sf code-analyzer` to simulate a successful run with no violations. It then asserts that `sf code-analyzer` was called with the expected command and arguments.
    *   **Outcome:** Confirms the service builds the correct command line invocation.

2.  **Violation Parsing:**
    *   **Input:** A target directory ("src") and mocked JSON output representing a code violation (specifically, an "any" type usage in `src/bad.ts`).
    *   **Behavior:** The test mocks `sf code-analyzer` to return the sample JSON output. It then calls the analyzer and asserts that the resulting violation list contains the expected violation details (rule name "no-any").
    *   **Outcome:** Validates the service can correctly interpret the analyzer’s output and extract relevant information.

3.  **Command Failure Handling:**
    *   **Input:** A target directory ("src").
    *   **Behavior:** The test mocks `sf code-analyzer` to throw an error (simulating a command not found). It then asserts that calling the analyzer results in the same error being thrown.
    *   **Outcome:**  Verifies the service handles errors during command execution and provides informative error messages.

**Dependencies:**

*   `@actions/exec`:  Used for executing the `sf code-analyzer` command. This is mocked during testing.

**Configuration:**

You can configure the target directories and rulesets passed to the analyzer through its `scan` method. For example: `await analyzer.scan(["src", "test"], "Security")`.

**Output:**

The `scan` method returns a promise that resolves to an object containing a `violations` array. Each element in the array represents a code violation and includes properties such as `rule`, `message`, `severity`, and `line`.  If an error occurs during execution, the promise will reject with an error object.