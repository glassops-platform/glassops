---
type: Documentation
domain: runtime-ts
origin: packages/runtime-ts/src/services/analyzer.test.ts
last_modified: 2026-01-31
generated: true
source: packages/runtime-ts/src/services/analyzer.test.ts
generated_at: 2026-01-31T10:11:22.078526
hash: 7d2ddbbab6796de15d6654bccb98337c6487a3e9b2166dbec435abeccbfeea3c
---

## Analyzer Service Documentation

This document details the functionality of the Analyzer service. This service is designed to execute an external analysis tool ("sf") and process its output to identify code violations.

**Purpose**

The Analyzer service provides a standardized way to integrate with static analysis tools. It handles execution, output parsing, error handling, and result formatting. You can use this service to scan codebases for potential issues based on defined rulesets.

**Functionality**

The core function of the Analyzer service is the `scan` method.

*   **`scan(directories: string[], ruleset?: string): Promise<AnalysisResult>`**

    This method executes the external analysis tool against the specified directories. 

    *   `directories`: An array of strings representing the directories to scan.
    *   `ruleset` (optional): A string specifying the ruleset to apply during the analysis. If not provided, a default ruleset is used.
    *   Returns: A Promise that resolves to an `AnalysisResult` object containing the scan results.

**AnalysisResult Interface**

The `scan` method returns an object with the following properties:

*   `violations`: An array of `Violation` objects representing the identified code violations.
*   `exitCode`: The exit code returned by the external analysis tool.

**Violation Interface**

Each `Violation` object in the `violations` array has the following properties:

*   `rule`: The name of the rule that was violated.
*   `description`: A description of the violation.
*   `severity`: An integer representing the severity of the violation.
*   `file`: The path to the file containing the violation.
*   `line`: The line number where the violation occurred.

**Error Handling**

The Analyzer service includes robust error handling:

*   **Execution Failures:** If the external analysis tool fails to execute, the `scan` method will reject with the error message from the tool.  An error message is also logged to the core output.
*   **JSON Parsing Errors:** If the output from the external analysis tool is not valid JSON, a warning is logged, and an empty array of violations is returned.
*   **Mixed Output:** The service is designed to handle cases where the output from the external analysis tool contains both JSON data and other text (e.g., debug logs). It attempts to parse the JSON data and ignores any non-JSON content.

**Command Line Arguments**

The service passes arguments to the external analysis tool. Specifically:

*   If a `ruleset` is provided to the `scan` method, it is passed to the tool as a `--ruleset` argument.

**Debugging**

The service captures both standard output (stdout) and standard error (stderr) from the external analysis tool.  The stderr is captured to aid in debugging issues with the analysis process.