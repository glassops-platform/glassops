---
type: Documentation
domain: runtime
origin: packages/runtime/internal/permit/permit_test.go
last_modified: 2026-02-01
generated: true
source: packages/runtime/internal/permit/permit_test.go
generated_at: 2026-02-01T19:42:45.595293
hash: 2e501bf163588d77f411b1ab391472d9d99a3eccb2aa8c342f9af5e83467be11
---

## Permit Package Documentation

This package is responsible for generating permit files. These files contain runtime information, organizational details, and contextual data needed for secure operations. The primary function of this package is to create a JSON-formatted permit based on provided inputs and environment variables.

**Key Types**

*   `Permit`: This is the core data structure representing the permit. It is a JSON-serializable type containing:
    *   `RuntimeID`: A string identifying the runtime environment.
    *   `OrgID`: A string identifying the organization.
    *   `Context`: A map of strings representing contextual information (e.g., the user triggering the action).
    *   `Inputs`: A map of strings representing input parameters (e.g., instance URLs).

**Functions**

*   `Generate(runtimeID string, orgID string, config *policy.Config, instanceURL string) (string, error)`:
    This function creates a permit file. It takes the runtime ID, organization ID, a policy configuration (currently unused in the test), and an instance URL as input. It constructs a `Permit` object, marshals it to JSON, and writes the JSON to a file within a temporary directory (defined by the `GITHUB_WORKSPACE` environment variable). The function returns the path to the generated permit file and an error if any occurred during the process.

**Error Handling**

The `Generate` function returns an error value. Errors can occur during file creation, JSON marshaling, or if there are issues accessing environment variables. The test suite verifies that errors are handled correctly.

**Environment Variables**

The `Generate` function relies on the following environment variables to populate the permit's context:

*   `GITHUB_WORKSPACE`: Specifies the directory where the permit file will be created.
*   `GITHUB_ACTOR`:  Provides the actor (user or system) initiating the action. This value is added to the permit's `Context`.
*   `GITHUB_REPOSITORY`: Provides the repository information.
*   `GITHUB_SHA`: Provides the commit SHA.

**Design Decisions**

*   **JSON Format:** The permit is serialized as JSON for easy parsing and portability.
*   **Environment Variable Integration:** The package leverages environment variables to inject contextual information, making it suitable for use in CI/CD pipelines and other automated environments.
*   **Temporary File Creation:** The permit file is created in a temporary directory to avoid conflicts and ensure clean operation. The test suite cleans up the environment variables after execution.
*   **Policy Configuration:** The `policy.Config` parameter is currently accepted by the `Generate` function but is not actively used in the provided code. This suggests potential future expansion to incorporate policy-based permit generation.

**Testing**

The included test case (`TestGenerate`) verifies the following:

*   The `Generate` function creates a file at the expected path.
*   The generated file contains valid JSON.
*   The `Permit` object unmarshaled from the file has the correct `RuntimeID`, `OrgID`, `Context`, and `Inputs` values, based on the input parameters and environment variables.