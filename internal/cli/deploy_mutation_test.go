package cli

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/glassops-platform/glassops/internal/ci"
	"github.com/glassops-platform/glassops/internal/core/exitcode"
	"github.com/glassops-platform/glassops/internal/providers"
	execprovider "github.com/glassops-platform/glassops/internal/providers/exec"
)

// deploy.go checkTree: the refusal names at most maxTreePaths paths and counts the rest with
// `len(paths)-maxTreePaths`. The mutant `+` gives a wrong count.
func TestMutDeployTreePathCount(t *testing.T) {
	for _, tc := range []struct {
		files int
		more  string
	}{
		{maxTreePaths, ""},
		{maxTreePaths + 1, " and 1 more"},
		{maxTreePaths + 3, " and 3 more"},
	} {
		t.Run(fmt.Sprint(tc.files), func(t *testing.T) {
			h := newDeployHarness(t, deployConfig, nil)
			var b strings.Builder
			for i := 0; i < tc.files; i++ {
				fmt.Fprintf(&b, " M force-app/f%02d.cls\n", i)
			}
			h.tools[deployStatusCmd] = b.String()
			err := h.app.checkTree(&globals{dir: h.dir}, deployCommitSHA)
			if err == nil {
				t.Fatal("a modified tree must be refused")
			}
			msg := err.Error()
			if tc.more == "" && strings.Contains(msg, " more") {
				t.Fatalf("no overflow expected: %s", msg)
			}
			if tc.more != "" && !strings.HasSuffix(msg, tc.more) {
				t.Fatalf("want suffix %q: %s", tc.more, msg)
			}
			if got := strings.Count(msg, "force-app/f"); got != maxTreePaths {
				t.Fatalf("names %d paths, want %d: %s", got, maxTreePaths, msg)
			}
		})
	}
}

// deploy.go checkTree: a porcelain line of exactly three characters carries no path and is
// skipped. `len(line) > 3` against `>= 3` would add an empty path to the refusal.
func TestMutDeployTreeSkipsPathlessLine(t *testing.T) {
	h := newDeployHarness(t, deployConfig, nil)
	h.tools[deployStatusCmd] = " M \n M force-app/a.cls\n"
	err := h.app.checkTree(&globals{dir: h.dir}, deployCommitSHA)
	if err == nil {
		t.Fatal("a modified tree must be refused")
	}
	if !strings.HasSuffix(err.Error(), ": force-app/a.cls") {
		t.Fatalf("only the real path is named: %s", err)
	}
}

// deploy.go deployCommit: with no commit from the CI it reads git. The commit comes from git when
// git works, and a failed or empty git answer is refused.
func TestMutDeployCommitFromGit(t *testing.T) {
	h := newDeployHarness(t, deployConfig, nil)
	noSHA, err := ci.Detect("github", func(string) string { return "" }, io.Discard, h.dir)
	if err != nil {
		t.Fatal(err)
	}
	withSHA, err := ci.Detect("github", func(k string) string {
		if k == "GITHUB_SHA" {
			return "  ci-sha  "
		}
		return ""
	}, io.Discard, h.dir)
	if err != nil {
		t.Fatal(err)
	}
	g := &globals{dir: h.dir}
	for _, tc := range []struct {
		name string
		c    ci.CI
		out  string
		err  error
		want string
		ok   bool
	}{
		{"the CI reports a commit", withSHA, "git-sha\n", nil, "ci-sha", true},
		{"git answers", noSHA, "git-sha\n", nil, "git-sha", true},
		{"git fails with output", noSHA, "git-sha\n", errors.New("boom"), "", false},
		{"git answers nothing", noSHA, "\n", nil, "", false},
		{"git fails with nothing", noSHA, "", errors.New("boom"), "", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			h.app.Run = func(string, string, ...string) ([]byte, error) { return []byte(tc.out), tc.err }
			got, err := h.app.deployCommit(g, tc.c)
			if (err == nil) != tc.ok || got != tc.want {
				t.Fatalf("got %q, %v, want %q ok=%v", got, err, tc.want, tc.ok)
			}
		})
	}
}

// mutBinder records the workspace deploy binds to a provider.
type mutBinder struct {
	*fakeProvider
	ws execprovider.Workspace
}

func (b *mutBinder) Bind(w execprovider.Workspace) { b.ws = w }

// deploy.go: the PERMIT path handed to the provider is --permit when given and
// <dir>/.glassops/permit.json otherwise.
func TestMutDeployBindsPermitPath(t *testing.T) {
	h := newDeployHarness(t, deployConfig, nil)
	b := &mutBinder{fakeProvider: h.fp}
	h.app.Providers = providers.Registry{b}
	if code := h.run(); code != exitcode.OK {
		t.Fatalf("deploy exit %d: %s", code, h.out)
	}
	if want := filepath.Join(h.dir, ".glassops", "permit.json"); b.ws.PermitPath != want {
		t.Fatalf("default PermitPath %q, want %q", b.ws.PermitPath, want)
	}
	data, err := os.ReadFile(filepath.Join(h.dir, ".glassops", "permit.json"))
	if err != nil {
		t.Fatal(err)
	}
	other := filepath.Join(h.dir, "elsewhere.json")
	if err := os.WriteFile(other, data, 0o600); err != nil {
		t.Fatal(err)
	}
	if code := h.run("--permit", other); code != exitcode.OK {
		t.Fatalf("deploy --permit exit %d: %s", code, h.out)
	}
	if b.ws.PermitPath != other {
		t.Fatalf("PermitPath %q, want %q", b.ws.PermitPath, other)
	}
}
