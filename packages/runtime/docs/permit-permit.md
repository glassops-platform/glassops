---
type: Documentation
domain: runtime
origin: packages/runtime/internal/permit/permit.go
last_modified: 2026-02-01
generated: true
source: packages/runtime/internal/permit/permit.go
generated_at: 2026-02-01T19:42:24.153145
hash: 8393641c9390699777e2e1d098f63abef81af3599ccd95e0f1c87d8411460877
---

## Permit Package Documentation

This package manages the creation and persistence of a permit file, `glassops-permit.json`, which serves as a secure context handoff mechanism. It encapsulates runtime information, organizational details, policy configurations, and relevant environment variables. This permit is designed to be consumed by external processes requiring authorized access to resources.

**Key Types**

*   `Permit`: This struct represents the permit itself. It contains the following fields:
    *   `RuntimeID`: A string identifying the runtime environment.
    *   `Timestamp`: A string representing the time of permit creation, formatted according to RFC3339.
    *   `OrgID`: A string identifying the organization associated with the permit.
    *   `Policy`: A pointer to a `policy.Config` struct, containing the applicable policy rules.
    *   `Context`: A map of strings providing contextual information about the environment, such as the GitHub actor, repository, commit SHA, and workspace path.
    *   `Inputs`: A map of strings containing input parameters, such as the instance URL, username, and client ID.

**Functions**

*   `Generate(runtimeID string, orgID string, config *policy.Config, instanceURL string) (string, error)`: This function creates a `Permit` instance, marshals it to JSON, and writes it to a file named `glassops-permit.json` within the designated workspace.
    *   It determines the workspace path using the `GITHUB_WORKSPACE` environment variable, defaulting to the current directory if the variable is not set.
    *   It populates the `Context` map with values from environment variables: `GITHUB_ACTOR`, `GITHUB_REPOSITORY`, `GITHUB_SHA`, and the resolved workspace path.
    *   It populates the `Inputs` map with the provided `instanceURL` and values retrieved using the `gha.GetInput` function for `username` and `client_id`.
    *   The function returns the path to the created permit file on success, and an error if any step fails (marshaling to JSON or writing to file).

**Error Handling**

The `Generate` function employs standard Go error handling practices. Errors encountered during JSON marshaling or file writing are wrapped using `fmt.Errorf` to provide context and preserve the original error for debugging.  These errors are then returned to the caller.

**Design Decisions**

*   **JSON Format:** The permit is serialized to JSON for portability and ease of parsing by various consumers.
*   **Environment Variable Reliance:** The package relies on environment variables (specifically those prefixed with `GITHUB_`) to gather contextual information. This makes it well-suited for use within GitHub Actions workflows.
*   **File Persistence:** Writing the permit to a file ensures that the context is available to subsequent steps in a workflow or to external processes.
*   **Workspace Determination:** The package attempts to determine the workspace dynamically using the `GITHUB_WORKSPACE` environment variable, providing flexibility in different environments.