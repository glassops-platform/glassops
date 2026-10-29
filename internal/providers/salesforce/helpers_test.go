package salesforce

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/glassops-platform/glassops/internal/providers"
)

// fixture returns testdata/sf/<name>.
func fixture(t *testing.T, name string) []byte {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("testdata", "sf", name))
	if err != nil {
		t.Fatal(err)
	}
	return data
}

// fakeReply is what fakeExec answers for one command.
type fakeReply struct {
	Stdout   []byte
	ExitCode int
	Err      error
}

// fakeExec answers commands by prefix. The key is the program name and its arguments joined
// by spaces, and the longest matching prefix wins. Every call is recorded.
type fakeExec struct {
	mu      sync.Mutex
	replies map[string]fakeReply
	calls   []Cmd
}

func newFakeExec(replies map[string]fakeReply) *fakeExec {
	return &fakeExec{replies: replies}
}

func (f *fakeExec) run(_ context.Context, c Cmd) (Output, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.calls = append(f.calls, c)
	key := strings.Join(append([]string{c.Name}, c.Args...), " ")
	best, found := "", false
	for prefix := range f.replies {
		if strings.HasPrefix(key, prefix) && len(prefix) >= len(best) {
			best, found = prefix, true
		}
	}
	if !found {
		return Output{Stdout: []byte(`{"status":1,"name":"UnexpectedCommand","message":"fakeExec has no reply for: ` + key + `"}`), ExitCode: 1}, nil
	}
	r := f.replies[best]
	return Output{Stdout: r.Stdout, ExitCode: r.ExitCode}, r.Err
}

// commands returns the recorded calls as joined strings.
func (f *fakeExec) commands() []string {
	f.mu.Lock()
	defer f.mu.Unlock()
	out := make([]string, len(f.calls))
	for i, c := range f.calls {
		out[i] = strings.Join(append([]string{c.Name}, c.Args...), " ")
	}
	return out
}

// newTestProvider returns a provider whose sf calls go to f and whose env lookups read env.
func newTestProvider(t *testing.T, f *fakeExec, env map[string]string) *Provider {
	t.Helper()
	return NewWithOptions(Options{
		Exec: f.run,
		LookupEnv: func(k string) (string, bool) {
			v, ok := env[k]
			return v, ok
		},
		TempDir: t.TempDir(),
	})
}

// newTestSession returns a session for target t with a temp home, as Authenticate would.
func newTestSession(t *testing.T, target providers.Target) *session {
	t.Helper()
	home := t.TempDir()
	return &session{target: target, home: home, alias: sessionAlias}
}
