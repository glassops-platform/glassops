---
type: Documentation
domain: runtime
last_modified: 2026-02-02
generated: true
source: packages/runtime/internal/services/retry_test.go
generated_at: 2026-02-02T22:41:10.949326
hash: 2d86655c452a0aa5f827ababbf9109c72bed0754675afeb768cf51979e2c8637
---

## Retry Service Documentation

This package provides a service for executing operations with automatic retries, handling transient failures and implementing configurable backoff strategies. It is designed to improve the resilience of operations that may occasionally fail due to temporary issues.

**Key Types:**

*   `RetryOptions`: This type encapsulates the configuration for the retry mechanism.
    *   `MaxRetries`: An integer defining the maximum number of retry attempts.
    *   `BackoffMs`: An integer specifying the initial delay in milliseconds between retries. The delay increases exponentially with each subsequent attempt.
    *   `ShouldRetry`: A function that takes an error as input and returns a boolean indicating whether the operation should be retried for that specific error. If this is nil, all errors are retried.

*   `testError`: A custom error type used in the tests to simulate different error scenarios. It implements the `error` interface.

**Functions:**

*   `DefaultRetryOptions()`: This function returns a `RetryOptions` struct with default values. The defaults are: `MaxRetries` set to 3, `BackoffMs` set to 1000 milliseconds, and `ShouldRetry` set to nil (meaning all errors will be retried).

*   `ExecuteWithRetry(operation func() (string, error), options *RetryOptions)`: This is the core function of the package. It accepts an operation (a function that returns a string and an error) and optional `RetryOptions`.
    *   It executes the provided `operation`.
    *   If the operation returns an error, it checks if the error should be retried based on the `ShouldRetry` function in the `options`.
    *   If the error should be retried, it waits for the specified `BackoffMs` duration (increasing exponentially with each attempt) and retries the operation up to `MaxRetries` times.
    *   If the operation eventually succeeds, it returns the result and a nil error.
    *   If the operation fails after exhausting all retry attempts, it returns an error.

**Error Handling:**

The `ExecuteWithRetry` function handles errors returned by the provided operation. The `ShouldRetry` function within the `RetryOptions` allows for fine-grained control over which errors trigger a retry. If `ShouldRetry` is nil, all errors are considered retryable. The package does not perform any specific error type checking beyond what is defined by the `ShouldRetry` function.

**Concurrency:**

This package does not explicitly employ goroutines or channels. The retries are performed sequentially within the `ExecuteWithRetry` function. The `time.Sleep` function is used to introduce delays between retries, but this does not involve concurrent execution.

**Design Decisions:**

*   **Configuration via Options:** The use of `RetryOptions` allows for flexible configuration of the retry behavior without requiring changes to the core `ExecuteWithRetry` function.
*   **Functional Approach:** The `operation` is passed as a function, promoting a functional programming style and making the retry service reusable with different operations.
*   **Exponential Backoff:** The backoff strategy increases the delay between retries, reducing the load on the failing service and potentially allowing it to recover.
*   **`ShouldRetry` Function:** This function provides a powerful mechanism for controlling which errors are retried, preventing infinite loops for non-transient failures.

**Usage:**

You can use this package to wrap operations that may be prone to transient failures. For example:

```go
result, err := ExecuteWithRetry(func() (string, error) {
    // Your operation here
    return "data", nil // or return "", error
}, &RetryOptions{
    MaxRetries: 5,
    BackoffMs: 500,
    ShouldRetry: func(err error) bool {
        // Check if the error is retryable
        return true
    },
})

if err != nil {
    // Handle the error
} else {
    // Process the result
}