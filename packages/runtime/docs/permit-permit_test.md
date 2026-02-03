---
type: Documentation
domain: runtime
last_modified: 2026-02-02
generated: true
source: packages/runtime/internal/permit/permit_test.go
generated_at: 2026-02-02T22:38:31.998460
hash: 2e501bf163588d77f411b1ab391472d9d99a3eccb2aa8c342f9af5e83467be11
---

## Permit Package Documentation

This package is responsible for generating permit files. These files contain runtime information, organizational details, and contextual data needed for secure operations. The primary function of this package is to create a JSON-formatted permit based on provided inputs and environment variables.

**Key Types**

*   `Permit`: This is the central data structure. It’s a JSON-serializable type that holds the runtime ID, organization ID, a context map for environment-specific information, and an inputs map for configuration details.  The structure is defined elsewhere in the codebase and is not explicitly shown in this file.

**Functions**

*   `Generate(runtimeID string, orgID string, config *policy.Config, instanceURL string) (string, error)`: This function creates a permit file.
    *   `runtimeID`: A string identifying the runtime environment.
    *   `orgID`: A string identifying the organization.
    *   `config`: A pointer to a `policy.Config` struct (defined in the `policy` package). This parameter is currently unused in the provided code.
    *   `instanceURL`: A string representing the instance URL.
    *   Return values:
        *   A string representing the file path of the generated permit.
        *   An error if the permit generation fails.

    The function leverages environment variables (`GITHUB_WORKSPACE`, `GITHUB_ACTOR`, `GITHUB_REPOSITORY`, `GITHUB_SHA`) to populate the permit's context. It constructs a `Permit` object, marshals it to JSON, and writes the JSON data to a file within the `GITHUB_WORKSPACE` directory.

**Error Handling**

The `Generate` function returns an error value. The test case checks for errors during file creation, file reading, and JSON unmarshaling.  Specific errors handled include file not found errors when verifying the permit file's existence and errors during JSON parsing.

**Testing**

The `TestGenerate` function verifies the functionality of the `Generate` function. It performs the following checks:

1.  Sets up a temporary directory and environment variables to simulate a CI/CD environment.
2.  Calls the `Generate` function with test data.
3.  Confirms that the permit file is created at the expected path.
4.  Reads the permit file content.
5.  Unmarshals the JSON content into a `Permit` object.
6.  Validates that the `RuntimeID`, `OrgID`, `Context["actor"]`, and `Inputs["instance_url"]` fields of the `Permit` object match the expected values.

**Design Decisions**

*   **Environment Variable Reliance:** The package depends on environment variables for contextual information. This design choice allows for flexibility and integration with CI/CD pipelines.
*   **JSON Format:** The permit is serialized as JSON, which is a widely supported and human-readable format.
*   **Temporary Directory:** The test case uses a temporary directory to avoid modifying the file system during testing.
*   **Unused Config Parameter:** The `policy.Config` parameter is passed to the `Generate` function but is not currently used. This suggests potential future expansion of the function's capabilities.