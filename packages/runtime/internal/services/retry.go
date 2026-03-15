package services

import (
	"time"
)

// RetryOptions configures retry behavior.
type RetryOptions struct {
	MaxRetries  int
	BackoffMs   int
	ShouldRetry func(error) bool
}

// DefaultRetryOptions returns sensible defaults.
func DefaultRetryOptions() RetryOptions {
	return RetryOptions{
		MaxRetries:  3,
		BackoffMs:   1000,
		ShouldRetry: func(error) bool { return true },
	}
}

// ExecuteWithRetry[T any] runs a function with exponential backoff retry logic.
func ExecuteWithRetry[T any](fn func() (T, error), opts *RetryOptions) (T, error) {
	if opts == nil {
		defaults := DefaultRetryOptions()
		opts = &defaults
	}

	var lastErr error
	var zero T

	for attempt := 0; attempt < opts.MaxRetries; attempt++ {
		result, err := fn()
		if err == nil {
			return result, nil
		}

		lastErr = err

		// Check if we should retry
		if opts.ShouldRetry != nil && !opts.ShouldRetry(err) {
			return zero, err
		}

		// Don't wait after last attempt
		if attempt < opts.MaxRetries-1 {
			delay := time.Duration(opts.BackoffMs*(1<<attempt)) * time.Millisecond
			time.Sleep(delay)
		}
	}

	return zero, lastErr
}
