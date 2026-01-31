/**
 * Retry utility with exponential backoff for transient failures.
 */

export interface RetryOptions {
  maxRetries?: number;
  backoffMs?: number;
  shouldRetry?: (error: Error) => boolean;
}

const DEFAULT_OPTIONS: Required<Omit<RetryOptions, "shouldRetry">> = {
  maxRetries: 3,
  backoffMs: 1000,
};

/**
 * Execute an async function with exponential backoff retry logic.
 *
 * @param fn - The async function to execute
 * @param options - Retry configuration
 * @returns The result of the function
 * @throws The last error if all retries are exhausted
 */
export async function executeWithRetry<T>(
  fn: () => Promise<T>,
  options?: RetryOptions,
): Promise<T> {
  const maxRetries = options?.maxRetries ?? DEFAULT_OPTIONS.maxRetries;
  const backoffMs = options?.backoffMs ?? DEFAULT_OPTIONS.backoffMs;
  const shouldRetry = options?.shouldRetry ?? (() => true);

  let lastError: Error | undefined;

  for (let attempt = 0; attempt < maxRetries; attempt++) {
    try {
      return await fn();
    } catch (error) {
      lastError = error instanceof Error ? error : new Error(String(error));

      // Check if we should retry this specific error
      if (!shouldRetry(lastError)) {
        throw lastError;
      }

      // Don't wait after the last attempt
      if (attempt < maxRetries - 1) {
        const delay = backoffMs * Math.pow(2, attempt);
        await new Promise((resolve) => setTimeout(resolve, delay));
      }
    }
  }

  throw lastError ?? new Error("Max retries exceeded");
}
