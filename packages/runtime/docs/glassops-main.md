---
type: Documentation
domain: runtime
last_modified: 2026-02-02
generated: true
source: packages/runtime/cmd/glassops/main.go
generated_at: 2026-02-02T22:34:10.208343
hash: ac422471f6c7e21810341ed39d4c4ff3918cb8385a5cfd46e01e56a27a48cb0e
---

## GlassOps Runtime Documentation

This document describes the GlassOps Runtime, a system designed to provide governed execution of operations, particularly within a Salesforce environment. It details the runtime’s purpose, key components, and operational flow.

**Package Purpose**

The `main` package serves as the entry point for the GlassOps Runtime. It orchestrates a series of validation, policy enforcement, identity resolution, and context generation steps to prepare for and authorize governed operations. The runtime is intended to be used within a GitHub Actions workflow, as indicated by its interaction with the `gha` package.

**Key Types and Interfaces**

*   **Config (from `policy` package):** Represents the loaded governance policies. This configuration drives policy checks and permit generation.
*   **AuthRequest (from `services` package):** A structure containing the necessary credentials and connection details for authenticating with a Salesforce instance. It includes fields for `ClientID`, `JWTKey`, `Username`, and `InstanceURL`.
*   **IdentityResolver (from `services` package):** An interface responsible for authenticating with Salesforce and retrieving the organization ID.

**Important Functions and Their Behavior**

*   **`main()`:** The primary function that initiates the runtime process. It sets up the context, calls the `run()` function, and handles any errors that occur during execution. If an error occurs, it sets the GitHub Action status to “failed” and exits.
*   **`run(ctx context.Context) error`:** This function contains the core logic of the runtime. It performs the following steps:
    1.  **Environment and Input Validation:** Validates the environment and input parameters to ensure they meet predefined criteria.
    2.  **Timeout Enforcement:** Starts a goroutine (`enforceTimeout`) to monitor execution time and terminate the runtime if it exceeds a configured limit (30 minutes).
    3.  **Policy Evaluation:** Loads and evaluates governance policies using the `policy` package. It checks for active freeze windows and sets an output variable (`is_locked`) to indicate whether policy restrictions are in effect. Static code analysis is performed if enabled in the policy configuration.
    4.  **Identity Resolution:** Authenticates with Salesforce using the `services` package. It can skip authentication if configured, using a dummy organization ID for testing purposes.
    5.  **Permit Generation:** Generates a GlassOps Permit, which encapsulates the runtime context and authorization information, using the `permit` package.
    6.  **Contract Generation:** Generates a deployment contract using the `contract` package.
    7.  **Output Setting:** Sets various output variables for the GitHub Action workflow, including the runtime ID, organization ID, contract path, and a “glassops\_ready” flag.
*   **`generateUUID()`:** Generates a Universally Unique Identifier (UUID) used as a runtime identifier.
*   **`enforceTimeout(limit time.Duration)`:** A goroutine that sleeps for the specified duration and then terminates the runtime if the limit is exceeded. It sets the GitHub Action status to “error” before exiting.

**Error Handling**

The runtime employs a consistent error handling pattern. Functions return an `error` value, which is checked by the calling function. If an error occurs, it is wrapped with additional context using `fmt.Errorf("%w", err)` to preserve the original error while adding information about where the error occurred.  Errors are also communicated to the GitHub Actions workflow using `gha.SetFailed()`. Warnings are logged using `gha.Warning()`.

**Concurrency**

The runtime uses a goroutine to enforce the execution timeout. This allows the main execution flow to continue while the timeout is monitored in the background. The `enforceTimeout` function sleeps for the specified duration and then terminates the runtime if the limit is exceeded.

**Notable Design Decisions**

*   **Modular Design:** The runtime is structured into distinct packages (e.g., `validator`, `policy`, `services`, `permit`, `contract`) to promote code organization, maintainability, and testability.
*   **GitHub Actions Integration:** The runtime is designed to be used within a GitHub Actions workflow, leveraging the `gha` package for input/output handling and status reporting.
*   **Policy-Driven Governance:** The runtime enforces governance policies defined in a configuration file, providing a flexible and configurable approach to controlling operations.
*   **Permit-Based Authorization:** The GlassOps Permit serves as a central artifact for authorization, encapsulating the runtime context and permissions.
*   **Telemetry Integration:** The runtime integrates with OpenTelemetry for tracing and monitoring. Initialization can fail gracefully, logging a warning if telemetry is unavailable.