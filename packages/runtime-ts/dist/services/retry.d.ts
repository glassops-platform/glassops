/**
 * Retry utility with exponential backoff for transient failures.
 */
export interface RetryOptions {
    maxRetries?: number;
    backoffMs?: number;
    shouldRetry?: (error: Error) => boolean;
}
/**
 * Execute an async function with exponential backoff retry logic.
 *
 * @param fn - The async function to execute
 * @param options - Retry configuration
 * @returns The result of the function
 * @throws The last error if all retries are exhausted
 */
export declare function executeWithRetry<T>(fn: () => Promise<T>, options?: RetryOptions): Promise<T>;
//# sourceMappingURL=retry.d.ts.map