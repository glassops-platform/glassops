// Package salesforce is the Salesforce provider. It drives the sf CLI and sfdx-git-delta.
//
// Every org call runs in a session: a private temp dir that holds the key file and an
// isolated sf state, so a login never reaches the user's own sf store. See session.go.
package salesforce

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/glassops-platform/glassops/internal/providers"
)

// ProjectFile marks a Salesforce DX project.
const ProjectFile = "sfdx-project.json"

// Runner runs a command and returns its stdout. Tool checks use it. Tests replace it.
type Runner func(name string, args ...string) ([]byte, error)

// Provider is the Salesforce provider.
type Provider struct {
	run Runner
	// exec runs sf and git for sessions, plans and validates. See sf.go.
	exec Exec
	// lookupEnv resolves the $VAR references in a target's auth.
	lookupEnv func(string) (string, bool)
	// tempDir is where session dirs are created. Empty means the OS temp dir.
	tempDir string
}

// New returns the provider using the real command runner.
func New() *Provider {
	return &Provider{
		run: func(name string, args ...string) ([]byte, error) {
			return exec.Command(name, args...).Output()
		},
		exec:      osExec,
		lookupEnv: os.LookupEnv,
	}
}

// NewWithRunner returns the provider with a custom runner, for tests.
func NewWithRunner(r Runner) *Provider {
	p := New()
	p.run = r
	return p
}

// Options replaces the provider's process environment, for tests. Zero fields keep the
// real behaviour.
type Options struct {
	Run       Runner
	Exec      Exec
	LookupEnv func(string) (string, bool)
	TempDir   string
}

// NewWithOptions returns the provider with the given replacements, for tests.
func NewWithOptions(o Options) *Provider {
	p := New()
	if o.Run != nil {
		p.run = o.Run
	}
	if o.Exec != nil {
		p.exec = o.Exec
	}
	if o.LookupEnv != nil {
		p.lookupEnv = o.LookupEnv
	}
	p.tempDir = o.TempDir
	return p
}

func (p *Provider) Name() string { return "salesforce" }

// CanonicalID compares Salesforce Ids by their first 15 characters, case-sensitively. The
// 18-character form only adds a case checksum.
func (p *Provider) CanonicalID(id string) string {
	if len(id) > 15 {
		return id[:15]
	}
	return id
}

func (p *Provider) Detect(dir string) (bool, error) {
	_, err := os.Stat(filepath.Join(dir, ProjectFile))
	if err == nil {
		return true, nil
	}
	if errors.Is(err, os.ErrNotExist) {
		return false, nil
	}
	return false, err
}

func (p *Provider) Tools() []providers.Tool {
	return []providers.Tool{
		{Name: "sf", Version: p.sfVersion},
		{Name: "sfdx-git-delta", Version: p.sgdVersion},
	}
}

// sfVersion parses "@salesforce/cli/2.155.7 linux-x64 node-v22.23.4".
func (p *Provider) sfVersion() (string, error) {
	out, err := p.run("sf", "--version")
	if err != nil {
		return "", fmt.Errorf("sf not found or not working: %w", err)
	}
	fields := strings.Fields(string(out))
	if len(fields) == 0 {
		return "", errors.New("sf --version printed nothing")
	}
	if i := strings.LastIndex(fields[0], "/"); i >= 0 {
		return fields[0][i+1:], nil
	}
	return fields[0], nil
}

type plugin struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

// parsePlugins reads "sf plugins --json". Current sf versions print a top-level array.
// The wrapped {"result": [...]} shape is accepted too.
func parsePlugins(out []byte) ([]plugin, error) {
	var list []plugin
	if err := json.Unmarshal(out, &list); err == nil {
		return list, nil
	}
	var wrapped struct {
		Result []plugin `json:"result"`
	}
	if err := json.Unmarshal(out, &wrapped); err != nil {
		return nil, err
	}
	return wrapped.Result, nil
}

// sgdVersion finds the sfdx-git-delta plugin in "sf plugins --json".
func (p *Provider) sgdVersion() (string, error) {
	out, err := p.run("sf", "plugins", "--json")
	if err != nil {
		return "", fmt.Errorf("could not list sf plugins: %w", err)
	}
	plugins, err := parsePlugins(out)
	if err != nil {
		return "", fmt.Errorf("could not read sf plugins output: %w", err)
	}
	for _, pl := range plugins {
		if pl.Name == "sfdx-git-delta" {
			return pl.Version, nil
		}
	}
	return "", errors.New("sfdx-git-delta is not installed (sf plugins install sfdx-git-delta)")
}
