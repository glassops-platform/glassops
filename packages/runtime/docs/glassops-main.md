---
type: Documentation
domain: runtime
origin: packages/runtime/cmd/glassops/main.go
last_modified: 2026-02-01
generated: true
source: packages/runtime/cmd/glassops/main.go
generated_at: 2026-02-01T19:37:59.897465
hash: ac422471f6c7e21810341ed39d4c4ff3918cb8385a5cfd46e01e56a27a48cb0e
---

## GlassOps Runtime Documentation

This document describes the GlassOps Runtime, a system designed to provide governed execution of operations, particularly within a Salesforce environment. It details the runtime’s purpose, key components, and operational flow.

**Package Purpose and Responsibilities**

The `main` package serves as the entry point for the GlassOps Runtime. Its primary responsibility is to orchestrate a series of validation, authentication, and configuration steps to establish a secure and governed environment for subsequent operations. The runtime ensures adherence to defined policies and provides contextual information for deployments.

**Key Types and Interfaces**

While this runtime doesn’t define explicit interfaces, it interacts with several internal packages that do. These packages provide the core functionality:

*   **validator:**  Handles validation of the environment, inputs, and context to ensure data integrity and compliance.
*   **policy:**  Manages and evaluates governance policies, including freeze windows, to control execution.
*   **gha:** Provides an abstraction layer for interacting with the GitHub Actions environment, handling inputs, outputs, and logging.
*   **services:** Contains services for authenticating with external systems, such as Salesforce.
*   **telemetry:**  Handles the initialization and shutdown of OpenTelemetry for tracing and monitoring.
*   **analyzer:** Performs static analysis of code to identify potential issues.
*   **permit:** Generates a permit file containing contextual information.
*   **contract:** Generates a deployment contract file.

**Important Functions and Their Behavior**

*   **`main()`:** This function is the program’s entry point. It initializes the runtime, calls the `run()` function to perform the core logic, and handles any errors that occur during execution. If an error occurs, it sets a failure status in the GitHub Actions environment and exits.
*   **`run(ctx context.Context) error`:** This function encapsulates the main workflow of the runtime. It performs the following steps:
    1.  **Environment and Input Validation:** Validates the environment and user-provided inputs.
    2.  **Timeout Enforcement:** Starts a goroutine (`enforceTimeout`) to terminate the runtime if it exceeds a predefined execution time (30 minutes).
    3.  **Policy Evaluation:** Loads and evaluates governance policies.  It checks for active freeze windows and sets an output indicating whether the runtime is locked. Static analysis is performed if enabled in the policy configuration.
    4.  **Salesforce Authentication:** Authenticates with Salesforce using provided credentials or skips authentication if configured.
    5.  **Permit Generation:** Generates a GlassOps Permit file containing runtime ID, organization ID, and policy information.
    6.  **Contract Generation:** Generates a deployment contract file.
    7.  **Output Session State:** Sets outputs in the GitHub Actions environment, including the organization ID, runtime ID, and a flag indicating that the runtime is ready.
*   **`generateUUID()`:** Generates a Universally Unique Identifier (UUID) used as a runtime identifier.
*   **`enforceTimeout(limit time.Duration)`:** This function, executed in a separate goroutine, sleeps for the specified duration and then terminates the runtime if the limit is exceeded.

**Error Handling Patterns**

The runtime employs a consistent error handling pattern. Functions return an `error` value. The `run()` function checks for errors after each step and returns immediately if an error is encountered. Errors are wrapped using `fmt.Errorf` with `%w` to preserve the original error context.  The `gha.SetFailed()` function is used to report errors to the GitHub Actions environment.

**Concurrency Patterns**

The runtime uses a single goroutine for timeout enforcement (`enforceTimeout`). This goroutine operates independently and terminates the runtime if the execution time exceeds the defined limit.

**Notable Design Decisions**

*   **GitHub Actions Integration:** The runtime is designed to be executed within a GitHub Actions workflow, leveraging the `gha` package for input, output, and logging.
*   **Policy-Driven Governance:** The runtime incorporates a policy engine to enforce governance rules and control execution.
*   **Contextual Information:** The runtime generates a GlassOps Permit and a deployment contract to provide contextual information for subsequent operations.
*   **Telemetry Integration:** The runtime integrates with OpenTelemetry for tracing and monitoring.
*   **Defensive Programming:** The runtime includes multiple validation steps to ensure data integrity and prevent unexpected behavior.
*   **Timeout Mechanism:** The use of a separate goroutine for timeout enforcement ensures that the runtime will terminate even if it becomes unresponsive.

**Usage Instructions**

You need to provide the following inputs to the runtime via GitHub Actions:

*   `client_id`: The Salesforce client ID.
*   `jwt_key`: The JWT key for Salesforce authentication.
*   `username`: The Salesforce username.
*   `instance_url`: The Salesforce instance URL (defaults to `https://login.salesforce.com`).
*   `enforce_policy`: Set to "true" to enforce policies, or "false" to disable policy enforcement.
*   `skip_auth`: Set to "true" to skip Salesforce authentication (for testing purposes only).
*   `otel_service_name`: The name of the OpenTelemetry service (defaults to `glassops-runtime`).

The runtime will output the following values:

*   `runtime_id`: A unique identifier for the runtime.
*   `is_locked`: Indicates whether the runtime is locked due to active freeze windows.
*   `org_id`: The Salesforce organization ID.
*   `contract_path`: The path to the generated deployment contract file.
*   `glassops_ready`: Set to "true" when the runtime is ready for governed execution.