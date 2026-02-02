---
type: Documentation
domain: runtime
origin: packages/runtime/internal/analyzer/analyzer.go
last_modified: 2026-02-01
generated: true
source: packages/runtime/internal/analyzer/analyzer.go
generated_at: 2026-02-01T19:38:40.791650
hash: c5b6ac129f4af1b79959556a81e3ca2f45ebf910e463ec949a678191d18f0eea
---

## Analyzer Package Documentation

This package provides an interface to the Salesforce Code Analyzer, a static analysis tool for Salesforce development. It allows You to scan Salesforce projects for potential issues based on defined rulesets and severity thresholds.

**Package Responsibilities:**

*   Executing the Salesforce Code Analyzer command-line interface (CLI).
*   Parsing the analyzer’s JSON output to extract violation details.
*   Integrating with policy configurations to determine when and how to run the analyzer.
*   Filtering violations based on configured severity thresholds.

**Key Types:**

*   `Result`: Represents the outcome of a code analysis scan. It contains a slice of `Violation` objects and the analyzer’s exit code.
*   `Violation`:  Describes a single code quality issue detected by the analyzer. It includes the rule name, a descriptive message, severity level, the file name, and the line number where the issue occurs.
*   `Analyzer`:  A struct that encapsulates the logic for interacting with the Salesforce Code Analyzer.

**Important Functions:**

*   `New()`:  Returns a new instance of the `Analyzer` struct. This is the standard way to create an analyzer object.
*   `Scan(paths []string, ruleset string) (*Result, error)`: Executes the Salesforce Code Analyzer against the specified file paths.  The `ruleset` parameter allows You to specify a custom ruleset to use during the analysis. The function returns a `Result` object containing the analysis findings and any errors encountered during execution.  The analyzer’s exit code is also captured in the `Result`.
*   `RunIfEnabled(config *policy.Config) error`:  Determines whether to run the analyzer based on the provided policy configuration. If the analyzer is enabled in the configuration, it performs a scan of the current directory (`.`) using the specified ruleset (if any). It then filters the violations based on the configured severity threshold. If critical violations (those meeting the threshold) are found, an error is returned.
*   `EnsureCompatibility() error`:  Currently a placeholder, this function is intended to verify that the environment is correctly configured for using the analyzer. It currently does not perform any checks, but could be extended to validate the Salesforce CLI version or the presence of legacy tools.
*   `parseOutput(jsonOutput string, exitCode int) *Result`:  Parses the JSON output from the Salesforce Code Analyzer and converts it into a `Result` object containing a slice of `Violation` objects. It handles potential errors during JSON unmarshaling and extracts the relevant information from the analyzer’s output.

**Error Handling:**

The package employs standard Go error handling practices. Functions return an `error` value to indicate failure. Errors are often wrapped using `fmt.Errorf` to provide context.  The `gha.Error` function is used to log errors for observability.  The `Scan` function handles non-zero exit codes from the analyzer as expected behavior (indicating violations) and continues processing the output.

**Concurrency:**

This package does not currently employ goroutines or channels for concurrent processing. The analyzer is executed as a single process.

**Design Decisions:**

*   The package is designed to be a wrapper around the existing Salesforce Code Analyzer CLI. This approach leverages the existing functionality and avoids duplicating logic.
*   The use of a policy configuration allows for flexible control over when and how the analyzer is run.
*   The severity threshold filtering provides a mechanism for focusing on the most critical issues.
*   The `parseOutput` function is designed to be resilient to variations in the analyzer’s output format by attempting to locate the JSON array within the overall output string.