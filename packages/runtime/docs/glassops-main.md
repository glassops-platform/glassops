---
type: Documentation
domain: runtime
origin: packages/runtime/cmd/glassops/main.go
last_modified: 2026-01-31
generated: true
source: packages/runtime/cmd/glassops/main.go
generated_at: 2026-01-31T09:57:50.306116
hash: c16e6c382845f65a8d2af02a90b0d828dc4354bf2a489904917c35b59490031c
---

## GlassOps Runtime Documentation

This document details the functionality and design of the GlassOps Runtime, a tool designed to provide governance and automation for Salesforce deployments within a CI/CD pipeline. It is intended for both technical users (developers, platform engineers) and non-technical stakeholders (managers, security officers).

**Package Purpose and Responsibilities**

The `main` package serves as the entry point for the GlassOps Runtime. It orchestrates a series of checks and processes to ensure deployments adhere to defined policies and standards. The runtime focuses on validating the environment, enforcing policies, bootstrapping necessary tools, authenticating with Salesforce, and generating a deployment contract.

**Key Types and Interfaces**

The runtime leverages several internal packages, but does not expose any public types or interfaces. Key internal components include:

*   `analyzer.CodeAnalyzer`: Performs static code analysis to identify potential issues.
*   `contract.DeploymentContract`: Represents the agreement between the deployment process and governance rules. It captures details about the deployment, quality checks, and audit information.
*   `gha.GitHubActions`: Provides an abstraction layer for interacting with the GitHub Actions environment, handling inputs, outputs, and status reporting.
*   `policy.PolicyEngine`: Loads and evaluates governance policies.
*   `services.RuntimeEnvironment`: Manages the Salesforce CLI and related tools.
*   `services.IdentityResolver`: Handles authentication with Salesforce.
*   `telemetry.TelemetryService`: Collects and exports runtime metrics.

**Important Functions and Their Behavior**

*   `main()`: The primary entry point. It initializes the runtime context and calls the `run()` function. Handles top-level error reporting to GitHub Actions.
*   `run(ctx context.Context) error`:  This function contains the core logic of the runtime. It performs the following stages:
    *   **Environment Validation:** Checks for the presence of required environment variables (`GITHUB_WORKSPACE`, `GITHUB_ACTOR`, `GITHUB_REPOSITORY`).
    *   **Input Validation:** Validates required inputs provided via GitHub Actions inputs (`client_id`, `jwt_key`, `username`).  It also validates the format of the JWT key and the Salesforce instance URL.
    *   **Resource Limits:** Implements a timeout mechanism to prevent runaway executions. A goroutine is launched to terminate the process if it exceeds a predefined time limit (30 minutes).
    *   **Data Integrity Checks:** Validates the context of the execution (e.g., pull request vs. direct commit) and issues warnings for potentially less secure scenarios (forked repositories). It also validates the repository format.
    *   **Initialization:** Generates a unique runtime ID and initializes OpenTelemetry for tracing and monitoring.
    *   **Policy Evaluation:** Loads and evaluates governance policies using the `policy.PolicyEngine`. This includes static code analysis (if enabled) and checks for active freeze windows.
    *   **Bootstrap:** Installs the Salesforce CLI using the `services.RuntimeEnvironment`.  It also supports the installation of plugins.
    *   **Identity Resolution:** Authenticates with Salesforce using the provided credentials via the `services.IdentityResolver`. Authentication can be skipped if configured.
    *   **Contract Generation:** Creates a deployment contract (`contract.DeploymentContract`) summarizing the deployment details, quality checks (test results, code coverage), and audit information. The contract is written to a file.
    *   **Output:** Sets GitHub Actions outputs with relevant information (runtime ID, org ID, contract path, lock status).
*   `isValidURL(s string) bool`: Checks if a given string is a valid URL.
*   `generateUUID() string`: Generates a Universally Unique Identifier (UUID).
*   `splitAndTrim(s, sep string) []string`: Splits a string by a separator and trims whitespace from each resulting part.
*   `parseFloat(s string, defaultVal float64) float64`: Parses a string to a float64, returning a default value if parsing fails.

**Error Handling Patterns**

The runtime employs a consistent error handling approach. Functions return an `error` value. Errors are checked immediately, and if an error occurs, the function returns, propagating the error up the call stack.  Errors are often wrapped using `fmt.Errorf("%w", err)` to provide additional context without losing the original error information.  Critical errors are reported to GitHub Actions using `gha.SetFailed()`, which sets the workflow status to failed.  Non-critical errors or warnings are reported using `gha.Warning()`.

**Concurrency Patterns**

The runtime uses a single goroutine for the timeout mechanism. This goroutine sleeps for a predefined duration and checks if the execution time has exceeded the limit. This prevents long-running or stalled deployments.

**Notable Design Decisions**

*   **GitHub Actions Integration:** The runtime is specifically designed to operate within a GitHub Actions environment, leveraging its inputs, outputs, and status reporting mechanisms.
*   **Policy-Driven Governance:** The runtime enforces governance policies through a dedicated `policy.PolicyEngine`, allowing for flexible and configurable rules.
*   **Deployment Contract:** The generation of a deployment contract provides a clear and auditable record of the deployment process and its adherence to governance standards.
*   **Modular Design:** The runtime is structured into distinct packages (analyzer, contract, policy, services) to promote code organization and maintainability.
*   **Telemetry Integration:** OpenTelemetry is integrated to provide insights into runtime performance and behavior.