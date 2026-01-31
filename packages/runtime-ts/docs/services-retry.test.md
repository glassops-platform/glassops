---
type: Documentation
domain: runtime-ts
origin: packages/runtime-ts/src/services/retry.test.ts
last_modified: 2026-01-31
generated: true
source: packages/runtime-ts/src/services/retry.test.ts
generated_at: 2026-01-31T09:17:29.383022
hash: 8f8177f9522970e0c73f095aa11dee7604a2394058095c6d80ef9136115711a4
---

## executeWithRetry Service Documentation

This document details the functionality of the `executeWithRetry` service, designed to reliably execute asynchronous operations that may encounter transient failures.

**Purpose**

The `executeWithRetry` service automatically retries a provided asynchronous function (a function that returns a Promise) if it fails, offering configurable control over the retry behavior. This enhances application resilience by mitigating the impact of temporary issues.

**Functionality**

The `executeWithRetry` function accepts an asynchronous function and an optional configuration object. It attempts to execute the function and, upon failure, retries based on the provided configuration.

**Parameters**

*   **`fn`**:  The asynchronous function (returning a Promise) to be executed.
*   **`options`** (Optional): An object configuring the retry behavior.  Available options include:
    *   **`maxRetries`**:  The maximum number of retry attempts. A value of 0 disables retries. Defaults to 3.
    *   **`backoffMs`**: The initial delay in milliseconds between retries. The delay increases exponentially with each subsequent attempt. Defaults to 1000ms (1 second).
    *   **`shouldRetry`**: A predicate function that determines whether a retry should be attempted based on the error encountered. It receives the error object as input and returns `true` to retry, or `false` to stop. If not provided, all errors will trigger a retry attempt.

**Return Value**

*   If the function succeeds on any attempt, `executeWithRetry` returns the resolved value of the Promise returned by the function.
*   If the function fails after exhausting all retry attempts, `executeWithRetry` rejects with the last error encountered.

**Behavior**

*   **Retry Logic:**  The service implements an exponential backoff strategy.  The delay between retries increases exponentially (2<sup>n</sup> * `backoffMs`), where n is the retry attempt number (starting from 0).
*   **Error Handling:** The service propagates the final error if all retry attempts fail. It also handles cases where the function throws a non-Error value.
*   **Predicate Control:** The `shouldRetry` function allows for selective retries based on the type of error encountered. This is useful for avoiding retries on errors that are unlikely to be resolved by repeating the operation.
*   **Zero Retries:** When `maxRetries` is set to 0, the function is executed only once. If it fails, the Promise is immediately rejected.
*   **Default Options:** If no options are provided, the service uses default values for `maxRetries` and `backoffMs`.



**Usage**

You can use `executeWithRetry` as follows:

```typescript
const result = await executeWithRetry(async () => {
  // Your asynchronous operation here
  return "Operation Result";
}, { maxRetries: 5, backoffMs: 500 });

console.log(result); // Output: Operation Result (or an error if all retries fail)