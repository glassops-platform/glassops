---
type: Documentation
domain: runtime
origin: packages/runtime/internal/services/retry_test.go
last_modified: 2026-02-01
generated: true
source: packages/runtime/internal/services/retry_test.go
generated_at: 2026-02-01T19:45:31.167049
hash: 2d86655c452a0aa5f827ababbf9109c72bed0754675afeb768cf51979e2c8637
---

## Retry Service Documentation

This package provides a service for executing a function with automatic retries based on configurable options. It is designed to handle transient errors and improve the reliability of operations that may occasionally fail.

**Key Types:**

*   `RetryOptions`: This type encapsulates the configuration for the retry mechanism.
    *   `MaxRetries`: An integer defining the maximum number of retry attempts. Defaults to 3.
    *   `BackoffMs`: An integer specifying the delay in milliseconds between retry attempts. Defaults to 1000.
    *   `ShouldRetry`: A function that takes an error as input and returns a boolean indicating whether a retry should be attempted for that specific error. If nil, all errors are considered retryable.

*   `testError`: A custom error type used in the tests to simulate different error scenarios. It implements the `error` interface.

**Functions:**

*   `DefaultRetryOptions()`: This function returns a `RetryOptions` struct with the default values for `MaxRetries` and `BackoffMs`, and a nil `ShouldRetry` function. You can use this as a starting point and customize the options as needed.

*   `ExecuteWithRetry(fn func() (string, error), opts *RetryOptions)`: This is the core function of the package. It accepts a function `fn` to execute and a pointer to `RetryOptions`.
    *   It repeatedly executes `fn` until either:
        *   `fn` returns a successful result (no error).
        *   The maximum number of retries (`MaxRetries`) is reached.
    *   If `ShouldRetry` is provided, it is called with the error returned by `fn` to determine if a retry should be attempted.
    *   Between retries, it waits for the duration specified by `BackoffMs`.
    *   It returns the successful result from `fn` and a nil error if successful.
    *   If all retries fail, it returns an empty string and the last error encountered.

**Error Handling:**

The `ExecuteWithRetry` function handles errors returned by the provided function `fn`. It respects the `ShouldRetry` function to determine if an error warrants a retry. If `ShouldRetry` is nil, all errors trigger a retry. When the maximum number of retries is exhausted, the last error is returned.

**Concurrency:**

This package does not explicitly employ goroutines or channels. The retries are performed sequentially within the `ExecuteWithRetry` function. The `time.Sleep` function is used to introduce delays between retries.

**Design Decisions:**

*   The function `fn` is expected to return a string and an error. This design choice simplifies the example and testing, but can be adapted to support different return types.
*   The `ShouldRetry` function provides flexibility in determining which errors should be retried. This allows you to handle specific error conditions differently.
*   The backoff mechanism uses a fixed delay between retries. More sophisticated backoff strategies (e.g., exponential backoff) could be implemented in future versions.
*   We provide a default set of retry options to offer sensible defaults while still allowing customization.