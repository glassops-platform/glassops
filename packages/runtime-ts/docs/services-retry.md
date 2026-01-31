---
type: Documentation
domain: runtime-ts
origin: packages/runtime-ts/src/services/retry.ts
last_modified: 2026-01-31
generated: true
source: packages/runtime-ts/src/services/retry.ts
generated_at: 2026-01-31T10:14:18.696162
hash: 66a78c21d7ac4ddc3abd59ad36043dc7099c50667d80546a7689ee9433fea96b
---

## Retry Service Documentation

This document details the functionality of the retry service, designed to handle transient failures in asynchronous operations through exponential backoff.

### Overview

The retry service provides a mechanism to automatically re-execute an asynchronous function if it fails, with increasing delays between attempts. This is particularly useful when interacting with external services that may experience temporary outages or rate limiting.

### RetryOptions Interface

The behavior of the retry service is configurable through the `RetryOptions` interface:

*   `maxRetries`: (Optional) The maximum number of retry attempts. Defaults to 3.
*   `backoffMs`: (Optional) The initial delay in milliseconds before the first retry. The delay increases exponentially with each subsequent attempt. Defaults to 1000ms (1 second).
*   `shouldRetry`: (Optional) A function that determines whether a retry should be attempted for a given error. It accepts an `Error` object as input and returns a boolean value. If not provided, all errors will trigger a retry.

### executeWithRetry Function

The core functionality is provided by the `executeWithRetry` function.

**Signature:**

```typescript
async function executeWithRetry<T>(
  fn: () => Promise<T>,
  options?: RetryOptions,
): Promise<T>
```

**Parameters:**

*   `fn`: A function that returns a `Promise`. This is the asynchronous operation to be retried.
*   `options`: (Optional) An object conforming to the `RetryOptions` interface, allowing customization of the retry behavior.

**Return Value:**

*   The result of the `fn` function if it completes successfully within the retry limit.

**Exceptions:**

*   The last error encountered if all retry attempts fail.

**Behavior:**

1.  The function attempts to execute the provided `fn`.
2.  If `fn` throws an error, the `shouldRetry` function (if provided) is called to determine if a retry is appropriate. If `shouldRetry` returns `false`, the error is immediately re-thrown.
3.  If a retry is allowed, the function waits for a specified duration before attempting `fn` again. The delay is calculated using exponential backoff: `backoffMs * 2^attempt`, where `attempt` is the current retry attempt number (starting from 0).
4.  Steps 1-3 are repeated up to `maxRetries` times.
5.  If `fn` fails after all retry attempts, the last error encountered is thrown.

**Example Usage:**

```typescript
async function myAsyncFunction(): Promise<string> {
  // Some asynchronous operation that might fail
  return "Success!";
}

async function runWithRetry() {
  try {
    const result = await executeWithRetry(myAsyncFunction, {
      maxRetries: 5,
      backoffMs: 500,
      shouldRetry: (error) => error.message.includes("temporary error"),
    });
    console.log("Operation succeeded:", result);
  } catch (error) {
    console.error("Operation failed after multiple retries:", error);
  }
}

runWithRetry();
```

In this example, `myAsyncFunction` will be retried up to 5 times with an initial delay of 500ms. A retry will only occur if the error message includes "temporary error".

### Considerations

*   The `shouldRetry` function is important for preventing infinite retry loops in cases where the error is not transient.
*   Exponential backoff helps to avoid overwhelming a failing service with repeated requests.
*   Carefully consider the appropriate values for `maxRetries` and `backoffMs` based on the specific requirements of your application and the expected behavior of the external service.