---
type: Documentation
domain: runtime-ts
origin: packages/runtime-ts/src/services/analyzer.test.ts
last_modified: 2026-01-31
generated: true
source: packages/runtime-ts/src/services/analyzer.test.ts
generated_at: 2026-01-31T09:14:52.224123
hash: 7d2ddbbab6796de15d6654bccb98337c6487a3e9b2166dbec435abeccbfeea3c
---

## Analyzer Service Documentation

This document details the functionality and behavior of the Analyzer service. This service is designed to execute an external static analysis tool ("sf") and process its output to identify code quality and security violations.

**Purpose**

The Analyzer service provides a standardized way to integrate static analysis into a workflow. It handles execution of the analysis tool, parsing of the results, and reporting of violations in a consistent format.

**Functionality**

The core function, `scan`, accepts an array of file paths to analyze and an optional ruleset identifier. It performs the following steps:

1.  **Execution:** Executes the "sf" command with the provided file paths and ruleset.
2.  **Output Handling:** Captures both standard output (stdout) and standard error (stderr) from the executed command.  Stderr is logged for debugging purposes.
3.  **Parsing:** Attempts to parse the stdout as JSON.  The service is resilient to malformed JSON and mixed content (e.g., debug logs interspersed with JSON).  If parsing fails, a warning is logged, and an empty result is returned.
4.  **Violation Mapping:**  Transforms violations from the external tool’s format into an internal, standardized format.
5.  **Result Reporting:** Returns a result object containing an array of violations and the exit code of the executed command.

**Result Format**

The `scan` function returns an object with the following properties:

*   `violations`: An array of violation objects. Each violation object has the following properties:
    *   `rule`: The name of the rule that was violated.
    *   `description`: A description of the violation.
    *   `severity`: An integer representing the severity of the violation.
    *   `file`: The path to the file where the violation occurred.
    *   `line`: The line number where the violation occurred.
*   `exitCode`: The exit code returned by the "sf" command.

**Error Handling**

*   If the "sf" command fails to execute, an error is thrown.
*   If the stdout cannot be parsed as valid JSON, a warning is logged, and an empty list of violations is returned.
*   The service captures and logs stderr from the "sf" command to assist with debugging.

**Configuration**

The Analyzer service currently uses a hardcoded command name, "sf".  The ruleset argument is configurable via the `scan` function.

**Usage**

You can call the `scan` function with an array of file paths:

```typescript
const result = await analyzer.scan(["src"]);
```

You can also specify a ruleset:

```typescript
const result = await analyzer.scan(["src"], "Security");
```

**Dependencies**

*   `@actions/exec`: Used for executing external commands.
*   `@actions/core`: Used for logging errors and warnings.