package salesforce

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"
)

func TestIsTransient(t *testing.T) {
	yes := []string{
		"getaddrinfo ENOTFOUND login.salesforce.com",
		"connect ETIMEDOUT 1.2.3.4:443",
		"read ECONNRESET",
		"connect ECONNREFUSED 127.0.0.1:443",
		"getaddrinfo EAI_AGAIN login.salesforce.com",
		"socket hang up",
		"HTTP 503",
		"HTTP/1.1 502 Bad Gateway",
		"request failed with status code 500",
		"Service Unavailable",
		"UNABLE_TO_LOCK_ROW: unable to obtain exclusive access to this record",
	}
	for _, m := range yes {
		if !isTransient(m) {
			t.Errorf("%q should be transient", m)
		}
	}
	no := []string{
		"invalid_grant: expired access/refresh token",
		"user hasn't approved this consumer",
		"invalid client credentials",
		"client identifier invalid",
		"External client app is not installed in this org",
		"Bad_Id",
		"the 500 records were skipped",
		"",
	}
	for _, m := range no {
		if isTransient(m) {
			t.Errorf("%q should not be transient", m)
		}
	}
}

// The real failures in the fixtures never heal, so each runs exactly once.
func TestRealLoginFailuresAreNeverRetried(t *testing.T) {
	cases := []struct {
		fixture string
		prefix  string
		target  func() string
	}{
		{"login-jwt-unknown-user.json", "sf org login jwt", func() string { return "jwt" }},
		{"login-jwt-wrong-key.json", "sf org login jwt", func() string { return "jwt" }},
		{"login-jwt-bad-client-id.json", "sf org login jwt", func() string { return "jwt" }},
		{"login-access-token-invalid.json", "sf org login access-token", func() string { return "token" }},
		{"login-sfdx-url-invalid.json", "sf org login sfdx-url", func() string { return "sfdx" }},
	}
	for _, tc := range cases {
		t.Run(tc.fixture, func(t *testing.T) {
			waits := noSleep(t)
			f := withReplies(map[string]fakeReply{tc.prefix: {Stdout: fixture(t, tc.fixture), ExitCode: 1}})
			p := newTestProvider(t, f, testEnv())
			target := jwtTarget()
			switch tc.target() {
			case "token":
				target = tokenTarget()
			case "sfdx":
				target = sfdxTarget()
			}
			if _, err := p.Authenticate(context.Background(), target); err == nil {
				t.Fatal("want an error")
			}
			if n := len(loginCalls(f)); n != 1 {
				t.Errorf("login ran %d times, want 1", n)
			}
			if len(*waits) != 0 {
				t.Errorf("slept %v before giving up", *waits)
			}
		})
	}
}

// sf login failures with a transient cause have no real fixture. These bodies are built in
// the test from the error words the allowlist names, in the envelope shape the real failures
// use.
func transientBody(t *testing.T, msg string) []byte {
	t.Helper()
	b, err := json.Marshal(map[string]any{"name": "GenericTimeoutError", "message": msg, "status": 1, "exitCode": 1})
	if err != nil {
		t.Fatal(err)
	}
	return b
}

func TestTransientFailuresRetryThreeTimes(t *testing.T) {
	for _, msg := range []string{"getaddrinfo ENOTFOUND login.salesforce.com", "connect ETIMEDOUT", "read ECONNRESET",
		"connect ECONNREFUSED", "getaddrinfo EAI_AGAIN", "socket hang up", "HTTP 503 Service Unavailable", "UNABLE_TO_LOCK_ROW"} {
		t.Run(msg, func(t *testing.T) {
			waits := noSleep(t)
			f := withReplies(map[string]fakeReply{"sf org login jwt": {Stdout: transientBody(t, msg), ExitCode: 1}})
			p := newTestProvider(t, f, testEnv())
			if _, err := p.Authenticate(context.Background(), jwtTarget()); err == nil {
				t.Fatal("want an error")
			}
			if n := len(loginCalls(f)); n != 3 {
				t.Errorf("login ran %d times, want 3", n)
			}
			if len(*waits) != 2 || (*waits)[1] != 2*(*waits)[0] {
				t.Errorf("backoff = %v, want two waits, the second double", *waits)
			}
			if got := entries(t, p.tempDir); len(got) != 0 {
				t.Errorf("home left behind: %v", got)
			}
		})
	}
}

func TestTransientThenSuccess(t *testing.T) {
	noSleep(t)
	n := 0
	exec := func(_ context.Context, c Cmd) (Output, error) {
		if c.Name == "icacls" {
			return Output{}, nil
		}
		n++
		if n < 3 {
			return Output{Stdout: transientBody(t, "socket hang up"), ExitCode: 1}, nil
		}
		return Output{Stdout: fixture(t, "login-sfdx-url-scratch.json")}, nil
	}
	p := NewWithOptions(Options{Exec: exec, TempDir: t.TempDir(),
		LookupEnv: func(k string) (string, bool) { v, ok := testEnv()[k]; return v, ok }})
	s, err := p.Authenticate(context.Background(), sfdxTarget())
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	if n != 3 {
		t.Errorf("calls = %d", n)
	}
}

func TestRetryHonoursContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	calls := 0
	r := retryPolicy{Attempts: 3, Backoff: time.Hour, Sleep: sleepCtx}
	go func() { time.Sleep(20 * time.Millisecond); cancel() }()
	err := r.do(ctx, func(error) bool { return true }, func() error { calls++; return errors.New("ENOTFOUND") })
	if !errors.Is(err, context.Canceled) {
		t.Errorf("err = %v", err)
	}
	if calls != 1 {
		t.Errorf("calls = %d", calls)
	}
}

func TestRetryStopsOnNonRetryable(t *testing.T) {
	calls := 0
	r := retryPolicy{Attempts: 3, Backoff: time.Millisecond, Sleep: func(context.Context, time.Duration) error { return nil }}
	err := r.do(context.Background(), func(error) bool { return false }, func() error { calls++; return errors.New("x") })
	if err == nil || calls != 1 {
		t.Errorf("err %v calls %d", err, calls)
	}
	if !strings.Contains(err.Error(), "x") {
		t.Error("error changed")
	}
}
