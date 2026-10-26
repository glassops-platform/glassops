// Package cli builds the glassops command tree. Everything the commands touch outside the
// process (output, environment, clock, external programs) comes through App, so tests can
// run commands end to end without a real CI or real tools.
package cli

import (
	"io"
	"os"
	"os/exec"
	"time"

	"github.com/glassops-platform/glassops/internal/providers"
	"github.com/glassops-platform/glassops/internal/providers/salesforce"
)

// Runner runs a program in dir and returns its stdout.
type Runner func(dir, name string, args ...string) ([]byte, error)

// App is the process environment the commands run in.
type App struct {
	Stdout    io.Writer
	Stderr    io.Writer
	Getenv    func(string) string
	LookupEnv func(string) (string, bool)
	Now       func() time.Time
	Run       Runner
	Providers providers.Registry
	Version   string
}

// NewApp returns an App wired to the real process.
func NewApp(version string) *App {
	return &App{
		Stdout:    os.Stdout,
		Stderr:    os.Stderr,
		Getenv:    os.Getenv,
		LookupEnv: os.LookupEnv,
		Now:       time.Now,
		Run: func(dir, name string, args ...string) ([]byte, error) {
			cmd := exec.Command(name, args...)
			cmd.Dir = dir
			return cmd.Output()
		},
		Providers: providers.Registry{salesforce.New()},
		Version:   version,
	}
}
