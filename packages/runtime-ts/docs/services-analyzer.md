---
type: Documentation
domain: runtime-ts
origin: packages/runtime-ts/src/services/analyzer.ts
last_modified: 2026-01-29
generated: true
source: packages/runtime-ts/src/services/analyzer.ts
generated_at: 2026-01-29T20:58:16.233950
hash: f4a83f4fd444af30bcd6ae7b68ee09520db2abfa8fa0347167c5df9af9b446f7
---

## Analyzer Service Documentation

This document details the functionality of the Analyzer service, designed to scan codebases for potential issues using the Salesforce Code Analyzer.

**Overview**

The Analyzer service provides a streamlined interface for running the Salesforce Code Analyzer and interpreting its results. It focuses on providing a consistent and reliable method for identifying code quality and security concerns within Salesforce projects.

**Functionality**

The core function of this service is the `scan` method. You can initiate a scan by providing the following:

*   **Paths:** An array of strings representing the directories or files to be analyzed.
*   **Ruleset (Optional):** A string specifying a custom ruleset to enforce during the analysis. If omitted, the default ruleset is used.

**Workflow**

1.  **Compatibility Check:** The `ensureCompatibility` function verifies the environment is configured to use the `code-analyzer` command. It enforces the use of `code-analyzer` over the older `scanner` command.
2.  **Analyzer Execution:** The `scan` method executes the Salesforce Code Analyzer using the provided paths and ruleset.  The analyzer’s output is captured.
3.  **Output Parsing:** The `parseOutput` method processes the analyzer’s JSON output, extracting individual violations. It handles potential inconsistencies in the output format.
4.  **Result Delivery:** The service returns an `AnalyzerResult` object containing an array of `Violation` objects and the analyzer’s exit code.

**Data Structures**

*   **AnalyzerResult:**
    *   `violations`: An array of `Violation` objects representing identified issues.
    *   `exitCode`: The exit code returned by the Salesforce Code Analyzer.

*   **Violation:**
    *   `rule`: The name of the rule that was violated.
    *   `description`: A human-readable description of the violation.
    *   `severity`: A numerical representation of the violation’s severity (lower values indicate higher severity).
    *   `file`: The path to the file containing the violation.
    *   `line`: The line number within the file where the violation occurred.

**Error Handling**

If the Salesforce Code Analyzer execution fails, the service logs an error message and re-throws the exception. If parsing the analyzer’s output fails, a warning is logged, and an empty `AnalyzerResult` is returned.

**Important Considerations**

*   The service is designed to work with the Salesforce CLI (`sf`) and assumes that the `code-analyzer` plugin is installed.
*   The structure of the analyzer’s output may vary depending on the version of the Salesforce Code Analyzer. The `parseOutput` method includes a generic mapping to accommodate common variations.
*   The analyzer may return a non-zero exit code even when violations are found. The service handles this by ignoring the return code and focusing on parsing the output.