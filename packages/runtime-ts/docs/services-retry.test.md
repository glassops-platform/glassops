---
type: Documentation
domain: runtime-ts
origin: packages/runtime-ts/src/services/retry.test.ts
last_modified: 2026-01-31
generated: true
source: packages/runtime-ts/src/services/retry.test.ts
generated_at: 2026-01-31T10:13:53.584589
hash: 8f8177f9522970e0c73f095aa11dee7604a2394058095c6d80ef9136115711a4
---

## executeWithRetry Service Documentation

This document details the functionality of the `executeWithRetry` service, designed to reliably execute asynchronous operations that may fail transiently.

**Purpose**

The `executeWithRetry` service automatically retries a provided asynchronous function if it fails, offering configurable control over the number of attempts and the delay between them. This improves application resilience when interacting with potentially unstable services or resources.

**Function Signature**

```typescript
function executeWithRetry<T>(
  fn: () => Promise<T>,
  options?: RetryOptions
): Promise<T>
```

**Parameters**

*   `fn`: A function that returns a Promise. This is the operation to be retried.
*   `options`: (Optional) An object configuring the retry behavior.

**RetryOptions Configuration**

The `options` parameter accepts the following properties:

*   `maxRetries`: (Number, default: 5) The maximum number of retry attempts.  The initial attempt is *not* included in this count.
*   `backoffMs`: (Number, default: 1000) The initial delay in milliseconds between retries.  The delay increases exponentially with each subsequent attempt.
*   `shouldRetry`: (Function, optional) A predicate function that determines whether a retry should be attempted. It receives the error thrown by `fn` as an argument and returns `true` to retry, or `false` to stop. If not provided, all errors will trigger a retry, up to `maxRetries`.

**Return Value**

*   If the function `fn` succeeds on any attempt, `executeWithRetry` returns a Promise that resolves with the successful result of `fn`.
*   If the function `fn` fails after exhausting all retry attempts (or if `shouldRetry` returns `false`), `executeWithRetry` returns a Promise that rejects with the last error thrown by `fn`.

**Behavioral Notes**

*   **Exponential Backoff:** The delay between retries increases exponentially. The first retry occurs after `backoffMs`, the second after `backoffMs * 2`, the third after `backoffMs * 4`, and so on.
*   **Error Handling:** The service propagates the original error from the function `fn` when all retries are exhausted. It also handles cases where `fn` throws a non-Error value.
*   **Zero Retries:** If `maxRetries` is set to 0, the function `fn` is executed only once. If it fails, the Promise is immediately rejected.
*   **Predicate Control:** The `shouldRetry` function allows for fine-grained control over which errors trigger a retry. This is useful for handling errors that are not transient and should not be retried.
*   **Default Options:** If no `options` object is provided, the service uses default values for `maxRetries` and `backoffMs`.



**Example Usage**

```typescript
// Example with default options
const result = await executeWithRetry(async () => {
  // Some asynchronous operation
  return "Success!";
});

// Example with custom options
const result = await executeWithRetry(async () => {
  // Some asynchronous operation
}, { maxRetries: 3, backoffMs: 500 });

// Example with a retry predicate
const result = await executeWithRetry(async () => {
  // Some asynchronous operation
}, {
  maxRetries: 2,
  shouldRetry: (error) => error.message !== "NonRetryableError"
});