// Package salesforce is the Salesforce provider. Phase 1 covers detection and tool checks.
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

// Runner runs a command and returns its stdout. Tests replace it.
type Runner func(name string, args ...string) ([]byte, error)

// Provider is the Salesforce provider.
type Provider struct {
	run Runner
}

// New returns the provider using the real command runner.
func New() *Provider {
	return &Provider{run: func(name string, args ...string) ([]byte, error) {
		return exec.Command(name, args...).Output()
	}}
}

// NewWithRunner returns the provider with a custom runner, for tests.
func NewWithRunner(r Runner) *Provider { return &Provider{run: r} }

func (p *Provider) Name() string { return "salesforce" }

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
