---
type: Documentation
domain: runtime
last_modified: 2026-02-02
generated: true
source: packages/runtime/internal/integration/identity_test.go
generated_at: 2026-02-02T22:37:48.628352
hash: 28902c7ad4f9dd34f617b2e42846dddebe74fdf681c48ea5294e97c9712c2337
---

## Identity Integration Test Documentation

This document details the purpose and functionality of the `integration` package, specifically focusing on the `identity_test.go` file. This package contains integration tests for the identity service, verifying its ability to parse and validate Salesforce authentication URLs.

**Package Responsibilities:**

The `integration` package focuses on testing the interaction between different components of the system, ensuring they work correctly together. In this case, it tests the `services.Identity` service’s ability to handle Salesforce authentication URLs. These tests are designed to confirm the service behaves as expected in a realistic environment.

**Key Types and Interfaces:**

*   `services.Identity`: This type, defined in the `internal/services` package, represents the identity service. It provides methods for parsing and validating Salesforce authentication URLs. We interact with this service through its methods during testing.
*   `TestWorkspace`: This is a custom type (defined elsewhere, not shown in this file) used to manage a test environment. It allows us to set up and clean up resources needed for the tests, including environment variables.

**Important Functions:**

*   `TestIdentityIntegration(t *testing.T)`: This is the main test function. It orchestrates several sub-tests to verify different aspects of the identity service. It skips execution if the test is run in short mode.
*   `SetupTestWorkspace(nil)`: This function (defined elsewhere) sets up a temporary test environment. It returns a `TestWorkspace` object that can be used to manage the environment.
*   `env.Cleanup()`: This function (method of `TestWorkspace`) cleans up the test environment after the tests have completed, releasing any resources that were allocated.
*   `env.SetEnvironment(map[string]string)`: This function (method of `TestWorkspace`) sets environment variables for the test process.
*   `identity.ParseAuthURL(authURL string)`: This method of the `services.Identity` type attempts to parse a Salesforce authentication URL and extract the organization ID. It returns the organization ID and an error if parsing fails.
*   `os.Getenv("INPUT_SFDX_AUTH_URL")`: This standard Go function retrieves the value of the environment variable named "INPUT_SFDX_AUTH_URL".

**Test Cases and Behavior:**

The `TestIdentityIntegration` function includes the following test cases:

1.  **Parses valid SFDX auth URL:** This test verifies that the `ParseAuthURL` method handles a valid, albeit simplified, authentication URL. It expects an error if the URL doesn't contain an organization ID (as it's a mock URL).
2.  **Validates auth URL format:** This test uses a table of test cases to verify that the `ParseAuthURL` method correctly validates the format of authentication URLs. It tests valid URLs (including sandbox and custom domain URLs) and invalid URLs (empty string and incorrect format). It asserts that the method returns an error for invalid URLs and no error for valid URLs.
3.  **Handles environment auth URL:** This test verifies that the service can read the authentication URL from the `INPUT_SFDX_AUTH_URL` environment variable. It sets the environment variable, then reads it back to confirm that the value is correct.

**Error Handling:**

The tests extensively check for expected errors. The `ParseAuthURL` method returns an error when it encounters an invalid URL or fails to parse the URL correctly. The tests assert that errors are returned in these cases and that no errors are returned when the URL is valid.  The tests use `t.Fatalf` to immediately stop execution if a critical setup error occurs, and `t.Error` or `t.Errorf` to report failures within individual test cases.

**Concurrency:**

This code does not exhibit any explicit concurrency patterns (goroutines, channels). The tests are executed sequentially within the `testing` framework.

**Design Decisions:**

*   **Integration Tests:** The use of integration tests ensures that the identity service works correctly with other components of the system, such as the test workspace environment setup.
*   **Table-Driven Tests:** The "validates auth URL format" test case uses a table-driven approach, which makes it easy to add new test cases and maintain the tests.
*   **Environment Variable Handling:** The test for environment variable handling verifies that the service can correctly retrieve the authentication URL from the environment, which is a common way to configure the service in a production environment.