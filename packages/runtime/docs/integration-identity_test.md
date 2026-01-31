---
type: Documentation
domain: runtime
origin: packages/runtime/internal/integration/identity_test.go
last_modified: 2026-01-29
generated: true
source: packages/runtime/internal/integration/identity_test.go
generated_at: 2026-01-29T21:24:57.776692
hash: 4b743bfd8fcfcdda70d960df2db21349f1b6a8a2773dbe2870f026b5a909204b
---

## Identity Integration Test Documentation

This document details the `integration` package, specifically the `identity_test.go` file, which contains integration tests for the identity service. The purpose of these tests is to verify the correct functionality of the identity service, particularly its ability to parse Salesforce authentication URLs.

**Package Responsibilities:**

The `integration` package focuses on testing the interaction between different components of the system, ensuring they work together as expected. This file specifically tests the `services.Identity` service.

**Key Types and Interfaces:**

*   `services.Identity`: This type (defined in the `internal/services` package) represents the identity service. It provides methods for parsing and validating Salesforce authentication URLs.  We interact with this service through its methods in these tests.
*   `TestWorkspace`: This is a custom type (defined within the test file, not exported) used to manage a test environment. It handles setup and cleanup tasks, including setting environment variables.

**Important Functions and Their Behavior:**

*   `TestIdentityIntegration(t *testing.T)`: This is the main test function. It orchestrates several sub-tests to cover different scenarios related to authentication URL parsing and validation. It skips execution if the test is run in short mode.
*   `SetupTestWorkspace(nil)`: This function sets up a temporary test environment. It returns a `TestWorkspace` object that can be used to manage the environment.  It handles any necessary initialization.
*   `env.Cleanup()`: This function, called via `defer`, cleans up the test environment after each test case, ensuring a clean state for subsequent tests.
*   `env.SetEnvironment(map[string]string)`: This function sets environment variables within the test environment. This is used to simulate scenarios where the authentication URL is provided via an environment variable.
*   `identity.ParseAuthURL(authURL string)`: This function, part of the `services.Identity` service, attempts to parse a Salesforce authentication URL and extract the organization ID. It returns the organization ID and an error if parsing fails.
*   `os.Getenv("INPUT_SFDX_AUTH_URL")`: This standard Go function retrieves the value of the `INPUT_SFDX_AUTH_URL` environment variable.

**Error Handling Patterns:**

The tests extensively check for expected errors returned by the `identity.ParseAuthURL` function.  We use `t.Fatalf` to immediately fail the test if the initial setup fails.  Within the URL parsing tests, we use `t.Error` and `t.Errorf` to report failures when the expected error condition is not met, or an unexpected error occurs.  The logging of expected errors provides insight into parsing behavior with incomplete URLs.

**Concurrency Patterns:**

This code does not exhibit any explicit concurrency patterns (goroutines, channels). The tests are executed sequentially within the `TestIdentityIntegration` function.

**Notable Design Decisions:**

*   **Integration Tests:** The use of integration tests ensures that the identity service interacts correctly with other parts of the system, such as the environment variable handling.
*   **Test Cases:** The `validates auth URL format` test uses a table-driven approach with multiple test cases to cover a range of valid and invalid authentication URL formats. This improves test coverage and readability.
*   **Environment Variable Simulation:** The `handles environment auth URL` test simulates a real-world scenario where the authentication URL is provided via an environment variable.
*   **Deferred Cleanup:** Using `defer env.Cleanup()` ensures that the test environment is always cleaned up, even if an error occurs during the test.
*   **Short Mode Skipping:** Skipping the integration test in short mode reduces test execution time during development.

**Function Signatures (with escaped generics):**

*   `identity.ParseAuthURL\[T any](authURL string) (T, error)`: Parses an authentication URL and returns the organization ID and an error. The type `\[T any]` indicates that the function can return any type, but in this case, it returns a string.