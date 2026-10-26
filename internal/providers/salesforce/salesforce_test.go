package salesforce

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/glassops-platform/glassops/internal/providers"
)

func TestDetect(t *testing.T) {
	dir := t.TempDir()
	p := New()
	if ok, err := p.Detect(dir); err != nil || ok {
		t.Fatalf("empty dir: ok=%v err=%v", ok, err)
	}
	if err := os.WriteFile(filepath.Join(dir, ProjectFile), []byte("{}"), 0o644); err != nil {
		t.Fatal(err)
	}
	if ok, err := p.Detect(dir); err != nil || !ok {
		t.Fatalf("sfdx project: ok=%v err=%v", ok, err)
	}
	reg := providers.Registry{p}
	if got, _ := reg.Detect(dir); got == nil || got.Name() != "salesforce" {
		t.Fatalf("registry detect = %v", got)
	}
	if reg.Get("salesforce") == nil || reg.Get("terraform") != nil {
		t.Fatal("registry get")
	}
}

func fakeRunner(outputs map[string]string, fail map[string]bool) Runner {
	return func(name string, args ...string) ([]byte, error) {
		key := strings.Join(append([]string{name}, args...), " ")
		if fail[key] {
			return nil, errors.New("exit status 1")
		}
		return []byte(outputs[key]), nil
	}
}

func TestToolVersions(t *testing.T) {
	// sf 2.154 prints a top-level array. The wrapped shape is accepted as well.
	shapes := map[string]string{
		"array":   `[{"name":"@salesforce/cli","version":"2.155.7"},{"name":"sfdx-git-delta","version":"6.1.0"}]`,
		"wrapped": `{"status":0,"result":[{"name":"sfdx-git-delta","version":"6.1.0"}]}`,
	}
	for name, plugins := range shapes {
		t.Run(name, func(t *testing.T) {
			p := NewWithRunner(fakeRunner(map[string]string{
				"sf --version":      "@salesforce/cli/2.155.7 linux-x64 node-v22.23.4\n",
				"sf plugins --json": plugins,
			}, nil))
			tools := p.Tools()
			if v, err := tools[0].Version(); err != nil || v != "2.155.7" {
				t.Fatalf("sf version = %q, %v", v, err)
			}
			if v, err := tools[1].Version(); err != nil || v != "6.1.0" {
				t.Fatalf("sgd version = %q, %v", v, err)
			}
		})
	}
}

func TestToolsMissing(t *testing.T) {
	p := NewWithRunner(fakeRunner(map[string]string{
		"sf plugins --json": `[{"name":"@salesforce/cli","version":"2.155.7"}]`,
	}, map[string]bool{"sf --version": true}))
	tools := p.Tools()
	if _, err := tools[0].Version(); err == nil {
		t.Fatal("expected sf error")
	}
	if _, err := tools[1].Version(); err == nil || !strings.Contains(err.Error(), "not installed") {
		t.Fatalf("sgd err = %v", err)
	}
}
