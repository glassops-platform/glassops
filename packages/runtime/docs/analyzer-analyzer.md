---
type: Documentation
domain: runtime
last_modified: 2026-02-02
generated: true
source: packages/runtime/internal/analyzer/analyzer.go
generated_at: 2026-02-02T22:34:49.432882
hash: c5b6ac129f4af1b79959556a81e3ca2f45ebf910e463ec949a678191d18f0eea
---

## Analyzer Package Documentation

This package provides an interface to the Salesforce Code Analyzer, a tool for static code analysis of Salesforce projects. We aim to integrate this analysis into a broader governance and compliance framework.

**Package Responsibilities:**

The `analyzer` package is responsible for:

-   Executing the Salesforce Code Analyzer command-line interface (CLI).
-   Parsing the analyzer’s JSON output to extract code violations.
-   Filtering violations based on configurable severity thresholds.
-   Enforcing compatibility by ensuring the correct analyzer version is in use.
-   Integrating with policy configurations to determine when and how to run the analysis.

**Key Types:**

-   `Result`: Represents the outcome of a code analysis scan. It contains a slice of `Violation` objects and the analyzer’s exit code.
    ```go
    type Result struct {
    	Violations []Violation
    	ExitCode   int
    }
    ```
-   `Violation`: Represents a single code violation found during analysis. It includes details such as the rule name, description, severity, file name, and line number.
    ```go
    type Violation struct {
    	Rule        string `json:"rule"`
    	Description string `json:"description"`
    	Severity    int    `json:"severity"`
    	File        string `json:"file"`
    	Line        int    `json:"line"`
    }
    ```
-   `Analyzer`:  A struct that encapsulates the logic for interacting with the Salesforce Code Analyzer. It provides methods for running scans and parsing results.
    ```go
    type Analyzer struct{}
    ```

**Important Functions:**

-   `New()`:  A constructor function that returns a new instance of the `Analyzer` struct.
    ```go
    func New() *Analyzer {
    	return &Analyzer{}
    }
    ```
-   `Scan(paths []string, ruleset string) (*Result, error)`: Executes the Salesforce Code Analyzer on the specified file paths.  You provide a list of paths to scan and an optional ruleset to apply. The function returns a `Result` object containing any violations found and the analyzer’s exit code.  It handles potential errors during execution and parsing.
    ```go
    func (a *Analyzer) Scan(paths []string, ruleset string) (*Result, error) {
    	// ... implementation details ...
    }
    ```
-   `RunIfEnabled(config *policy.Config) error`:  Determines whether to run the analyzer based on the provided policy configuration. If the analyzer is enabled in the configuration, it executes a scan on the current directory (`.`) using the specified ruleset (if any). It filters violations based on the configured severity threshold and returns an error if critical violations are found.
    ```go
    func (a *Analyzer) RunIfEnabled(config *policy.Config) error {
    	// ... implementation details ...
    }
    ```
-   `EnsureCompatibility() error`:  Performs checks to ensure the environment is correctly configured for using the analyzer. Currently, this function serves as a placeholder for future compatibility enforcement, such as verifying the correct version of the Salesforce CLI is installed.
    ```go
    func (a *Analyzer) EnsureCompatibility() error {
    	// ... implementation details ...
    }
    ```
-   `parseOutput(jsonOutput string, exitCode int) *Result`: Parses the JSON output from the Salesforce Code Analyzer and extracts the code violations. It handles cases where the output may contain extraneous characters and gracefully handles parsing errors.
    ```go
    func (a *Analyzer) parseOutput(jsonOutput string, exitCode int) *Result {
    	// ... implementation details ...
    }
    ```

**Error Handling:**

The package employs standard Go error handling practices. Functions return an `error` value to indicate failure. Errors are often wrapped using `fmt.Errorf` to provide context.  The `gha.Error` function is used to log errors for visibility within the broader platform.

**Concurrency:**

This package does not currently employ goroutines or channels for concurrent execution. The Salesforce Code Analyzer is executed as a synchronous process.

**Design Decisions:**

-   We chose to wrap the Salesforce Code Analyzer CLI rather than reimplementing its functionality to leverage its existing capabilities and avoid maintenance overhead.
-   The `RunIfEnabled` function provides a convenient way to integrate the analyzer into a policy-driven workflow.
-   The `parseOutput` function is designed to be resilient to variations in the analyzer’s output format.
-   The package is designed to be extensible, allowing for future additions such as support for multiple rulesets and more sophisticated violation filtering.