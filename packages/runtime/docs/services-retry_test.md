---
type: Documentation
domain: runtime
origin: packages/runtime/internal/services/retry_test.go
last_modified: 2026-01-29
generated: true
source: packages/runtime/internal/services/retry_test.go
generated_at: 2026-01-29T21:29:13.226368
hash: 2d86655c452a0aa5f827ababbf9109c72bed0754675afeb768cf51979e2c8637
---

## Retry Service Documentation

This package provides a service for executing operations with automatic retries, handling transient failures and implementing configurable backoff strategies. It is designed to improve the resilience of operations that may occasionally fail due to temporary issues.

**Key Types:**

*   `RetryOptions`: This struct configures the retry behavior.
    *   `MaxRetries`: An integer defining the maximum number of retry attempts.
    *   `BackoffMs`: An integer specifying the initial delay in milliseconds between retries. The delay increases exponentially with each subsequent attempt.
    *   `ShouldRetry`: A function\[err error] that determines whether a given error warrants a retry. If nil, all errors are considered retryable.

*   `testError`: A custom error type used in tests to simulate specific error conditions. It implements the `error` interface.

**Functions:**

*   `DefaultRetryOptions()`: This function returns a `RetryOptions` struct with default values. The defaults are: `MaxRetries` set to 3, `BackoffMs` set to 1000 milliseconds, and `ShouldRetry` set to nil (meaning all errors are retryable).

*   `ExecuteWithRetry\[T any](fn func() (T, error), opts *RetryOptions)`: This is the core function of the package. It executes the provided function `fn` and retries it if it returns an error, up to the configured `MaxRetries`.
    *   `fn`: A function that performs the operation to be retried. It should return a value of type `T` and an error.
    *   `opts`: A pointer to a `RetryOptions` struct that configures the retry behavior. If `opts` is nil, the default options are used.
    *   The function returns the result of the last execution of `fn` and an error. If the operation succeeds on any attempt, the error will be nil. If all retries fail, the error returned will be the error from the last attempt.

**Error Handling:**

The `ExecuteWithRetry` function handles errors returned by the provided function `fn`. It respects the `ShouldRetry` function to determine if an error should trigger a retry. If `ShouldRetry` is nil, all errors are retried.  If the maximum number of retries is reached and an error still occurs, the function returns that error.

**Concurrency:**

This package does not explicitly employ goroutines or channels for its core retry logic. Retries are performed sequentially with a delay between each attempt.

**Design Decisions:**

*   **Configurable Backoff:** The exponential backoff strategy helps to avoid overwhelming a failing service with repeated requests.
*   **`ShouldRetry` Function:**  The `ShouldRetry` function provides flexibility in determining which errors are transient and should be retried. This allows You to avoid retrying errors that are unlikely to be resolved by retrying.
*   **Default Options:** Providing default retry options simplifies usage for common scenarios.
*   **Generic Function:** The `ExecuteWithRetry` function is generic, allowing it to work with functions that return any type.