package salesforce

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/glassops-platform/glassops/internal/providers"
)

func TestSessionCloseIsIdempotentAndRemovesHome(t *testing.T) {
	s := newTestSession(t, providers.Target{Name: "x"})
	if err := os.WriteFile(filepath.Join(s.home, "jwt.key"), []byte("k"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := s.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(s.home); !os.IsNotExist(err) {
		t.Errorf("home exists: %v", err)
	}
	if err := s.Close(); err != nil {
		t.Errorf("second close: %v", err)
	}
}

func TestSessionStopRunsOnClose(t *testing.T) {
	s := newTestSession(t, providers.Target{Name: "x"})
	stopped := 0
	s.stop = func() bool { stopped++; return true }
	_ = s.Close()
	_ = s.Close()
	if stopped != 1 {
		t.Errorf("stop ran %d times", stopped)
	}
}

func TestSessionEnvAndUnwrap(t *testing.T) {
	s := newTestSession(t, providers.Target{Name: "x"})
	env := s.env()
	if len(env) != 2 || env[0] != "HOME="+s.home || env[1] != "USERPROFILE="+s.home {
		t.Errorf("env = %v", env)
	}
	got, ok := asSession(providers.Session(s))
	if !ok || got != s {
		t.Error("asSession did not unwrap")
	}
}
