package cli

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/glassops-platform/glassops/internal/core/exitcode"
	"github.com/glassops-platform/glassops/internal/core/run"
)

// gate.go: `if commandSHA != ""` writes the gate.deploy_command_sha256 fact only when a deploy
// command was bound. The mutant `==` flips that.
func TestMutGateDeployCommandFact(t *testing.T) {
	for _, tc := range []struct {
		name string
		args []string
		want bool
	}{
		{"no command", nil, false},
		{"a command", []string{"--deploy-command", `["./other.sh","two words"]`}, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			p := newExecProject(t, execConfig)
			p.plan(t)
			args := append([]string{"gate", "--target", "qa", "--no-connect"}, tc.args...)
			if code := p.exec(args...); code != exitcode.OK {
				t.Fatalf("gate exit %d: %s", code, p.out.String())
			}
			data, err := os.ReadFile(filepath.Join(p.dir, ".glassops", "results", "gate.json"))
			if err != nil {
				t.Fatal(err)
			}
			var r run.Result
			if err := json.Unmarshal(data, &r); err != nil {
				t.Fatal(err)
			}
			v, has := r.Facts["gate.deploy_command_sha256"]
			if has != tc.want || (has && v == "") {
				t.Fatalf("fact present=%v value=%v, want present=%v", has, v, tc.want)
			}
		})
	}
}
