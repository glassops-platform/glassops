---
type: Documentation
domain: runtime
origin: packages/runtime/internal/services/retry.go
last_modified: 2026-02-01
generated: true
source: packages/runtime/internal/services/retry.go
generated_at: 2026-02-01T19:45:15.935130
hash: e1d1e95b765287a86618a65e5ff3ad75f02277e048316a35ff564620915da90a
---

## Retry Service Documentation

This document describes the `retry` service, a component designed to provide resilient execution of functions that may fail transiently. It implements an exponential backoff and retry mechanism.

**Package Responsibilities**

The `services` package provides reusable services for common operational concerns. The `retry` service specifically handles executing a given function multiple times with increasing delays between attempts, based on configurable options.

**Key Types**

*   **`RetryOptions`**: This struct configures the retry behavior.
    *   `MaxRetries int`:  The maximum number of times to retry the function.
    *   `BackoffMs int`: The initial delay in milliseconds before the first retry.  Subsequent delays increase exponentially.
    *   `ShouldRetry func(error) bool`: An optional function that determines whether a retry should be attempted based on the error returned by the function. If nil, all errors trigger a retry.

**Functions**

*   **`DefaultRetryOptions() RetryOptions`**: This function returns a `RetryOptions` struct with sensible default values. The defaults are:
    *   `MaxRetries = 3`
    *   `BackoffMs = 1000` (1 second)
    *   `ShouldRetry` always returns true, meaning all errors will trigger a retry.

*   **`ExecuteWithRetry[T any](fn func() (T, error), opts *RetryOptions) (T, error)`**: This is the core function of the service. It executes the provided function `fn` with retry logic.
    *   `fn func() (T, error)`: The function to execute. It should return a value of type `T` and an error.
    *   `opts *RetryOptions`: A pointer to a `RetryOptions` struct that configures the retry behavior. If `opts` is nil, the function uses the default options returned by `DefaultRetryOptions()`.
    *   Returns: The result of the function `fn` if successful, and nil error. If all retries fail, it returns the zero value of type `T` and the last error encountered.

**Error Handling**

The `ExecuteWithRetry` function handles errors returned by the provided function `fn`. It checks if the `ShouldRetry` function is defined. If it is, the function is called with the error. If `ShouldRetry` returns `true`, a retry is attempted. If `ShouldRetry` returns `false` or is nil, the function immediately returns the error. The last error encountered during all retry attempts is returned if the function ultimately fails.

**Concurrency**

This service does not directly employ goroutines or channels. The `time.Sleep` function is used to introduce delays between retries, but this is done sequentially within a single goroutine.

**Design Decisions**

*   **Exponential Backoff:** The delay between retries increases exponentially (`opts.BackoffMs * (1 << attempt)`). This helps to avoid overwhelming a failing service with repeated requests.
*   **`ShouldRetry` Function:** The `ShouldRetry` function provides flexibility in determining which errors should trigger a retry. This allows You to handle specific error conditions differently.
*   **Zero Value on Failure:** Returning the zero value of type `T` on failure provides a consistent way to indicate that the function did not succeed in producing a valid result.
*   **Options Pattern:** The use of a `RetryOptions` struct allows for easy configuration of the retry behavior without requiring a large number of function parameters.