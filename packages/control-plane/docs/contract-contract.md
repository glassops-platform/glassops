---
type: Documentation
domain: control-plane
origin: packages/control-plane/internal/contract/contract.go
last_modified: 2026-02-01
generated: true
source: packages/control-plane/internal/contract/contract.go
generated_at: 2026-02-01T19:26:18.446166
hash: f6b1fae98e3691bd1375c6e96ecbe12185ba2ce76954ca639a7efb13859d2c67
---

## Deployment Contract Package Documentation

This package defines the `DeploymentContract` data structure, which represents the complete state of a deployment operation. It serves as a standardized format for communicating deployment results and associated metadata between different components of the system. We designed this contract to provide a clear and consistent view of deployment outcomes, enabling effective monitoring, auditing, and decision-making.

### Key Data Structures

**DeploymentContract:** This is the central structure. It encapsulates all information related to a deployment.

*   `SchemaVersion`: A string indicating the version of the contract schema used. This allows for future evolution of the contract without breaking compatibility.
*   `Meta`: Contains metadata about the deployment itself, such as the adapter used, the execution engine, and the triggering event.
*   `Status`: A string representing the overall status of the deployment. Possible values include "Succeeded", "Failed", or "Blocked".
*   `Quality`: Holds information about the quality checks performed during the deployment.
*   `Audit`: Provides audit trail information, identifying who or what triggered the deployment and the associated repository details.

**Meta:**  Metadata about the deployment execution.

*   `Adapter`: The name of the adapter responsible for initiating the deployment.
*   `Engine`: The deployment engine used (e.g., "native", "hardis", "custom").
*   `Timestamp`: The time the deployment was triggered.
*   `Trigger`:  A description of the event that initiated the deployment.

**Quality:**  Aggregates quality-related data.

*   `Coverage`: Details about code coverage metrics.
*   `Tests`: Results of automated tests.
*   `StaticAnalysis`: Findings from static analysis tools (optional, introduced in Phase 1.5).

**Coverage:** Represents code coverage information.

*   `Actual`: The actual code coverage percentage achieved.
*   `Required`: The minimum required code coverage percentage.
*   `Met`: A boolean indicating whether the required coverage was met.

**Tests:** Summarizes test execution results.

*   `Total`: The total number of tests executed.
*   `Passed`: The number of tests that passed.
*   `Failed`: The number of tests that failed.

**StaticAnalysis:**  Details findings from static analysis tools.

*   `Tool`: The name of the static analysis tool used.
*   `Met`: A boolean indicating whether the static analysis criteria were met.
*   `CriticalViolations`: The number of critical violations found.
*   `HighViolations`: The number of high-severity violations found.
*   `BlockingViolations`: A list of strings describing violations that are blocking the deployment.

**Audit:**  Provides information for auditing purposes.

*   `TriggeredBy`: The user or system that triggered the deployment.
*   `OrgID`: The organization ID associated with the deployment.
*   `Repository`: The repository where the deployment occurred.
*   `Commit`: The commit hash associated with the deployment.

### Error Handling

This package primarily focuses on data representation. Error handling is expected to be managed by the components that consume and process the `DeploymentContract` data. We anticipate that consumers will validate the data within the contract and handle any inconsistencies or invalid states appropriately.

### Concurrency

This package does not directly involve concurrency. The data structures are designed to be safely accessed and modified by concurrent processes, but concurrency management is the responsibility of the calling code.

### Design Decisions

*   **Schema Versioning:** The inclusion of `SchemaVersion` in the `DeploymentContract` is a deliberate design choice to support future evolution of the contract. This allows us to introduce new fields or modify existing ones without breaking compatibility with older systems.
*   **Optional Static Analysis:** The `StaticAnalysis` field is optional (`omitempty`) to accommodate scenarios where static analysis is not performed or not available.
*   **Clear Status Indicators:** The `Status` field provides a simple and unambiguous indication of the deployment outcome.
*   **Detailed Audit Information:** The `Audit` structure provides comprehensive information for tracking and auditing deployments.

You can create instances of these structures to represent the state of your deployments. For example:

```go
deployment := DeploymentContract{
    SchemaVersion: "1.0",
    Status:        "Succeeded",
    Meta: Meta{
        Adapter:   "github-adapter",
        Engine:    "native",
        Timestamp: time.Now(),
        Trigger:   "Pull Request Merge",
    },
    Quality: Quality{
        Coverage: Coverage{
            Actual:   85.0,
            Required: 80.0,
            Met:      true,
        },
        Tests: Tests{
            Total:  100,
            Passed: 95,
            Failed: 5,
        },
    },
    Audit: Audit{
        TriggeredBy: "user123",
        OrgID:       "my-org",
        Repository:  "my-repo",
        Commit:      "abcdef123456",
    },
}