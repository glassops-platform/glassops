//go:build !windows

package cli

import (
	"context"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"testing"
	"time"
)

// A command that leaves a child running: the child is killed when the command ends, so it
// cannot keep using the session after the step.
func TestSessionShellRunnerKillsWhatTheCommandLeft(t *testing.T) {
	dir := t.TempDir()
	pidFile := filepath.Join(dir, "pid")
	// The child's output goes to /dev/null so it does not hold the step's pipes open.
	cmd := "sleep 60 >/dev/null 2>&1 & echo $! > " + pidFile + "; exit 4"
	code, err := sessionShellRunner(context.Background(), dir, os.Environ(), cmd, io.Discard, io.Discard)
	if err != nil || code != 4 {
		t.Fatalf("code %d, err %v, want exit 4", code, err)
	}
	data, err := os.ReadFile(pidFile)
	if err != nil {
		t.Fatal(err)
	}
	pid, err := strconv.Atoi(strings.TrimSpace(string(data)))
	if err != nil {
		t.Fatal(err)
	}
	deadline := time.Now().Add(5 * time.Second)
	for {
		if err := syscall.Kill(pid, 0); err == syscall.ESRCH {
			return
		}
		if time.Now().After(deadline) {
			_ = syscall.Kill(pid, syscall.SIGKILL)
			t.Fatalf("the child %d still runs after the step ended", pid)
		}
		time.Sleep(50 * time.Millisecond)
	}
}
