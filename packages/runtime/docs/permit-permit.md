---
type: Documentation
domain: runtime
last_modified: 2026-02-03
generated: true
source: packages/runtime/internal/permit/permit.go
generated_at: 2026-02-03T18:07:38.315210
hash: 8b4453576fd65b807d23f9aeca69c49444e73fd7a13404e19a50f2462a67a2c9
---

## Permit Package Documentation

This package is responsible for creating and persisting a permit artifact to disk. This artifact represents a governance decision point within a workflow, capturing information about the actor requesting access, the policies evaluated, and the context in which the request was made. It serves as a handoff contract for downstream processes.

**Key Types:**

*   **Identity:** Represents the entity requesting access. It contains the following fields:
    *   `Subject`: A human-readable identifier for the actor.
    *   `Provider`: The authentication provider used (e.g., GitHub, Google).
    *   `ProviderID`: The unique identifier assigned by the provider.
    *   `Verified`: A boolean indicating whether the identity has been verified.

*   **PolicyEvaluation:**  Holds the results of policy checks.
    *   `Allowed`: A boolean indicating whether the request is permitted based on policy.
    *   `Evaluated`: A string slice listing the policies that were evaluated.
    *   `Violations`: A string slice containing details of any policies that were violated. This field is omitted if no violations occurred.

*   **Permit:** The central data structure representing the governance decision.
    *   `Version`: The version of the permit schema. Currently "1.0".
    *   `PermitID`: A unique identifier for this permit instance.
    *   `Timestamp`: The time the permit was generated, in RFC3339 format.
    *   `Actor`: The `Identity` of the actor making the request.
    *   `Policies`: The `PolicyEvaluation` result.
    *   `Context`: A map of strings providing contextual information about the environment.
    *   `Inputs`: A map of strings containing input parameters relevant to the permit.

**Important Functions:**

*   **Generate(permitID string, actor Identity, evaluation PolicyEvaluation, instanceURL string) (string, error):**
    This function creates a `Permit` instance, serializes it to JSON, and writes it to a file named `glassops-permit.json` within a `.glassops` directory in the workspace.

    *   `permitID`: A string providing a unique identifier for the permit.
    *   `actor`: The `Identity` of the actor requesting access.
    *   `evaluation`: The `PolicyEvaluation` result.
    *   `instanceURL`: A string representing the URL of the instance.

    The function first determines the workspace directory, defaulting to the current directory if the `GITHUB_WORKSPACE` environment variable is not set. It then populates the `Permit` struct with relevant data, including environment variables for repository, commit SHA, and workspace. Input parameters `username` and `client_id` are retrieved using the `gha.GetInput` function.  It creates the `.glassops` directory if it doesn't exist. Finally, it marshals the `Permit` to JSON with indentation for readability and writes the JSON to the specified file. The function returns the path to the created permit file and an error if any step fails.

**Error Handling:**

The `Generate` function employs standard Go error handling practices. Errors encountered during directory creation, JSON marshaling, or file writing are wrapped with context using `fmt.Errorf` to provide more informative error messages. These errors are then returned to the caller.

**Concurrency:**

This package does not currently employ any explicit concurrency mechanisms like goroutines or channels. The operations performed are primarily I/O bound and are executed sequentially within the `Generate` function.

**Design Decisions:**

*   **JSON Format:** The permit is serialized to JSON for portability and ease of consumption by other systems.
*   **File-Based Persistence:**  The permit is written to a file to provide a durable record of the governance decision. The location is within a hidden `.glassops` directory to avoid accidental modification.
*   **Environment Variable Integration:** The function leverages environment variables (e.g., `GITHUB_WORKSPACE`, `GITHUB_REPOSITORY`, `GITHUB_SHA`) to gather contextual information about the execution environment.
*   **Input Retrieval:** The package depends on the `gha` package to retrieve input parameters, indicating an intended integration with GitHub Actions.