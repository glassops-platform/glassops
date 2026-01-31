---
type: Documentation
domain: runtime
origin: packages/runtime/internal/services/retry.go
last_modified: 2026-01-31
generated: true
source: packages/runtime/internal/services/retry.go
generated_at: 2026-01-31T10:05:58.599464
hash: e1d1e95b765287a86618a65e5ff3ad75f02277e048316a35ff564620915da90a
---

## Retry Service Documentation

This document describes the `retry` service, a component designed to provide resilient execution of functions that may fail transiently. It implements an exponential backoff and retry mechanism.

**Package Responsibilities**

The `services` package provides core functionalities for managing application services. The `retry` service specifically handles the execution of operations with automatic retries based on configurable parameters. This helps improve the robustness of applications by automatically handling temporary failures.

**Key Types**

*   **`RetryOptions`**: This struct configures the retry behavior.
    *   `MaxRetries int`:  Specifies the maximum number of retry attempts.
    *   `BackoffMs int`: Defines the initial delay in milliseconds before the first retry. The delay increases exponentially with each subsequent attempt.
    *   `ShouldRetry func(error) bool`: A function that determines whether a retry should be attempted based on the error encountered. If nil, all errors will trigger a retry.

**Functions**

*   **`DefaultRetryOptions() RetryOptions`**: This function returns a `RetryOptions` struct initialized with default values. The defaults are: `MaxRetries` set to 3, `BackoffMs` set to 1000 milliseconds, and `ShouldRetry` set to a function that always returns true (meaning retry on any error).

*   **`ExecuteWithRetry[T any](fn func() (T, error), opts *RetryOptions) (T, error)`**: This is the core function for executing a function with retry logic.
    *   `fn func() (T, error)`: The function to be executed. It should return a value of type `T` and an error.
    *   `opts *RetryOptions`: A pointer to a `RetryOptions` struct that configures the retry behavior. If `opts` is nil, the function uses the default options returned by `DefaultRetryOptions()`.
    *   Returns: The result of the function `fn` if successful, along with a nil error. If all retry attempts fail, it returns the zero value of type `T` and the last error encountered.

**Error Handling**

The `ExecuteWithRetry` function handles errors returned by the provided function `fn`. It checks if the `ShouldRetry` function (if provided) returns true for the given error. If it does, the function retries the operation. If `ShouldRetry` is nil, or returns true, the function will retry. If `ShouldRetry` returns false, the function immediately returns the error, halting the retry process. The last error encountered during all retry attempts is returned if the operation ultimately fails.

**Concurrency**

This service does not directly employ goroutines or channels. However, the `time.Sleep` function introduces a delay between retry attempts, allowing other operations to proceed.

**Design Decisions**

*   **Exponential Backoff:** The use of exponential backoff (`opts.BackoffMs*(1<<attempt)`) is a standard practice for retry mechanisms. It helps avoid overwhelming a failing service with repeated requests in quick succession.
*   **`ShouldRetry` Function:** The `ShouldRetry` function provides flexibility in determining which errors should trigger a retry. This allows you to handle specific error conditions differently.
*   **Zero Value on Failure:** Returning the zero value of type `T` on failure is a common Go idiom for indicating that a valid result could not be obtained.
*   **Options Pattern:** The use of a `RetryOptions` struct and a pointer to it allows for easy configuration and customization of the retry behavior. Providing default options simplifies usage when only basic retry functionality is needed.