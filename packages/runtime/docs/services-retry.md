---
type: Documentation
domain: runtime
origin: packages/runtime/internal/services/retry.go
last_modified: 2026-01-29
generated: true
source: packages/runtime/internal/services/retry.go
generated_at: 2026-01-29T21:28:55.282686
hash: e1d1e95b765287a86618a65e5ff3ad75f02277e048316a35ff564620915da90a
---

## Retry Service Documentation

This document details the `retry` service, a component designed to provide resilient execution of functions that may fail transiently. It implements an exponential backoff and retry mechanism.

**Package Responsibilities:**

The `services` package provides functionality for executing operations with automatic retries. This is particularly useful for interacting with external systems or performing operations that are susceptible to temporary failures. The retry service aims to improve the reliability of applications by automatically handling transient errors.

**Key Types:**

*   **`RetryOptions`**: This struct configures the retry behavior.
    *   `MaxRetries`: An integer defining the maximum number of retry attempts.
    *   `BackoffMs`: An integer specifying the initial backoff duration in milliseconds. The delay between retries increases exponentially.
    *   `ShouldRetry`: A function that determines whether a retry should be attempted for a given error. It accepts an `error` as input and returns a boolean. If nil, all errors will trigger a retry.

**Functions:**

*   **`DefaultRetryOptions()`**: This function returns a `RetryOptions` struct initialized with default values. The defaults are: `MaxRetries` set to 3, `BackoffMs` set to 1000 milliseconds, and `ShouldRetry` set to a function that always returns true (meaning all errors will be retried).

*   **`ExecuteWithRetry\[T any](fn func() (T, error), opts *RetryOptions) (T, error)`**: This is the core function of the service. It executes the provided function `fn` with retry logic.
    *   `fn`: A function that performs the operation to be retried. It should return a value of type `T` and an `error`.
    *   `opts`: A pointer to a `RetryOptions` struct that configures the retry behavior. If `opts` is nil, the function uses the default options returned by `DefaultRetryOptions()`.
    *   Returns: The result of the function `fn` (of type `T`) and an error. If the function succeeds on any attempt, the result and a nil error are returned. If all retries fail, a zero value of type `T` and the last encountered error are returned.

**Error Handling:**

The `ExecuteWithRetry` function handles errors returned by the provided function `fn`. It checks if the `ShouldRetry` function (if provided) returns true for the error. If it does, the function retries the operation. If `ShouldRetry` is nil, all errors trigger a retry. If `ShouldRetry` returns false, the function immediately returns the error, halting further retries. The last encountered error is always returned if all retry attempts fail.

**Concurrency:**

This service does not directly employ goroutines or channels. However, the `time.Sleep` function introduces a pause between retries, allowing other operations to proceed.

**Design Decisions:**

*   **Exponential Backoff:** The use of exponential backoff helps to avoid overwhelming a failing service with repeated requests. The delay increases with each attempt, giving the service time to recover.
*   **Configurable Retry Logic:** The `RetryOptions` struct allows you to customize the retry behavior to suit your specific needs. You can control the maximum number of retries, the initial backoff duration, and the conditions under which a retry should be attempted.
*   **Generic Function:** The `ExecuteWithRetry` function is a generic function, allowing it to work with functions that return any type of value. This makes the service reusable across a wide range of applications.
*   **Zero Value on Failure:** Returning the zero value of the generic type `T` on failure provides a consistent way to indicate that the operation failed to produce a valid result. You should check the returned error to confirm the failure.