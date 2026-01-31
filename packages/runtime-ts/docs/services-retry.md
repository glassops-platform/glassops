---
type: Documentation
domain: runtime-ts
origin: packages/runtime-ts/src/services/retry.ts
last_modified: 2026-01-31
generated: true
source: packages/runtime-ts/src/services/retry.ts
generated_at: 2026-01-31T09:17:47.397771
hash: 66a78c21d7ac4ddc3abd59ad36043dc7099c50667d80546a7689ee9433fea96b
---

## Retry Service Documentation

This document details the functionality of the retry service, designed to handle transient failures in asynchronous operations through exponential backoff.

**Purpose**

The retry service provides a mechanism to automatically re-execute an asynchronous function if it fails, with increasing delays between attempts. This improves application resilience when interacting with unreliable services or resources.

**RetryOptions Interface**

The behavior of the retry service is configurable through the `RetryOptions` interface:

*   `maxRetries`: (Optional) The maximum number of retry attempts. Defaults to 3.
*   `backoffMs`: (Optional) The initial delay in milliseconds before the first retry. The delay increases exponentially with each subsequent attempt. Defaults to 1000ms (1 second).
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
      shouldRetry: (error) => {
        // Only retry if the error is a network error
        return error.message.includes("network");
      },
    });
    console.log("Operation succeeded:", result);
  } catch (error) {
    console.error("Operation failed after multiple retries:", error);
  }
}

runWithRetry();
```

**Behavior Details**

*   **Exponential Backoff:** The delay between retries increases exponentially. For example, with a `backoffMs` of 1000, the delays will be 1000ms, 2000ms, 4000ms, and so on.
*   **Last Error:** The service preserves the last error encountered during the retry process. This ensures that the caller receives the most informative error message when all retries have been exhausted.
*   **Default Options:** If you do not provide any options, the service will use default values for `maxRetries` (3) and `backoffMs` (1000ms). The default `shouldRetry` function always returns `true`, meaning all errors will trigger a retry.