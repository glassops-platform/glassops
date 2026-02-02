---
type: Documentation
domain: runtime
origin: packages/runtime/internal/integration/identity_test.go
last_modified: 2026-02-01
generated: true
source: packages/runtime/internal/integration/identity_test.go
generated_at: 2026-02-01T19:41:48.904505
hash: 28902c7ad4f9dd34f617b2e42846dddebe74fdf681c48ea5294e97c9712c2337
---

## Identity Integration Test Documentation

This document describes the `integration` package, specifically the `identity_test.go` file, which focuses on testing the integration of identity-related functionality within the system. The primary responsibility of this package is to verify the correct parsing and handling of Salesforce authentication URLs.

**Key Types and Interfaces**

The tests interact with the `services.Identity` type. This type, defined in the `github.com/glassops-platform/glassops/packages/runtime/internal/services` package, is responsible for parsing authentication URLs and extracting relevant information, such as the organization ID.  The tests do not directly define any new types or interfaces; they primarily exercise the functionality of the existing `Identity` service.

**Important Functions and Behavior**

The core functionality being tested resides within the `TestIdentityIntegration` function, which contains several sub-tests:

*   **`TestIdentityIntegration/parses valid SFDX auth URL`**: This test case verifies the `ParseAuthURL` function's ability to process a valid, albeit simplified, Salesforce authentication URL (`force://PlatformCLI::refresh_token@login.salesforce.com`). It anticipates a potential error because the provided URL does not contain an organization ID, testing the parsing logic's behavior in such scenarios. The returned organization ID is ignored in this test.

*   **`TestIdentityIntegration/validates auth URL format`**: This test case employs a table-driven approach to validate the `ParseAuthURL` function against a variety of authentication URL formats. It includes tests for:
    *   Valid URLs with standard domains (`login.salesforce.com`).
    *   Sandbox URLs (`test.salesforce.com`).
    *   Custom domain URLs (`company.my.salesforce.com`).
    *   Invalid inputs (empty string, malformed URL).
    The test asserts that `ParseAuthURL` returns an error for invalid URLs and does not return an error for valid URLs.

*   **`TestIdentityIntegration/handles environment auth URL`**: This test verifies that the system correctly retrieves the authentication URL from the environment variable `INPUT_SFDX_AUTH_URL`. It sets this environment variable to a specific value, then reads it back using `os.Getenv` to confirm that the value is correctly propagated.

**Error Handling**

The tests extensively check for expected errors when invalid input is provided to the `ParseAuthURL` function.  The tests use `t.Error` and `t.Errorf` to report failures when errors are not returned when expected, or when unexpected errors occur.  The `t.Logf` function is used to log expected errors for specific test cases, such as the mock URL test.

**Concurrency**

This code does not employ any concurrency patterns such as goroutines or channels. The tests are sequential and operate within a single goroutine.

**Notable Design Decisions**

*   **Integration Tests**: The tests are designed as integration tests, meaning they verify the interaction between different components (in this case, the `Identity` service and environment variable handling).
*   **Test Workspace**: The `SetupTestWorkspace` and `env.Cleanup()` functions are used to create and tear down a controlled test environment. This ensures that tests do not interfere with each other or the system's state.
*   **Table-Driven Tests**: The `validates auth URL format` test case uses a table-driven approach, which makes the tests more readable, maintainable, and easier to extend with new test cases.
*   **Short Mode**: The `testing.Short()` check allows skipping the integration tests when running in short mode, which is useful for faster development cycles.