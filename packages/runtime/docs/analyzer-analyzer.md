---
type: Documentation
domain: runtime
origin: packages/runtime/internal/analyzer/analyzer.go
last_modified: 2026-01-29
generated: true
source: packages/runtime/internal/analyzer/analyzer.go
generated_at: 2026-01-29T21:20:39.334573
hash: ab8077a3cce90b5885928134a18eb94679791ce20f5d36dac52f7283c6149388
---

## Analyzer Package Documentation

This package wraps the Salesforce Code Analyzer command-line tool, providing a Go interface for running static code analysis on Salesforce projects. It is designed to integrate with continuous integration and continuous delivery pipelines to enforce code quality standards.

**Package Responsibilities:**

*   Executing the Salesforce Code Analyzer.
*   Parsing the analyzer’s JSON output.
*   Providing a structured representation of analysis results (violations and exit code).
*   Ensuring compatibility with the supported analyzer version.

**Key Types:**

*   **`Result`**:  Represents the outcome of a code analysis scan. It contains a slice of `Violation` objects detailing any issues found, and the exit code returned by the analyzer.
*   **`Violation`**:  Describes a single code quality issue detected by the analyzer.  It includes the rule that was violated (`Rule`), a descriptive message (`Description`), the severity of the issue (`Severity`), the file where the issue was found (`File`), and the line number (`Line`).
*   **`Analyzer`**:  A struct that encapsulates the logic for interacting with the Salesforce Code Analyzer.  Currently, it is a simple wrapper, but it allows for future expansion with more complex features.

**Important Functions:**

*   **`New() *Analyzer`**:  Creates and returns a new instance of the `Analyzer` struct.  This is the standard way to obtain an analyzer object.
*   **`Scan(paths []string, ruleset string) (*Result, error)`**:  This is the primary function for performing a code analysis scan.
    *   `paths`: A slice of strings representing the paths to the Salesforce project directories or files to analyze.
    *   `ruleset`: An optional string specifying the name of a ruleset to apply during the analysis. If empty, the default ruleset is used.
    *   Returns: A pointer to a `Result` struct containing the analysis results, and an error if any occurred during execution.
    *   Behavior: This function first calls `EnsureCompatibility` to verify the environment. It then constructs the command-line arguments for the `sf code-analyzer run` command, executes the command using `exec.Command`, and parses the JSON output.  The function handles both successful execution and errors, returning appropriate results.  A non-zero exit code from the analyzer is treated as indicating violations, rather than a fatal error.
*   **`EnsureCompatibility() error`**:  This function checks if the environment is correctly configured for use with the analyzer. Currently, it is a placeholder for future compatibility checks. It is designed to prevent the use of older, deprecated Salesforce CLI commands (like `sf scanner`) and enforce the use of `code-analyzer`.
*   **`parseOutput(jsonOutput string, exitCode int) *Result`**:  This function takes the raw JSON output from the Salesforce Code Analyzer and the analyzer's exit code, and converts it into a structured `Result` object.
    *   `jsonOutput`: The raw JSON string returned by the analyzer.
    *   `exitCode`: The exit code returned by the analyzer process.
    *   Returns: A pointer to a `Result` struct containing the parsed violations and exit code.
    *   Behavior: The function attempts to extract the JSON array from the output string, handling potential clutter. It then unmarshals the JSON into a temporary data structure, and iterates through the results to populate the `Result` struct with `Violation` objects.  Error handling is included for JSON parsing failures.

**Error Handling:**

The package employs standard Go error handling practices. Functions return an `error` value to indicate failure.  The `Scan` function specifically handles `exec.ExitError` to differentiate between analyzer errors (which may indicate violations) and genuine execution failures.  Errors during JSON parsing in `parseOutput` are logged as warnings, and the function continues to return a `Result` with any successfully parsed violations.

**Concurrency:**

This package does not currently employ goroutines or channels for concurrent execution. The `exec.Command` function is used to run the analyzer in a separate process, but the Go code itself remains single-threaded.

**Design Decisions:**

*   **External Dependency:** The package relies on the Salesforce CLI (`sf`) being installed and available in the system’s PATH.
*   **JSON Parsing Robustness:** The `parseOutput` function includes logic to handle potentially malformed JSON output from the analyzer, attempting to extract the relevant data even if the output contains extraneous characters.
*   **Compatibility Enforcement:** The `EnsureCompatibility` function provides a mechanism for enforcing the use of the supported analyzer version and preventing the use of deprecated tools.
*   **Exit Code Handling:** The analyzer returns non-zero exit codes when violations are found. This package interprets these as expected behavior, rather than errors, and includes the exit code in the `Result` for further processing.