---
type: Documentation
domain: runtime
origin: packages/runtime/cmd/glassops/main.go
last_modified: 2026-01-31
generated: true
source: packages/runtime/cmd/glassops/main.go
generated_at: 2026-01-31T09:02:12.286119
hash: c16e6c382845f65a8d2af02a90b0d828dc4354bf2a489904917c35b59490031c
---

## GlassOps Runtime Documentation

This document describes the GlassOps Runtime, a tool designed to provide governance and automation for Salesforce deployments within a CI/CD pipeline. It acts as a central control point, enforcing policies and ensuring compliance throughout the deployment process.

**Package Purpose:**

The `main` package serves as the entry point for the GlassOps Runtime. It orchestrates the various phases of a deployment workflow, including environment validation, policy enforcement, Salesforce CLI bootstrapping, identity resolution, and contract generation.  The runtime is intended to be executed within a GitHub Actions environment, leveraging its inputs and outputs.

**Key Types and Interfaces:**

*   **config (internal/policy):** Represents the loaded governance policies, defining rules for deployments.
*   **RuntimeEnvironment (internal/services):** Manages the Salesforce CLI installation and provides access to its functionality.
*   **IdentityResolver (internal/services):** Handles authentication with Salesforce, obtaining an organization ID.
*   **contract.DeploymentContract:**  A data structure representing the deployment contract, containing metadata, quality metrics (test coverage, test results), and audit information.
*   **analyzer.Violation:** Represents a single violation found during static code analysis.
*   **gha.Inputs:**  Represents the inputs provided to the GitHub Action.

**Important Functions and Their Behavior:**

*   **`main()`:** The primary function that initializes the runtime context and calls the `run()` function to execute the deployment workflow. Handles top-level error handling and sets the GitHub Action status.
*   **`run(ctx context.Context)`:**  The core function that orchestrates the entire deployment process. It performs the following phases:
    *   **Environment Validation:** Checks for the presence of required environment variables (e.g., `GITHUB_WORKSPACE`, `GITHUB_REPOSITORY`).
    *   **Input Validation:** Validates required inputs provided to the GitHub Action (e.g., `client_id`, `jwt_key`).  Includes format validation for the JWT key and instance URL.
    *   **Resource Limits Validation:** Implements a timeout mechanism to prevent runaway executions.
    *   **Data Integrity Checks:** Validates the context of the execution (e.g., pull request context) and the repository format.
    *   **Initialization:** Generates a unique runtime ID, initializes OpenTelemetry for tracing, and prepares for cache operations (currently a placeholder).
    *   **Policy Phase:** Loads and evaluates governance policies.  Includes static code analysis using an analyzer, and checks for active freeze windows.
    *   **Bootstrap Phase:** Installs the Salesforce CLI and any specified plugins.
    *   **Identity Phase:** Authenticates with Salesforce using provided credentials, obtaining the organization ID.  Can be skipped if configured.
    *   **Contract Validation Phase:** Generates a deployment contract containing metadata about the deployment, quality metrics, and audit information.  Writes the contract to a JSON file.
    *   **Output Session State:** Sets outputs for the GitHub Action, indicating the status of the runtime and providing key information (e.g., organization ID, contract path).
*   **`isValidURL(s string)`:**  A helper function to validate if a given string is a valid URL.
*   **`generateUUID()`:** Generates a universally unique identifier (UUID) for tracking runtime instances.
*   **`splitAndTrim(s, sep string)`:** Splits a string by a separator and trims whitespace from each resulting part.
*   **`parseFloat(s string, defaultVal float64)`:** Parses a string to a float64, returning a default value if parsing fails or the input is empty.

**Error Handling:**

The runtime employs a consistent error handling pattern. Functions return an `error` value, which is checked by the caller. Errors are wrapped using `fmt.Errorf("%w", err)` to preserve the original error context.  Critical errors result in the GitHub Action being marked as failed using `gha.SetFailed()`. Warnings are logged using `gha.Warning()`.

**Concurrency:**

The runtime uses a single goroutine for the main execution flow. A separate goroutine is launched to implement the execution timeout, ensuring that the runtime does not run indefinitely.

**Notable Design Decisions:**

*   **GitHub Actions Integration:** The runtime is specifically designed to operate within a GitHub Actions environment, leveraging its input/output mechanisms and environment variables.
*   **Policy-Driven Governance:** The runtime enforces governance policies defined in a configuration file, providing a centralized mechanism for controlling deployments.
*   **Deployment Contract:** The generation of a deployment contract provides a standardized way to capture metadata about the deployment, enabling auditing and compliance.
*   **OpenTelemetry Integration:** The inclusion of OpenTelemetry allows for tracing and monitoring of the runtime's execution.
*   **Modular Design:** The runtime is structured into internal packages (e.g., `analyzer`, `policy`, `services`) to promote code organization and maintainability.