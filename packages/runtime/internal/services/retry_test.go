package services

import (
	"testing"
	"time"
)

func TestDefaultRetryOptions(t *testing.T) {
	opts := DefaultRetryOptions()

	if opts.MaxRetries != 3 {
		t.Errorf("expected MaxRetries 3, got %d", opts.MaxRetries)
	}

	if opts.BackoffMs != 1000 {
		t.Errorf("expected BackoffMs 1000, got %d", opts.BackoffMs)
	}

	if opts.ShouldRetry == nil {
		t.Error("expected ShouldRetry to be non-nil")
	}
}

func TestExecuteWithRetrySuccess(t *testing.T) {
	attempts := 0
	result, err := ExecuteWithRetry(func() (string, error) {
		attempts++
		return "success", nil
	}, nil)

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	if result != "success" {
		t.Errorf("expected 'success', got '%s'", result)
	}

	if attempts != 1 {
		t.Errorf("expected 1 attempt, got %d", attempts)
	}
}

func TestExecuteWithRetryFailThenSucceed(t *testing.T) {
	attempts := 0
	result, err := ExecuteWithRetry(func() (string, error) {
		attempts++
		if attempts < 3 {
			return "", &testError{"transient failure"}
		}
		return "success", nil
	}, &RetryOptions{
		MaxRetries: 3,
		BackoffMs:  10, // Small backoff for testing
	})

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	if result != "success" {
		t.Errorf("expected 'success', got '%s'", result)
	}

	if attempts != 3 {
		t.Errorf("expected 3 attempts, got %d", attempts)
	}
}

func TestExecuteWithRetryExhausted(t *testing.T) {
	attempts := 0
	start := time.Now()

	_, err := ExecuteWithRetry(func() (string, error) {
		attempts++
		return "", &testError{"always fails"}
	}, &RetryOptions{
		MaxRetries: 3,
		BackoffMs:  10,
	})

	elapsed := time.Since(start)

	if err == nil {
		t.Error("expected error after retries exhausted")
	}

	if attempts != 3 {
		t.Errorf("expected 3 attempts, got %d", attempts)
	}

	// Should have waited ~30ms total (10 + 20 for first two attempts)
	if elapsed < 20*time.Millisecond {
		t.Errorf("expected backoff delays, elapsed: %v", elapsed)
	}
}

func TestExecuteWithRetryShouldRetryFalse(t *testing.T) {
	attempts := 0
	_, err := ExecuteWithRetry(func() (string, error) {
		attempts++
		return "", &testError{"non-retryable"}
	}, &RetryOptions{
		MaxRetries:  3,
		BackoffMs:   10,
		ShouldRetry: func(err error) bool { return false },
	})

	if err == nil {
		t.Error("expected error")
	}

	if attempts != 1 {
		t.Errorf("expected 1 attempt when ShouldRetry returns false, got %d", attempts)
	}
}

// testError is a simple error type for testing
type testError struct {
	msg string
}

func (e *testError) Error() string {
	return e.msg
}
