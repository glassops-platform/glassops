---
type: Documentation
domain: runtime
origin: packages/runtime/src/services/analyzer.test.ts
last_modified: 2026-01-26
generated: true
source: packages/runtime/src/services/analyzer.test.ts
generated_at: 2026-01-26T14:19:30.323Z
hash: 97176fb9346eb8b494d860cf7f50ae7dc537e30166de2cbf45f8a916c6a013de
---

## Analyzer Service Documentation

This document details the functionality of the Analyzer service, responsible for executing external analysis tools and processing their output.

**Overview**

The Analyzer service executes a command-line tool ("sf") to scan source code for potential issues. It captures the tool’s output, parses it as JSON, and transforms the findings into a standardized violation format.  The service is designed to handle potential errors during execution and output processing, providing informative feedback through logging.

**Functionality**

The primary function, `scan`, accepts an array of source code paths to analyze and an optional ruleset identifier.  

*   **Execution:** It executes the "sf" command with the provided source paths and ruleset.
*   **Output Handling:** It captures both standard output (stdout) and standard error (stderr) from the executed command.  Stderr is logged for debugging purposes.
*   **JSON Parsing:** It attempts to parse the stdout as a JSON array of violation objects.  The service includes error handling for malformed JSON.
*   **Violation Transformation:**  Parsed violations are transformed into an internal format containing the rule name, description, severity, file path, and line number.
*   **Error Handling:**  The service gracefully handles command execution failures and JSON parsing errors, logging appropriate error or warning messages.
*   **Return Value:** The `scan` function returns an object containing an array of violations and the exit code of the executed command.

**Configuration**

You can specify a ruleset for the analysis by providing a string as the second argument to the `scan` function. If no ruleset is provided, a default ruleset is used.

**Error and Logging**

The service uses the following mechanisms for reporting issues:

*   **Error Logging:**  Critical errors, such as command execution failures, are logged using `core.error`.
*   **Warning Logging:**  Issues like JSON parsing failures are logged using `core.warning`.
*   **Standard Error Capture:**  The stderr output from the executed command is captured and logged to aid in debugging.

**Data Structures**

*   **Violation Object (Internal Format):**
    *   `rule`: The name of the violated rule.
    *   `description`: A description of the violation.
    *   `severity`: An integer representing the severity of the violation.
    *   `file`: The path to the file containing the violation.
    *   `line`: The line number where the violation occurred.

**Dependencies**

*   `@actions/exec`: Used for executing the external command.
*   `@actions/core`: Used for logging errors and warnings.