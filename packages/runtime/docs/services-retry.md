---
type: Documentation
domain: runtime
origin: packages/runtime/internal/services/retry.go
last_modified: 2026-01-31
generated: true
source: packages/runtime/internal/services/retry.go
generated_at: 2026-01-31T09:09:57.231928
hash: e1d1e95b765287a86618a65e5ff3ad75f02277e048316a35ff564620915da90a
---

## Retry Service Documentation

This document describes the `retry` service, a component designed to provide resilient execution of functions that may fail transiently. We aim to simplify the implementation of retry logic within applications.

**Package Responsibilities**

The `services` package provides a mechanism for automatically retrying function calls with configurable backoff and retry conditions. It handles the complexities of exponential backoff and allows developers to define custom logic for determining whether a retry is appropriate.

**Key Types**

*   **`RetryOptions`**: This struct configures the retry behavior.
    *   `MaxRetries`:  An integer defining the maximum number of retry attempts.
    *   `BackoffMs`: An integer specifying the initial backoff duration in milliseconds. The actual delay increases exponentially with each attempt.
    *   `ShouldRetry`: A function that takes an error as input and returns a boolean.  If true, the function will be retried; otherwise, the error is returned.

**Functions**

*   **`DefaultRetryOptions()`**: This function returns a `RetryOptions` struct initialized with sensible default values: `MaxRetries` set to 3, `BackoffMs` set to 1000 milliseconds, and `ShouldRetry` set to a function that always returns true (meaning all errors will trigger a retry).

*   **`ExecuteWithRetry[T any](fn func() (T, error), opts *RetryOptions) (T, error)`**: This is the core function of the service. It accepts a function `fn` to execute and a pointer to `RetryOptions` to configure the retry behavior.
    *   `fn`: The function to be executed. It should return a value of type `T` and an error.
    *   `opts`: A pointer to a `RetryOptions` struct. If `opts` is nil, the function uses the default retry options.
    *   The function executes `fn` up to `MaxRetries` times.
    *   If `fn` returns an error, the `ShouldRetry` function (if provided) is called to determine if a retry is warranted.
    *   Between retries, the function pauses for a duration calculated using exponential backoff, based on the `BackoffMs` value. The delay increases with each attempt.
    *   If `fn` eventually succeeds, the function returns the result and a nil error.
    *   If `fn` fails after all retry attempts, the function returns a zero value of type `T` and the last error encountered.

**Error Handling**

The `ExecuteWithRetry` function handles errors returned by the provided function `fn`. It allows you to control which errors trigger a retry using the `ShouldRetry` function. If `ShouldRetry` is not provided, all errors will result in a retry. The last error encountered is returned if all retries fail.

**Concurrency**

This service does not directly employ goroutines or channels. The `time.Sleep` function introduces pauses between retries, but the retry logic itself is sequential within a single goroutine.

**Design Decisions**

*   **Generics**: We use generics (`[T any]`) to allow the `ExecuteWithRetry` function to work with functions that return any type.
*   **Options Pattern**: The `RetryOptions` struct and the use of a pointer to this struct allow for flexible configuration of the retry behavior.  Providing default options simplifies usage when custom configuration is not needed.
*   **Exponential Backoff**: Exponential backoff is used to avoid overwhelming a failing service with repeated requests.
*   **`ShouldRetry` Function**: The `ShouldRetry` function provides a customizable way to determine whether an error is transient and warrants a retry. You can implement specific logic to handle different error types.