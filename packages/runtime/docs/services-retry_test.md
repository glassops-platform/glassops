---
type: Documentation
domain: runtime
origin: packages/runtime/internal/services/retry_test.go
last_modified: 2026-01-31
generated: true
source: packages/runtime/internal/services/retry_test.go
generated_at: 2026-01-31T09:10:13.390779
hash: 2d86655c452a0aa5f827ababbf9109c72bed0754675afeb768cf51979e2c8637
---

## Retry Service Documentation

This package provides a service for executing a function with automatic retries based on configurable options. It is designed to handle transient errors and improve the reliability of operations that may occasionally fail.

**Key Types and Interfaces**

*   **`RetryOptions`**: This struct configures the retry behavior. It contains the following fields:
    *   `MaxRetries`:  An integer defining the maximum number of retry attempts.
    *   `BackoffMs`: An integer specifying the delay in milliseconds between retry attempts.
    *   `ShouldRetry`: A function that takes an error as input and returns a boolean indicating whether a retry should be attempted for that specific error. If nil, all errors are considered retryable.
*   **`testError`**: A custom error type used in the tests to simulate different error scenarios. It implements the `error` interface.

**Important Functions**

*   **`DefaultRetryOptions()`**: This function returns a `RetryOptions` struct with default values. The defaults are: `MaxRetries` set to 3, `BackoffMs` set to 1000 milliseconds, and `ShouldRetry` set to a non-nil function (implicitly allowing all errors to be retried).
*   **`ExecuteWithRetry(fn func() (string, error), opts *RetryOptions)`**: This is the core function of the package. It accepts a function `fn` to execute and optional `RetryOptions`.  It repeatedly executes `fn` until either:
    *   `fn` returns a successful result (no error).
    *   The maximum number of retries (`MaxRetries`) is reached.
    *   The `ShouldRetry` function returns `false` for the error returned by `fn`.

    The function returns the successful result from `fn` and a `nil` error if successful. If retries are exhausted or `ShouldRetry` returns `false`, it returns the last error encountered.

**Error Handling**

The `ExecuteWithRetry` function handles errors returned by the provided function `fn`. It respects the `ShouldRetry` function to determine if an error warrants a retry. If the maximum number of retries is reached or `ShouldRetry` returns `false`, the last error encountered is returned to the caller.  The package uses a custom `testError` type for testing purposes, allowing for simulation of retryable and non-retryable errors.

**Concurrency**

This package does not explicitly use goroutines or channels. The retries are performed sequentially within the `ExecuteWithRetry` function. The backoff delay is implemented using `time.Sleep`.

**Design Decisions**

*   **Configuration via Options**: The retry behavior is configurable through the `RetryOptions` struct, providing flexibility without requiring multiple function overloads.
*   **`ShouldRetry` Function**: The `ShouldRetry` function allows for fine-grained control over which errors should trigger a retry, enabling the handling of specific error conditions.
*   **String Return Type**: The example function signature in the tests returns a string. This is for demonstration and can be easily adapted to any return type.
*   **Backoff Strategy**: A simple exponential backoff is implemented using `time.Sleep` with the configured `BackoffMs` value.