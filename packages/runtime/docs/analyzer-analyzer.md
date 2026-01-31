---
type: Documentation
domain: runtime
origin: packages/runtime/internal/analyzer/analyzer.go
last_modified: 2026-01-31
generated: true
source: packages/runtime/internal/analyzer/analyzer.go
generated_at: 2026-01-31T09:58:18.932701
hash: ab8077a3cce90b5885928134a18eb94679791ce20f5d36dac52f7283c6149388
---

## Analyzer Package Documentation

This package provides an interface to the Salesforce Code Analyzer command-line tool (`sf code-analyzer`). It allows You to scan Salesforce projects for code quality issues and security vulnerabilities.

**Package Responsibilities:**

The primary responsibility of this package is to execute the `sf code-analyzer` tool, parse its output, and present the results in a structured format. It handles the complexities of running the external process and converting the JSON output into Go data structures.

**Key Types:**

*   **`Result`**: Represents the overall outcome of a code analysis scan. It contains a slice of `Violation` objects and the exit code of the analyzer process.
    ```go
    type Result struct {
    	Violations []Violation
    	ExitCode   int
    }
    ```

*   **`Violation`**: Represents a single code quality or security issue found during the analysis. It includes details such as the rule that was violated, a description of the issue, its severity, the file name, and the line number.
    ```go
    type Violation struct {
    	Rule        string `json:"rule"`
    	Description string `json:"description"`
    	Severity    int    `json:"severity"`
    	File        string `json:"file"`
    	Line        int    `json:"line"`
    }
    ```

*   **`Analyzer`**:  A struct that encapsulates the logic for interacting with the `sf code-analyzer` tool.
    ```go
    type Analyzer struct{}
    ```

**Important Functions:**

*   **`New()`**:  A constructor function that returns a new instance of the `Analyzer` struct.
    ```go
    func New() *Analyzer {
    	return &Analyzer{}
    }
    ```

*   **`Scan(paths []string, ruleset string) (*Result, error)`**: This is the main function for performing a code analysis scan.
    *   `paths`: A slice of strings representing the file paths or directories to be scanned.
    *   `ruleset`: An optional string specifying the name of a ruleset to use for the analysis. If empty, the default ruleset is used.
    *   Returns: A pointer to a `Result` struct containing the scan results, or an error if the scan failed.
    This function executes the `sf code-analyzer run` command with the provided arguments, captures the output, and parses it to extract violations. It handles potential errors during command execution and JSON parsing.

*   **`EnsureCompatibility()`**: This function checks if the environment is correctly configured for using the `sf code-analyzer` tool. Currently, it is a placeholder for future compatibility checks. It is designed to prevent the use of older, deprecated scanning tools.
    ```go
    func (a *Analyzer) EnsureCompatibility() error {
    	// Placeholder for opinionated policy enforcement.
    	return nil
    }
    ```

*   **`parseOutput(jsonOutput string, exitCode int) *Result`**: This function parses the JSON output from the `sf code-analyzer` tool and converts it into a `Result` struct. It handles cases where the JSON output might be incomplete or contain extraneous data.
    ```go
    func (a *Analyzer) parseOutput(jsonOutput string, exitCode int) *Result {
    	// ... parsing logic ...
    }
    ```

**Error Handling:**

The `Scan` function handles errors in the following ways:

*   If `EnsureCompatibility` returns an error, the `Scan` function immediately returns that error.
*   If the `sf code-analyzer` command fails to execute, the `Scan` function logs the error using `gha.Error` and returns it.
*   If the JSON output from the analyzer cannot be parsed, the `Scan` function logs a warning using `gha.Warning` and returns a `Result` struct with an empty slice of violations.
*   The exit code of the `sf code-analyzer` command is captured and included in the `Result` struct, even if the command completes successfully (as the analyzer uses non-zero exit codes to indicate violations).

**Concurrency:**

This package does not currently employ goroutines or channels for concurrent processing. The `Scan` function executes the `sf code-analyzer` command synchronously. Future versions might introduce concurrency to improve performance.

**Design Decisions:**

*   **External Dependency:** The package relies on the `sf code-analyzer` command-line tool, which must be installed and configured separately. This allows the package to leverage the existing functionality and expertise of the Salesforce Code Analyzer.
*   **JSON Parsing:** The package uses the `encoding/json` package to parse the JSON output from the analyzer. This provides a flexible and reliable way to extract the violation data.
*   **Error Logging:** The package uses the `gha.Error` and `gha.Warning` functions to log errors and warnings. This provides a consistent way to report issues to the user.