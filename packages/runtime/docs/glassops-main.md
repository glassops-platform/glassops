---
type: Documentation
domain: runtime
origin: packages/runtime/cmd/glassops/main.go
last_modified: 2026-01-29
generated: true
source: packages/runtime/cmd/glassops/main.go
generated_at: 2026-01-29T21:20:12.222419
hash: 47e48ad51fa034d9267a0ae4674361570d2ddf748deee350cf8d56e8c8795084
---

## GlassOps Runtime Documentation

This document details the functionality and design of the GlassOps Runtime, a tool designed to provide governance and automation for Salesforce deployments within a CI/CD pipeline. It is intended for both technical users (developers, platform engineers) and non-technical stakeholders (security officers, release managers).

**Package Purpose:**

The `main` package serves as the entry point for the GlassOps Runtime. It orchestrates a series of checks and processes to ensure deployments adhere to defined policies and standards. The runtime integrates with GitHub Actions to provide a secure and auditable deployment process.

**Key Types and Interfaces:**

*   **config (internal/policy):** Represents the loaded governance policies, defining rules for deployments, static analysis, and resource limits.
*   **RuntimeEnvironment (internal/services):**  Manages the Salesforce CLI (sfdx) environment, including installation, plugin management, and health checks.
*   **IdentityResolver (internal/services):** Handles authentication with Salesforce, obtaining an organization ID.
*   **contract.DeploymentContract:** A data structure representing the deployment contract, containing metadata about the deployment, quality metrics (test coverage, test results), and audit information.
*   **analyzer.Violation:** Represents a single violation found during static code analysis.
*   **gha.Inputs:**  Represents the inputs provided to the GitHub Action.

**Important Functions and Their Behavior:**

*   **`main()`:** The primary entry point. Initializes the runtime context and calls the `run()` function. Handles top-level error handling by setting the GitHub Action status to failed and exiting with a non-zero code.
*   **`run(ctx context.Context) error`:**  The core function that executes the runtime logic. It performs the following stages:
    *   **Environment and Input Validation:** Checks for required environment variables (e.g., `GITHUB_WORKSPACE`, `GITHUB_REPOSITORY`) and inputs (e.g., `client_id`, `jwt_key`, `username`). Validates the format of the JWT key and Salesforce instance URL.
    *   **Resource Limits Validation:** Implements a timeout mechanism to prevent runaway executions. A goroutine is launched to terminate the process if it exceeds a predefined time limit.
    *   **Data Integrity & Compliance Checks:** Validates the context of the execution (e.g., pull request vs. direct commit) and performs additional security checks when running on forked repositories. Validates the repository format.
    *   **Telemetry Initialization:** Initializes OpenTelemetry for tracing and monitoring.
    *   **Policy Evaluation:** Loads and evaluates governance policies using the `policy.New()` and `policyEngine.Load()` functions.  Includes static code analysis using the `analyzer` package if enabled in the policy.
    *   **Bootstrap Phase:** Installs the Salesforce CLI using the `services.NewRuntimeEnvironment()` and `runtime.Install()` functions. Performs a health check to ensure the CLI is functioning correctly. Installs plugins if specified.
    *   **Identity Phase:** Authenticates with Salesforce using the `services.NewIdentityResolver()` and `identity.Authenticate()` functions.
    *   **Contract Validation Phase:** Generates a deployment contract containing metadata about the deployment, quality metrics, and audit information. Writes the contract to a JSON file.
    *   **Output Session State:** Sets GitHub Action outputs (e.g., `org_id`, `glassops_ready`) to provide information to subsequent steps in the pipeline.
*   **`isValidURL(s string) bool`:** Checks if a given string is a valid URL.
*   **`generateUUID() string`:** Generates a unique identifier (UUID) for the runtime execution.
*   **`splitAndTrim(s, sep string) []string`:** Splits a string by a separator and trims whitespace from each resulting part.
*   **`parseFloat(s string, defaultVal float64) float64`:** Parses a string to a float64, returning a default value if parsing fails.

**Error Handling:**

The runtime employs a consistent error handling pattern. Functions return an `error` value. Errors are checked immediately, and if an error occurs, the function returns the error, which is then propagated up the call stack.  The `fmt.Errorf` function is used to wrap errors with additional context.  GitHub Action status is set to failed when a critical error occurs.

**Concurrency:**

The runtime uses a single goroutine for the main execution flow. A separate goroutine is launched to implement the execution timeout mechanism. This allows the main process to continue while the timeout is monitored in the background.

**Notable Design Decisions:**

*   **Policy-Driven:** The runtime is designed to be highly configurable through governance policies. This allows organizations to tailor the runtime to their specific security and compliance requirements.
*   **GitHub Actions Integration:** The runtime is specifically designed to integrate with GitHub Actions, leveraging its inputs, outputs, and status reporting mechanisms.
*   **Modular Design:** The runtime is composed of several independent packages (e.g., `analyzer`, `policy`, `services`), promoting code reusability and maintainability.
*   **Contract Generation:** The generation of a deployment contract provides a clear and auditable record of the deployment process, including quality metrics and audit information.
*   **Telemetry:** Integration with OpenTelemetry provides valuable insights into runtime performance and behavior.