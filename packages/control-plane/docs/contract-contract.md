---
type: Documentation
domain: control-plane
last_modified: 2026-02-02
generated: true
source: packages/control-plane/internal/contract/contract.go
generated_at: 2026-02-02T22:23:03.246611
hash: f6b1fae98e3691bd1375c6e96ecbe12185ba2ce76954ca639a7efb13859d2c67
---

## Deployment Contract Package Documentation

This package defines the `DeploymentContract` data structure, which represents the complete state of a deployment operation within the system. It serves as a standardized format for communicating deployment results and associated metadata between different components of the platform. We designed this contract to provide a clear, consistent view of deployment outcomes, enabling effective monitoring, auditing, and decision-making.

**Key Types**

*   **`DeploymentContract`**: This is the central type. It encapsulates all information related to a deployment, including its status, quality metrics, and audit trail.
    *   `SchemaVersion`: A string indicating the version of the contract schema used. This allows for future evolution of the contract without breaking compatibility.
    *   `Meta`: Contains metadata about the deployment itself.
    *   `Status`: A string representing the overall deployment status. Possible values include "Succeeded", "Failed", and "Blocked".
    *   `Quality`: Holds quality-related metrics for the deployment.
    *   `Audit`: Stores information about the deployment's origin and context.

*   **`Meta`**:  Provides contextual information about how the deployment was executed.
    *   `Adapter`: The name of the adapter used to initiate the deployment.
    *   `Engine`: The deployment engine used (e.g., "native", "hardis", "custom").
    *   `Timestamp`: The time the deployment was triggered.
    *   `Trigger`:  Indicates what caused the deployment (e.g., a pull request, a manual trigger).

*   **`Quality`**:  Aggregates quality assurance data.
    *   `Coverage`:  Details about code coverage.
        *   `Actual`: The actual code coverage achieved.
        *   `Required`: The minimum required code coverage.
        *   `Met`: A boolean indicating whether the required coverage was met.
    *   `Tests`:  Information about the executed tests.
        *   `Total`: The total number of tests executed.
        *   `Passed`: The number of tests that passed.
        *   `Failed`: The number of tests that failed.
    *   `StaticAnalysis`: (Optional, introduced in Phase 1.5) Results from static analysis tools.

*   **`StaticAnalysis`**: Represents the output of static analysis tools like MegaLinter or similar scanners.
    *   `Tool`: The name of the static analysis tool used.
    *   `Met`: A boolean indicating whether the static analysis criteria were met.
    *   `CriticalViolations`: The number of critical violations found.
    *   `HighViolations`: The number of high-severity violations found.
    *   `BlockingViolations`: A list of strings describing violations that are blocking the deployment.

*   **`Audit`**:  Provides traceability information.
    *   `TriggeredBy`: The user or system that initiated the deployment.
    *   `OrgID`: The organization ID associated with the deployment.
    *   `Repository`: The repository where the code was deployed from.
    *   `Commit`: The commit hash associated with the deployment.

**Important Functions**

This package currently focuses on data structures and does not include any functions. Future iterations may include functions for validating or manipulating `DeploymentContract` instances.

**Error Handling**

This package does not currently define any error types or error handling logic. Error handling is expected to be implemented in the components that consume and process the `DeploymentContract` data.

**Concurrency**

This package does not employ any concurrency patterns (goroutines or channels) as it primarily defines data structures.

**Design Decisions**

*   **Schema Versioning**: The inclusion of `SchemaVersion` in the `DeploymentContract` is a deliberate design choice to support future evolution of the contract. This allows us to introduce changes without breaking compatibility with existing systems.
*   **Optional Static Analysis**: The `StaticAnalysis` field is optional (`omitempty`) to accommodate scenarios where static analysis is not performed or not available.
*   **Clear Status Values**: The `Status` field uses a limited set of predefined values ("Succeeded", "Failed", "Blocked") to ensure clarity and consistency in reporting deployment outcomes.