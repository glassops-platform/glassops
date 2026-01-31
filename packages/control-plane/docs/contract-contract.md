---
type: Documentation
domain: control-plane
origin: packages/control-plane/internal/contract/contract.go
last_modified: 2026-01-31
generated: true
source: packages/control-plane/internal/contract/contract.go
generated_at: 2026-01-31T08:51:24.921785
hash: f6b1fae98e3691bd1375c6e96ecbe12185ba2ce76954ca639a7efb13859d2c67
---

## Deployment Contract Specification

This document details the structure of the Deployment Contract used within the system. This contract serves as a standardized format for representing the outcome of a deployment attempt, encompassing metadata, status, quality metrics, and audit information. It is designed to be machine-readable and human-interpretable, facilitating consistent reporting and decision-making.

**Package Responsibilities:**

The `contract` package defines the data structures used to communicate the results of a deployment process. It provides a common language for different components of the system to understand and react to deployment outcomes.

**Key Data Structures:**

* **`DeploymentContract`:** This is the primary structure representing the complete contract. It encapsulates all information related to a single deployment.
    * `SchemaVersion`: A string indicating the version of the contract schema used. This allows for future evolution of the contract without breaking compatibility.
    * `Meta`: Contains metadata about the deployment.
    * `Status`: A string representing the overall status of the deployment ("Succeeded", "Failed", or "Blocked").
    * `Quality`: Holds quality-related metrics for the deployment.
    * `Audit`: Provides audit information, tracking the origin and context of the deployment.

* **`Meta`:**  Stores contextual information about the deployment.
    * `Adapter`: The name of the adapter used to initiate the deployment.
    * `Engine`: The deployment engine used (e.g., "native", "hardis", "custom").
    * `Timestamp`: The time the deployment was triggered.
    * `Trigger`:  Identifies the event that initiated the deployment.

* **`Quality`:**  Aggregates quality assurance data.
    * `Coverage`:  Details code coverage metrics.
        * `Actual`: The actual code coverage achieved.
        * `Required`: The minimum required code coverage.
        * `Met`: A boolean indicating whether the required coverage was met.
    * `Tests`:  Summarizes test execution results.
        * `Total`: The total number of tests executed.
        * `Passed`: The number of tests that passed.
        * `Failed`: The number of tests that failed.
    * `StaticAnalysis`: (Optional) Contains results from static analysis tools.  This field is included to support Phase 1.5 enhancements.

* **`StaticAnalysis`:** Represents the findings from static analysis tools like MegaLinter or similar scanners.
    * `Tool`: The name of the static analysis tool used.
    * `Met`: A boolean indicating whether the static analysis criteria were met.
    * `CriticalViolations`: The number of critical violations found.
    * `HighViolations`: The number of high-severity violations found.
    * `BlockingViolations`: A list of strings describing violations that are blocking the deployment.

* **`Audit`:**  Provides traceability information.
    * `TriggeredBy`: The user or system that initiated the deployment.
    * `OrgID`: The organization ID associated with the deployment.
    * `Repository`: The repository where the deployment occurred.
    * `Commit`: The commit hash associated with the deployment.

**Important Functions:**

This package primarily defines data structures and does not contain functions.  The structures are populated and used by other components of the system.

**Error Handling:**

This package does not directly handle errors. Error handling is the responsibility of the components that create and process `DeploymentContract` instances.  Validation of the data within the contract should be performed by consuming services.

**Concurrency:**

This package does not employ concurrency patterns directly. However, the structures it defines may be accessed concurrently by different goroutines in other parts of the system.  It is the responsibility of those components to ensure appropriate synchronization if necessary.

**Design Decisions:**

* **JSON Serialization:** The structures are designed to be easily serialized to JSON for communication and storage. The `json:"..."` tags facilitate this process.
* **Optional Fields:** The `StaticAnalysis` field within `Quality` is optional (using `json:",omitempty"`). This allows for flexibility and supports scenarios where static analysis is not performed.
* **Clear Status Values:** The `Status` field uses a limited set of predefined values ("Succeeded", "Failed", "Blocked") to ensure clarity and consistency.