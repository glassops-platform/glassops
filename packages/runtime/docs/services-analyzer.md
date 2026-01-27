---
type: Documentation
domain: runtime
origin: packages/runtime/src/services/analyzer.ts
last_modified: 2026-01-26
generated: true
source: packages/runtime/src/services/analyzer.ts
generated_at: 2026-01-26T14:20:02.152Z
hash: 9cd5e5c0a77097739e1532e0db567b8b90db41e7a939265fa4f76e2f9ed05883
---

## Analyzer Service Documentation

This document details the functionality of the Analyzer service, responsible for executing the Salesforce Code Analyzer and processing its results.

**Overview**

The Analyzer service provides a standardized way to scan Salesforce projects for code quality issues. It leverages the Salesforce CLI (`sf`) to run the `code-analyzer` command and interprets the output to identify violations against defined rules.  I am designed to be adaptable to different rulesets and project structures.

**Key Components**

*   **Analyzer Class:** The core component that orchestrates the analysis process.
*   **AnalyzerResult Interface:** Defines the structure of the analysis results, including a list of violations and the analyzer's exit code.
*   **Violation Interface:**  Describes a single code violation, including the rule triggered, a descriptive message, severity level, the file name, and the line number.

**Functionality**

1.  **`scan(paths: string[], ruleset?: string): Promise<AnalyzerResult>`**

    This asynchronous function initiates the code analysis process.

    *   **`paths`**:  An array of strings representing the directories or files to be scanned.  For example: `["src", "test"]`.
    *   **`ruleset`**: (Optional) A string specifying the ruleset to enforce during the analysis. If not provided, a default ruleset is used.
    *   **Return Value**: A `Promise` that resolves to an `AnalyzerResult` object containing the identified violations and the analyzer's exit code.

    The `scan` function performs the following steps:

    *   Calls `ensureCompatibility()` to verify the environment.
    *   Constructs the command-line arguments for the `sf code-analyzer run` command.
    *   Executes the command using `@actions/exec`.
    *   Captures the standard output and standard error streams.
    *   Parses the JSON output from the analyzer.
    *   Returns an `AnalyzerResult` object.

2.  **`ensureCompatibility(): Promise<void>`**

    This asynchronous function enforces a policy that requires the use of `sf code-analyzer`.  It currently serves as a placeholder to ensure that users migrate from older scanning tools.  In future versions, this function may include checks to verify the presence of the `sf code-analyzer` command.

3.  **`parseOutput(jsonOutput: string, exitCode: number): AnalyzerResult`**

    This private function parses the JSON output from the `sf code-analyzer` command.

    *   **`jsonOutput`**: The raw JSON string received from the analyzer.
    *   **`exitCode`**: The exit code returned by the analyzer.
    *   **Return Value**: An `AnalyzerResult` object.

    The `parseOutput` function performs the following steps:

    *   Extracts the JSON array from the output string.
    *   Parses the JSON string into a JavaScript object.
    *   Transforms the raw analyzer output into a list of `Violation` objects.
    *   Handles potential parsing errors by logging a warning and returning an empty list of violations.

**Error Handling**

The `scan` function includes error handling to catch exceptions during analyzer execution.  Errors are logged using `@actions/core`, and the error is re-thrown to allow calling functions to handle the failure.  Parsing errors in `parseOutput` are logged as warnings, and an empty result is returned to prevent the process from halting.

**Configuration**

You can configure the analysis by:

*   Specifying the `paths` to scan.
*   Providing a custom `ruleset` to enforce specific code quality rules.