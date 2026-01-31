---
type: Documentation
domain: runtime
origin: packages/runtime/internal/services/retry_test.go
last_modified: 2026-01-31
generated: true
source: packages/runtime/internal/services/retry_test.go
generated_at: 2026-01-31T10:06:14.828237
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

The `ExecuteWithRetry` function handles errors returned by the provided function `fn`. It respects the `ShouldRetry` function, allowing you to define specific error types that should or should not trigger a retry. If no `ShouldRetry` function is provided, all errors are considered retryable.  The function ultimately returns the last error encountered if the maximum number of retries is exhausted.

**Concurrency:**

This package does not explicitly employ goroutines or channels. The retries are performed sequentially within the `ExecuteWithRetry` function. The `time.Sleep` function is used to introduce delays between retries, but this does not involve concurrent execution.

**Design Decisions:**

*   **Configuration via Options:** We chose to use a `RetryOptions` struct to provide a flexible and extensible way to configure the retry behavior. This allows you to customize the retry mechanism without modifying the core function.
*   **`ShouldRetry` Function:** The inclusion of a `ShouldRetry` function provides fine-grained control over which errors trigger retries. This is important for avoiding infinite retry loops on non-transient errors.
*   **String Return Type:** The example function signature uses a string return type for simplicity. You can adapt this to any return type as needed.
*   **Backoff Strategy:** A simple exponential backoff is implemented using `time.Sleep`. More sophisticated backoff strategies could be added in the future.