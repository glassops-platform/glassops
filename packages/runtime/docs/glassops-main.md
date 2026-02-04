---
type: Documentation
domain: runtime
last_modified: 2026-02-03
generated: true
source: packages/runtime/cmd/glassops/main.go
generated_at: 2026-02-03T18:06:57.411570
hash: 3e5f202f397336cce0a0b3fd539acc45f5da69d512f51ba1bbe34c3703c1923a
---

## GlassOps Runtime Documentation

This document describes the GlassOps Runtime, a core component responsible for governing and securing deployments. It orchestrates policy evaluation, identity management, and context generation to ensure compliant and authorized operations.

**Package Purpose:**

The `main` package serves as the entry point for the GlassOps Runtime. It coordinates the execution of several internal packages to provide a secure and governed environment for deployments. The runtime is designed to be invoked as part of a larger automation pipeline, such as a GitHub Action.

**Key Types and Interfaces:**

*   **`permit.PolicyEvaluation`**: Represents the outcome of policy checks. It includes a boolean `Allowed` flag indicating whether the operation is permitted, a list of `Evaluated` policies, and any `Violations` encountered.
*   **`permit.Identity`**:  Encapsulates information about the actor attempting to perform an action. It includes the `Subject` (user identifier), `Provider` (authentication source), `ProviderID`, and a `Verified` status.
*   **`services.AuthRequest`**: A structure containing the necessary information for authenticating with Salesforce, including `ClientID`, `JWTKey`, `Username`, and `InstanceURL`.

**Important Functions:**

*   **`main()`**: The primary entry point of the runtime. It initializes the environment, calls the `run()` function, and handles any errors that occur during execution.
*   **`run(ctx context.Context) error`**:  This function contains the core logic of the runtime. It performs the following steps:
    1.  **Environment Validation**: Validates the runtime environment using the `validator` package.
    2.  **Input Validation**: Validates and sanitizes input parameters received from the environment. It can load a JWT key from a file specified by the `jwt_key_file` input.
    3.  **Resource Limits**: Enforces a maximum execution time using a timeout mechanism implemented with the `enforceTimeout` function.
    4.  **Policy Evaluation**: Loads and evaluates governance policies using the `policy` package.  It checks for active freeze windows if policy enforcement is enabled.
    5.  **Identity Resolution**: Authenticates with Salesforce using the `services` package, obtaining the organization ID. Authentication can be skipped for testing purposes.
    6.  **Permit Generation**: Generates a GlassOps Permit containing the runtime ID, actor identity, policy evaluation results, and instance URL using the `permit` package.
    7.  **Contract Generation**: Generates a deployment contract using the `contract` package.
    8.  **Output**: Sets output variables for use in subsequent workflow steps, including the runtime ID, organization ID, contract path, and a flag indicating readiness.
*   **`generateUUID()`**: Generates a Universally Unique Identifier (UUID) used as a runtime identifier.
*   **`enforceTimeout(limit time.Duration)`**:  Implements a timeout mechanism. It sleeps for the specified duration and then terminates the runtime if the limit is exceeded.

**Error Handling:**

The runtime employs a consistent error handling pattern. Functions return an `error` value, which is checked by the caller. If an error occurs, the function immediately returns, and the error is propagated up the call stack.  The `fmt.Errorf` function is used to wrap errors, providing additional context.  Critical errors result in the runtime exiting with a non-zero status code, and failures are reported to the GitHub Actions workflow using `gha.SetFailed`.

**Concurrency:**

The runtime uses a single goroutine for its main execution flow. However, the `enforceTimeout` function is launched as a separate goroutine to monitor execution time without blocking the main process.

**Notable Design Decisions:**

*   **Modular Design**: The runtime is composed of several internal packages, each responsible for a specific aspect of the governance process. This promotes code reusability and maintainability.
*   **Configuration-Driven**: The runtime's behavior is largely driven by configuration loaded from policy files. This allows for flexible and customizable governance rules.
*   **GitHub Actions Integration**: The runtime is designed to be easily integrated with GitHub Actions, using the `gha` package to interact with the workflow environment.
*   **Telemetry**: The runtime includes optional OpenTelemetry integration for monitoring and tracing.
*   **Skip Authentication**: The ability to skip authentication is provided for testing and development purposes, but should not be used in production environments.