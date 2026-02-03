---
type: Documentation
domain: runtime
last_modified: 2026-02-02
generated: true
source: packages/runtime/internal/permit/permit.go
generated_at: 2026-02-02T22:38:18.536544
hash: 8393641c9390699777e2e1d098f63abef81af3599ccd95e0f1c87d8411460877
---

## Permit Package Documentation

This package manages the creation and persistence of a permit file, `glassops-permit.json`, which serves as a contract for context handoff during runtime operations. It facilitates secure and controlled execution by providing necessary information about the environment, policy, and inputs.

**Key Types:**

*   **Permit:** This struct encapsulates all the information required for a runtime operation.
    *   `RuntimeID`: A string identifying the specific runtime instance.
    *   `Timestamp`: A string representing the time the permit was generated, formatted according to RFC3339.
    *   `OrgID`: A string identifying the organization associated with the operation.
    *   `Policy`: A pointer to a `policy.Config` struct, containing the defined policy for the runtime.
    *   `Context`: A map of strings providing contextual information about the environment, such as the GitHub actor, repository, commit SHA, and workspace.
    *   `Inputs`: A map of strings containing input parameters required for the runtime operation, including the instance URL, username, and client ID.

**Functions:**

*   **Generate(runtimeID string, orgID string, config \*policy.Config, instanceURL string) (string, error):**
    This function creates a `Permit` instance, marshals it to JSON, and writes it to a file named `glassops-permit.json` within the designated workspace.
    1.  It determines the workspace directory, defaulting to the current directory (`.`) if the `GITHUB_WORKSPACE` environment variable is not set.
    2.  It populates the `Permit` struct with runtime ID, timestamp, organization ID, policy configuration, contextual information (obtained from environment variables like `GITHUB_ACTOR`, `GITHUB_REPOSITORY`, `GITHUB_SHA`, and the workspace), and input parameters (including the provided `instanceURL` and values retrieved using `gha.GetInput`).
    3.  It marshals the `Permit` struct into a JSON format with indentation for readability.
    4.  It writes the JSON data to the `glassops-permit.json` file in the workspace with permissions set to 0644.
    5.  It returns the path to the created permit file and a potential error.

**Error Handling:**

The `Generate` function handles potential errors during JSON marshaling and file writing. If either of these operations fails, it returns an error message that includes the original error using `fmt.Errorf` with `%w` for error wrapping, providing context for debugging.

**Dependencies:**

*   `encoding/json`: For marshaling the `Permit` struct to JSON.
*   `fmt`: For formatted string output and error creation.
*   `os`: For interacting with the operating system, including reading environment variables and writing files.
*   `path/filepath`: For constructing file paths.
*   `time`: For generating timestamps.
*   `github.com/glassops-platform/glassops/packages/runtime/internal/gha`: For retrieving input values from the GitHub Actions environment.
*   `github.com/glassops-platform/glassops/packages/runtime/internal/policy`: For accessing the policy configuration.

**Design Decisions:**

*   The permit is written to a file (`glassops-permit.json`) to provide a persistent and easily accessible context for the runtime operation.
*   Environment variables are used to gather contextual information, making the permit generation process adaptable to different environments.
*   Error wrapping is used to provide more informative error messages, aiding in debugging and troubleshooting.
*   The use of a dedicated package promotes modularity and separation of concerns.