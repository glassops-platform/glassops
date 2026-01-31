---
type: Documentation
domain: control-plane
origin: packages/control-plane/internal/contract/contract.go
last_modified: 2026-01-31
generated: true
source: packages/control-plane/internal/contract/contract.go
generated_at: 2026-01-31T09:46:21.406006
hash: f6b1fae98e3691bd1375c6e96ecbe12185ba2ce76954ca639a7efb13859d2c67
---

## Deployment Contract Package Documentation

This package defines the `DeploymentContract` data structure, which represents the complete state of a deployment operation. It serves as a standardized format for communicating deployment results and associated metadata between different components of the system. We designed this contract to provide a clear and consistent view of deployment outcomes, enabling reliable monitoring, reporting, and auditing.

### Key Data Structures

**DeploymentContract:** This is the central structure. It encapsulates all information related to a deployment.

*   `SchemaVersion`: A string indicating the version of the contract schema used. This allows for future evolution of the contract without breaking compatibility.
*   `Meta`: Contains metadata about the deployment itself, such as the adapter used, the execution engine, and the triggering event.
*   `Status`: A string representing the overall status of the deployment. Possible values include "Succeeded", "Failed", or "Blocked".
*   `Quality`: Holds information about the quality checks performed during the deployment.
*   `Audit`: Contains audit information, detailing who or what initiated the deployment and the associated repository details.

**Meta:**  Provides contextual information about the deployment.

*   `Adapter`: The name of the adapter responsible for initiating the deployment.
*   `Engine`: The deployment engine used (e.g., "native", "hardis", "custom").
*   `Timestamp`: The time the deployment was triggered.
*   `Trigger`:  A string describing the event that triggered the deployment.

**Quality:**  Aggregates quality-related metrics.

*   `Coverage`: Details code coverage results.
*   `Tests`: Summarizes test execution results.
*   `StaticAnalysis`: (Optional) Contains results from static analysis tools like MegaLinter.  This field is included starting with Phase 1.5.

**Coverage:** Represents code coverage information.

*   `Actual`: The actual code coverage percentage achieved.
*   `Required`: The minimum required code coverage percentage.
*   `Met`: A boolean indicating whether the required coverage was met.

**Tests:** Summarizes test results.

*   `Total`: The total number of tests executed.
*   `Passed`: The number of tests that passed.
*   `Failed`: The number of tests that failed.

**StaticAnalysis:** Represents findings from static analysis tools.

*   `Tool`: The name of the static analysis tool used.
*   `Met`: A boolean indicating whether the static analysis criteria were met.
*   `CriticalViolations`: The number of critical violations found.
*   `HighViolations`: The number of high-severity violations found.
*   `BlockingViolations`: A list of strings describing violations that are blocking the deployment.

**Audit:**  Provides audit trail information.

*   `TriggeredBy`: The user or system that triggered the deployment.
*   `OrgID`: The organization ID associated with the deployment.
*   `Repository`: The repository where the deployment occurred.
*   `Commit`: The commit hash associated with the deployment.

### Error Handling

This package primarily focuses on data representation. Error handling is expected to be managed by the components that consume and produce instances of the `DeploymentContract`. We anticipate that consumers will check the `Status` field and examine the `Quality` and `Audit` fields for details in case of failures.

### Concurrency

This package does not directly involve concurrency. The structures defined here are intended to be passed between goroutines, but concurrency management is the responsibility of the calling code.

### Design Decisions

We chose a structured approach with nested data structures to represent the deployment contract. This allows for a clear and organized representation of complex deployment information. The inclusion of the `SchemaVersion` field is important for maintaining compatibility as the contract evolves. The optional `StaticAnalysis` field allows for flexibility and the addition of new quality checks without breaking existing integrations.