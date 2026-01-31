---
type: Documentation
domain: runtime
origin: packages/runtime/internal/analyzer/analyzer.go
last_modified: 2026-01-31
generated: true
source: packages/runtime/internal/analyzer/analyzer.go
generated_at: 2026-01-31T09:02:33.697415
hash: ab8077a3cce90b5885928134a18eb94679791ce20f5d36dac52f7283c6149388
---

## Analyzer Package Documentation

This package provides an interface to the Salesforce Code Analyzer command-line tool. It allows You to scan Salesforce projects for code quality issues and security vulnerabilities. We aim to provide a robust and reliable way to integrate code analysis into automated workflows.

**Key Types:**

*   **Result:** Represents the outcome of a code analysis scan. It contains a slice of `Violation` objects detailing any issues found, and the exit code from the underlying analyzer tool.
*   **Violation:**  Describes a single code quality or security issue identified by the analyzer. It includes the rule that was violated (`Rule`), a descriptive message (`Description`), the severity of the issue (`Severity`), the file where the issue was found (`File`), and the line number (`Line`).
*   **Analyzer:**  This struct encapsulates the logic for interacting with the Salesforce Code Analyzer. It currently serves as a container for methods, but is designed to accommodate more complex logic or configuration in the future.

**Important Functions:**

*   **New():**  A constructor function that returns a new instance of the `Analyzer` struct.
*   **Scan(paths []string, ruleset string) (*Result, error):** This is the primary function for performing a code analysis scan.
    *   `paths`: A slice of strings representing the file paths or directories to be scanned.
    *   `ruleset`: An optional string specifying the ruleset to use for the analysis. If empty, the default ruleset is applied.
    *   It executes the `sf code-analyzer run` command with the provided paths and ruleset.
    *   It captures the output of the command and parses it to extract violations.
    *   It returns a `Result` object containing the analysis findings and an error, if any occurred during execution.
*   **EnsureCompatibility():**  This function currently serves as a placeholder for compatibility checks. It is intended to enforce policies and prevent the use of older, deprecated Salesforce CLI commands (like `sf scanner`). Currently, it does not perform any checks but is designed for future expansion.
*   **parseOutput(jsonOutput string, exitCode int) *Result:** This function takes the raw JSON output from the Salesforce Code Analyzer and converts it into a structured `Result` object.
    *   It extracts the JSON array from the potentially cluttered output.
    *   It unmarshals the JSON into a temporary data structure.
    *   It transforms the temporary data structure into a slice of `Violation` objects and populates the `Result` object.

**Error Handling:**

*   The `Scan` function returns an error if the `sf code-analyzer` command fails to execute.
*   The `parseOutput` function logs a warning if it fails to parse the JSON output, but does not return an error. This allows the scan to continue even if the output is malformed, providing partial results.
*   We use the `github.com/glassops-platform/glassops/packages/runtime/internal/gha` package for logging errors and warnings, which is intended for integration with GitHub Actions workflows.

**Concurrency:**

This package does not currently employ goroutines or channels for concurrent processing. The `Scan` function executes the `sf code-analyzer` command synchronously.  Future versions may incorporate concurrency to improve performance, particularly when scanning large projects.

**Design Decisions:**

*   **External Dependency:** The package relies on the Salesforce CLI (`sf`) being installed and configured in the environment.
*   **JSON Parsing:** The package assumes the Salesforce Code Analyzer outputs JSON. The `parseOutput` function is designed to handle potential variations in the JSON output format by extracting the relevant JSON array.
*   **Exit Code Handling:** The Salesforce Code Analyzer returns a non-zero exit code when violations are found. We treat this as a normal condition and continue processing the output to extract the violations.
*   **Compatibility Enforcement:** The `EnsureCompatibility` function is a key part of our strategy to migrate users away from older, less effective analysis tools.