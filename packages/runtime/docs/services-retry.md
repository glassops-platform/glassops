---
type: Documentation
domain: runtime
last_modified: 2026-02-02
generated: true
source: packages/runtime/internal/services/retry.go
generated_at: 2026-02-02T22:40:52.904412
hash: e1d1e95b765287a86618a65e5ff3ad75f02277e048316a35ff564620915da90a
---

## Retry Service Documentation

This document describes the `retry` service, a component designed to provide resilient execution of functions that may fail transiently. We aim to simplify the implementation of retry logic within applications.

**Package Responsibilities**

The `services` package provides a mechanism for automatically retrying function calls with configurable backoff and retry conditions. It handles the complexities of exponential backoff and allows developers to define custom logic for determining whether a retry is appropriate.

**Key Types**

*   **`RetryOptions`**: This struct configures the retry behavior.
    *   `MaxRetries int`:  Specifies the maximum number of times the function will be retried.
    *   `BackoffMs int`: Defines the initial backoff duration in milliseconds. The actual delay increases exponentially with each retry attempt.
    *   `ShouldRetry func(error) bool`: A function that determines whether a retry should be attempted based on the error returned by the function being executed. If nil, all errors will trigger a retry.

**Functions**

*   **`DefaultRetryOptions() RetryOptions`**: This function returns a `RetryOptions` struct initialized with sensible default values. The defaults are: `MaxRetries` set to 3, `BackoffMs` set to 1000 milliseconds, and `ShouldRetry` set to a function that always returns true (meaning retry on any error).

*   **`ExecuteWithRetry[T any](fn func() (T, error), opts *RetryOptions) (T, error)`**: This is the core function of the service. It executes the provided function `fn` with retry logic.
    *   `fn func() (T, error)`: The function to be executed. It should return a value of type `T` and an error.
    *   `opts *RetryOptions`: A pointer to a `RetryOptions` struct that configures the retry behavior. If `opts` is nil, the function uses the default options returned by `DefaultRetryOptions()`.
    *   Returns: The result of the function `fn` if it succeeds, and a nil error. If the function fails after all retries, it returns the zero value of type `T` and the last error encountered.

**Error Handling**

The `ExecuteWithRetry` function handles errors returned by the provided function `fn`. It stores the last error encountered and returns it if all retry attempts fail. The `ShouldRetry` function allows you to customize which errors should trigger a retry. If `ShouldRetry` is nil, all errors will result in a retry attempt.

**Concurrency**

This service does not directly employ goroutines or channels. The `time.Sleep` function introduces pauses between retries, but these pauses occur within a single goroutine.

**Design Decisions**

*   **Exponential Backoff:** We chose exponential backoff to avoid overwhelming the target service with repeated requests immediately after a failure. This strategy provides increasing delays between retries, giving the service time to recover.
*   **Configurable Retry Logic:** The `ShouldRetry` function provides flexibility in determining when a retry is appropriate. This allows you to handle specific error conditions differently.
*   **Zero Value on Failure:** Returning the zero value of type `T` on failure provides a clear indication that the function did not succeed. You should check the returned error to determine the cause of the failure.
*   **Options Pattern:** The use of a `RetryOptions` struct and a default options function promotes configurability and allows for easy extension of the retry behavior in the future.

**Usage Example**

You can use this service as follows:

```go
result, err := ExecuteWithRetry(func() (string, error) {
    // Your function that might fail
    return "Success", nil // Or return an error
}, &RetryOptions{MaxRetries: 5, BackoffMs: 500})

if err != nil {
    // Handle the error
} else {
    // Process the result
}