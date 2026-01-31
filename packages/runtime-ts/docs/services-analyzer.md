---
type: Documentation
domain: runtime-ts
origin: packages/runtime-ts/src/services/analyzer.ts
last_modified: 2026-01-31
generated: true
source: packages/runtime-ts/src/services/analyzer.ts
generated_at: 2026-01-31T09:15:14.851700
hash: f4a83f4fd444af30bcd6ae7b68ee09520db2abfa8fa0347167c5df9af9b446f7
---

## Analyzer Service Documentation

This document details the functionality of the Analyzer service, designed to scan codebases for potential issues using the Salesforce Code Analyzer.

**Overview**

The Analyzer service provides a streamlined interface for running the Salesforce Code Analyzer and interpreting its results. It focuses on providing a consistent and reliable method for identifying code quality and security concerns within Salesforce projects.

**Key Features**

*   **Code Scanning:** Executes the Salesforce Code Analyzer against specified file paths or directories.
*   **Ruleset Support:** Allows you to enforce specific coding standards by applying a custom ruleset.
*   **Output Parsing:**  Processes the analyzer’s output, extracting violations and presenting them in a structured format.
*   **Compatibility Enforcement:**  Promotes the use of the `code-analyzer` command, phasing out older scanning methods.

**Interfaces**

*   **AnalyzerResult:** Represents the outcome of a scan.
    *   `violations`: An array of `Violation` objects detailing identified issues.
    *   `exitCode`: The exit code returned by the analyzer process.
*   **Violation:** Describes a single code violation.
    *   `rule`: The name of the rule that was violated.
    *   `description`: A human-readable description of the violation.
    *   `severity`: A numerical representation of the violation’s severity (lower values indicate higher severity).
    *   `file`: The path to the file containing the violation.
    *   `line`: The line number where the violation occurred.

**Usage**

You can interact with the Analyzer service through its `scan` method.

```typescript
const analyzer = new Analyzer();
const result = await analyzer.scan(
  ["src", "test"],
  "my-custom-ruleset" // Optional ruleset
);

// Process the results
if (result.violations.length > 0) {
  console.log("Violations found:");
  for (const violation of result.violations) {
    console.log(
      `${violation.file}:${violation.line} - ${violation.rule}: ${violation.description}`
    );
  }
} else {
  console.log("No violations found.");
}
```

**`scan` Method**

The `scan` method initiates the code analysis process.

*   **Parameters:**
    *   `paths`: An array of strings representing the directories or files to scan.
    *   `ruleset`: (Optional) A string specifying the ruleset to apply during the scan.
*   **Return Value:** An `AnalyzerResult` object containing the scan results.

**Compatibility Policy**

I enforce a policy to encourage the use of the `code-analyzer` command.  The service is designed to work with this command and does not support older scanning tools.

**Output Parsing Details**

The `parseOutput` method handles the extraction of violation data from the analyzer’s JSON output. It is designed to be resilient to variations in the output format, attempting to locate and parse the JSON array even if it is embedded within other text.  If parsing fails, it logs a warning and returns an empty set of violations. The service maps the analyzer’s output to the standardized `Violation` interface.