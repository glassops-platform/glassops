---
type: Documentation
domain: runtime-ts
origin: packages/runtime-ts/src/services/retry.ts
last_modified: 2026-01-29
generated: true
source: packages/runtime-ts/src/services/retry.ts
generated_at: 2026-01-29T21:01:16.081786
hash: 66a78c21d7ac4ddc3abd59ad36043dc7099c50667d80546a7689ee9433fea96b
---

## Retry Service Documentation

This document details the functionality of the retry service, designed to handle transient failures in asynchronous operations through exponential backoff.

**Purpose**

The retry service provides a mechanism to automatically re-execute an asynchronous function if it fails, with increasing delays between attempts. This improves application resilience when interacting with unreliable services or resources.

**RetryOptions Interface**

The behavior of the retry service is configurable through the `RetryOptions` interface:

*   `maxRetries`: (Optional) The maximum number of retry attempts. Defaults to 3.
*   `backoffMs`: (Optional) The initial delay in milliseconds before the first retry. The delay increases exponentially with each subsequent attempt. Defaults to 1000 milliseconds (1 second).
*   `shouldRetry`: (Optional) A function that determines whether a retry should be attempted for a given error. It accepts an `Error` object as input and returns a boolean value. If not provided, all errors will trigger a retry.

**executeWithRetry Function**

The core functionality is provided by the `executeWithRetry` function.

*   **Parameters:**
    *   `fn`: An asynchronous function (returning a Promise) that you want to execute with retry logic.
    *   `options`: (Optional) A `RetryOptions` object to configure the retry behavior.
*   **Return Value:**
    *   The result of the `fn` function if it completes successfully within the retry limit.
*   **Error Handling:**
    *   If the `fn` function fails after all retry attempts, the last error encountered is thrown.
    *   If the `shouldRetry` function returns `false` for a specific error, that error is immediately thrown, and no further retries are attempted.

**Usage**

To use the retry service, you call `executeWithRetry`, passing in the asynchronous function you want to execute and, optionally, a `RetryOptions` object.

```typescript
async function myAsyncFunction(): Promise<string> {
  // Your asynchronous operation here
  return "Success!";
}

async function runWithRetry() {
  try {
    const result = await executeWithRetry(myAsyncFunction, {
      maxRetries: 5,
      backoffMs: 2000,
      shouldRetry: (error) => error.message.includes("temporary error"),
    });
    console.log("Operation succeeded:", result);
  } catch (error) {
    console.error("Operation failed after multiple retries:", error);
  }
}

runWithRetry();
```

In this example, `myAsyncFunction` will be retried up to 5 times with an initial delay of 2 seconds. Retries will only occur if the error message includes "temporary error".

**Implementation Details**

We implement exponential backoff by multiplying the `backoffMs` value by 2 raised to the power of the current attempt number. This ensures that the delay increases with each retry, preventing rapid re-attempts that could overload a failing service.  The `shouldRetry` function allows you to customize retry behavior based on the specific error encountered. If an error is not retryable, it is immediately propagated.