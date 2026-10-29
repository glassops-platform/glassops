package salesforce

import (
	"context"
	"regexp"
	"strings"
	"time"
)

// retryPolicy controls how a login is retried. Tests replace loginRetry to skip the waits.
type retryPolicy struct {
	// Attempts is the total number of tries, including the first.
	Attempts int
	// Backoff is the wait before the second try. It doubles for each later one.
	Backoff time.Duration
	// Sleep waits for d or until ctx ends, and returns ctx's error in that case.
	Sleep func(ctx context.Context, d time.Duration) error
}

var loginRetry = retryPolicy{Attempts: 3, Backoff: time.Second, Sleep: sleepCtx}

func sleepCtx(ctx context.Context, d time.Duration) error {
	t := time.NewTimer(d)
	defer t.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-t.C:
		return nil
	}
}

// transientWords are the failures worth another try. This is an allowlist on purpose: a bad
// key, an unknown user, an expired token or an app that is not installed never heals, and
// sf already retries JWT logins itself. Anything not listed here is final.
var transientWords = []string{
	"enotfound",
	"etimedout",
	"econnreset",
	"econnrefused",
	"eai_again",
	"socket hang up",
	"service unavailable",
	"bad gateway",
	"gateway timeout",
	"unable_to_lock_row",
}

// http5xx matches an HTTP status in the 500s when it is written as a status, not as a
// stray number. For example "status code 503", "HTTP 502" or "HTTP/1.1 500".
var http5xx = regexp.MustCompile(`(?i)(http[/0-9. ]*|status( code)?:? ?)\b5\d\d\b`)

// isTransient reports whether msg describes a failure that can heal by waiting.
func isTransient(msg string) bool {
	low := strings.ToLower(msg)
	for _, w := range transientWords {
		if strings.Contains(low, w) {
			return true
		}
	}
	return http5xx.MatchString(msg)
}

// do runs try until it succeeds, fails with an error that is not transient, runs out of
// attempts or ctx ends. retryable decides whether an error is transient.
func (r retryPolicy) do(ctx context.Context, retryable func(error) bool, try func() error) error {
	var err error
	wait := r.Backoff
	for attempt := 1; ; attempt++ {
		if err = try(); err == nil {
			return nil
		}
		if ctx.Err() != nil {
			return ctx.Err()
		}
		if attempt >= r.Attempts || !retryable(err) {
			return err
		}
		if serr := r.Sleep(ctx, wait); serr != nil {
			return serr
		}
		wait *= 2
	}
}
