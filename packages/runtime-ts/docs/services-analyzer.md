---
type: Documentation
domain: runtime-ts
origin: packages/runtime-ts/src/services/analyzer.ts
last_modified: 2026-01-31
generated: true
source: packages/runtime-ts/src/services/analyzer.ts
generated_at: 2026-01-31T10:11:43.204796
hash: f4a83f4fd444af30bcd6ae7b68ee09520db2abfa8fa0347167c5df9af9b446f7
---

## Analyzer Service Documentation

This document details the functionality of the Analyzer service, designed to scan codebases for violations based on configurable rules.

**Overview**

The Analyzer service integrates with the Salesforce CLI (`sf`) to execute code analysis. It provides a standardized interface for running the `sf code-analyzer` command, parsing its output, and reporting violations.  The service prioritizes the use of `sf code-analyzer` and enforces this through compatibility checks.

**Key Components**

*   **AnalyzerResult Interface:** Defines the structure of the analysis results, containing an array of `Violation` objects and the exit code from the analyzer execution.
*   **Violation Interface:** Represents a single code violation, including the rule name, description, severity, file path, and line number.
*   **Analyzer Class:** The core class responsible for executing the analysis and processing the results.

**Functionality**

**1. `scan(paths: string[], ruleset?: string): Promise<AnalyzerResult>`**

This asynchronous function performs the code analysis.

*   **Parameters:**
    *   `paths`: An array of strings representing the directories or files to scan.
    *   `ruleset?`: An optional string specifying the ruleset to enforce during the analysis. If not provided, a default ruleset is used.
*   **Process:**
    1.  Calls `ensureCompatibility()` to verify the environment is configured correctly.
    2.  Constructs the command-line arguments for the `sf code-analyzer run` command.
    3.  Executes the command using `@actions/exec`.
    4.  Captures the standard output and standard error streams.
    5.  Parses the JSON output from the analyzer.
    6.  Returns an `AnalyzerResult` object containing the identified violations and the analyzer's exit code.
*   **Return Value:** A `Promise` that resolves to an `AnalyzerResult` object.

**2. `ensureCompatibility(): Promise<void>`**

This asynchronous function enforces the use of `sf code-analyzer`.

*   **Process:**
    Currently, this function serves as a placeholder for enforcing the policy of using `sf code-analyzer` instead of older tools like `sf scanner`. In a future implementation, it may include checks to verify the presence of `sf code-analyzer` and issue warnings or errors if it is not installed.
*   **Return Value:** A `Promise` that resolves when the compatibility check is complete.

**3. `parseOutput(jsonOutput: string, exitCode: number): AnalyzerResult`**

This private function parses the JSON output from the `sf code-analyzer` command.

*   **Parameters:**
    *   `jsonOutput`: The raw JSON string output from the analyzer.
    *   `exitCode`: The exit code returned by the analyzer execution.
*   **Process:**
    1.  Extracts the JSON array from the output string, handling potential clutter.
    2.  Parses the JSON string into a JavaScript object.
    3.  Maps the raw analyzer output to a simplified array of `Violation` objects. The mapping handles variations in the output format from different versions of the analyzer.
    4.  Handles cases where the output is invalid or empty.
*   **Return Value:** An `AnalyzerResult` object containing the parsed violations and the original exit code.

**Error Handling**

The `scan` function includes error handling to catch exceptions during analyzer execution. Errors are logged using `@actions/core`, and the exception is re-thrown to allow calling functions to handle the failure. The `parseOutput` function also includes error handling to gracefully handle invalid JSON output, logging a warning and returning an empty result.

**Usage Notes**

You should provide the paths to the code you want to analyze as an array of strings to the `scan` function. You can optionally specify a ruleset to customize the analysis. The service is designed to work with the Salesforce CLI and assumes it is installed and configured in the execution environment.