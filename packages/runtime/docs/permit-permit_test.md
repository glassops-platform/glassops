---
type: Documentation
domain: runtime
last_modified: 2026-02-03
generated: true
source: packages/runtime/internal/permit/permit_test.go
generated_at: 2026-02-03T18:07:55.937582
hash: 494997cfde68a132050998b1f0d8b336c707ad583d1514100f4e39074c5794d4
---

## Permit Package Documentation

This package is responsible for generating permit files. These files represent authorization decisions made at runtime, containing information about the actor requesting access, the policies evaluated, and relevant contextual inputs. The permits are designed to be used by downstream systems to enforce access control.

**Key Types:**

*   `Identity`: Represents the authenticated user or service account. It includes:
    *   `Subject`: A human-readable identifier for the actor (e.g., username, service account name).
    *   `Provider`: The authentication provider (e.g., "github").
    *   `ProviderID`: A unique identifier assigned by the provider.
    *   `Verified`: A boolean indicating if the identity has been verified.
*   `PolicyEvaluation`: Represents the result of evaluating policies against an actor. It includes:
    *   `Allowed`: A boolean indicating whether access is granted.
    *   `Evaluated`: A slice of strings representing the names of the policies that were evaluated.
*   `Permit`: Represents the generated permit file. It contains:
    *   `PermitID`: A unique identifier for the permit, often corresponding to the runtime identifier.
    *   `Actor`: The `Identity` of the actor the permit applies to.
    *   `Policies`: The `PolicyEvaluation` result.
    *   `Inputs`: A map of string keys to string values, holding contextual information relevant to the permit (e.g., instance URL).

**Important Functions:**

*   `Generate(runtimeID string, actor Identity, evaluation PolicyEvaluation, instanceURL string) (string, error)`: This function generates a permit file.
    *   It takes the runtime identifier, actor information, policy evaluation result, and instance URL as input.
    *   It creates a `Permit` object populated with the provided data.
    *   It serializes the `Permit` object to JSON.
    *   It writes the JSON data to a file. The filename is derived from the `runtimeID`.
    *   It returns the path to the generated permit file and an error if any occurred during the process.

**Error Handling:**

The `Generate` function returns an error value. Common errors include:

*   Errors during JSON serialization.
*   Errors during file creation or writing.

The test suite verifies that errors are handled correctly and that the expected files are created with the correct content.

**Design Decisions:**

*   **JSON Format:** Permits are serialized to JSON for easy parsing and consumption by other systems.
*   **File-Based Output:** Permits are written to files, providing a persistent record of authorization decisions. The file path is based on the `runtimeID` to ensure uniqueness.
*   **Environment Variables:** The test suite uses environment variables (`GITHUB_WORKSPACE`, `GITHUB_ACTOR`, `GITHUB_REPOSITORY`, `GITHUB_SHA`) to simulate a typical runtime environment, such as a CI/CD pipeline. This allows for realistic testing of the permit generation process.
*   **Testability:** The package is designed to be easily testable, with a comprehensive test suite that verifies the core functionality.