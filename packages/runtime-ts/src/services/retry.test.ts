import { executeWithRetry } from "./retry";

describe("executeWithRetry", () => {
  beforeEach(() => {
    jest.useFakeTimers();
  });

  afterEach(() => {
    jest.useRealTimers();
  });

  it("should return result on first successful attempt", async () => {
    const fn = jest.fn().mockResolvedValue("success");

    const promise = executeWithRetry(fn);
    const result = await promise;

    expect(result).toBe("success");
    expect(fn).toHaveBeenCalledTimes(1);
  });

  it("should retry on failure and succeed on subsequent attempt", async () => {
    const fn = jest
      .fn()
      .mockRejectedValueOnce(new Error("transient failure"))
      .mockResolvedValueOnce("success after retry");

    const promise = executeWithRetry(fn, { maxRetries: 3, backoffMs: 100 });

    // Fast-forward through the backoff
    await jest.advanceTimersByTimeAsync(100);

    const result = await promise;
    expect(result).toBe("success after retry");
    expect(fn).toHaveBeenCalledTimes(2);
  });

  it("should throw after exhausting all retries", async () => {
    const fn = jest.fn().mockRejectedValue(new Error("persistent failure"));

    const promise = executeWithRetry(fn, { maxRetries: 3, backoffMs: 100 });
    // Prevent unhandled rejection warning during timer advancement
    promise.catch(() => {});

    // First retry backoff (100ms)
    await jest.advanceTimersByTimeAsync(100);
    await Promise.resolve(); // Flash microtasks

    // Second retry backoff (200ms)
    await jest.advanceTimersByTimeAsync(200);
    await Promise.resolve();

    await expect(promise).rejects.toThrow("persistent failure");
    expect(fn).toHaveBeenCalledTimes(3);
  });

  it("should use exponential backoff", async () => {
    const fn = jest.fn().mockRejectedValue(new Error("failure"));
    const backoffMs = 1000;

    executeWithRetry(fn, { maxRetries: 4, backoffMs }).catch(() => {});

    // First attempt is immediate
    expect(fn).toHaveBeenCalledTimes(1);

    // After 1000ms (2^0 * 1000), second attempt
    await jest.advanceTimersByTimeAsync(1000);
    await Promise.resolve();
    expect(fn).toHaveBeenCalledTimes(2);

    // After 2000ms (2^1 * 1000), third attempt
    await jest.advanceTimersByTimeAsync(2000);
    await Promise.resolve();
    expect(fn).toHaveBeenCalledTimes(3);

    // After 4000ms (2^2 * 1000), fourth attempt
    await jest.advanceTimersByTimeAsync(4000);
    await Promise.resolve();
    expect(fn).toHaveBeenCalledTimes(4);
  });

  it("should handle maxRetries=0 (no retries)", async () => {
    const fn = jest.fn().mockRejectedValue(new Error("failure"));
    await expect(executeWithRetry(fn, { maxRetries: 0 })).rejects.toThrow("Max retries exceeded");
    // Should verify it wasn't called? Or maybe maxRetries implies attempts? 
    // Usually maxRetries includes the initial attempt? No, retries usually means *re*-tries.
    // My code loop: for (let attempt = 0; attempt < maxRetries; attempt++)
    // So maxRetries is "maxAttempts".
    // If maxRetries is 0, loop doesn't run.
    expect(fn).not.toHaveBeenCalled();
  });

  it("should respect shouldRetry predicate", async () => {
    const nonRetryableError = new Error("DO_NOT_RETRY");
    const fn = jest.fn().mockRejectedValue(nonRetryableError);

    const shouldRetry = jest.fn((err: Error) => !err.message.includes("DO_NOT_RETRY"));

    await expect(
      executeWithRetry(fn, { maxRetries: 3, shouldRetry }),
    ).rejects.toThrow("DO_NOT_RETRY");

    // Should only attempt once since shouldRetry returns false
    expect(fn).toHaveBeenCalledTimes(1);
    expect(shouldRetry).toHaveBeenCalledWith(nonRetryableError);
  });

  it("should handle non-Error throws", async () => {
    const fn = jest.fn().mockRejectedValue("string error");

    const promise = executeWithRetry(fn, { maxRetries: 1 });

    await expect(promise).rejects.toThrow("string error");
  });

  it("should use default options when none provided", async () => {
    const fn = jest.fn().mockResolvedValue("result");

    const result = await executeWithRetry(fn);

    expect(result).toBe("result");
  });
});
