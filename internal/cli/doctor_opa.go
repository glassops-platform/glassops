package cli

import (
	"context"
	"strings"

	"github.com/glassops-platform/glassops/internal/core/config"
	"github.com/glassops-platform/glassops/internal/core/policy"
	"github.com/glassops-platform/glassops/internal/core/run"
)

// opaInstall is where doctor points when opa is missing.
const opaInstall = "https://www.openpolicyagent.org/docs/latest/#running-opa"

func init() {
	doctorExtras = append(doctorExtras, doctorOPA)
	// Rego files in policy.rego are relative to the project dir, the same as every other path
	// in glassops.yaml.
	policy.RegoDir = projectDirOf
}

// projectDirOf returns the project dir recorded for cfg, or "" when none was.
func projectDirOf(cfg *config.Config) string {
	if cfg == nil {
		return ""
	}
	if v, ok := projectDirs.Load(cfg); ok {
		if dir, ok := v.(string); ok {
			return dir
		}
	}
	return ""
}

// doctorOPA checks what gate needs when policy.rego names files: opa on PATH, and Rego files
// that load and define data.glassops.deny as a set. It is silent when policy.rego is not set.
func doctorOPA(app *App, cfg *config.Config) []run.Check {
	if !cfg.Policy.Rego.Enabled() {
		return nil
	}
	dir := projectDirOf(cfg)
	if dir == "" {
		dir = "."
	}
	tool := app.opaCheck(dir)
	if tool.Status != run.StatusPass {
		return []run.Check{tool, {Name: "policy.rego", Status: run.StatusSkipped, Detail: "opa is not available"}}
	}
	runner := func(_ context.Context, dir string, args ...string) ([]byte, error) {
		return app.Run(dir, "opa", args...)
	}
	files := run.Check{Name: "policy.rego", Status: run.StatusPass, Detail: strings.Join(cfg.Policy.Rego.Files, ", ")}
	if err := policy.CheckRego(context.Background(), runner, dir, cfg.Policy.Rego.Files); err != nil {
		files.Status, files.Detail = run.StatusFail, err.Error()
	}
	return []run.Check{tool, files}
}

// opaCheck reads "opa version", whose first line is "Version: <semver>".
func (app *App) opaCheck(dir string) run.Check {
	name := "tool.opa"
	out, err := app.Run(dir, "opa", "version")
	if err != nil {
		return run.Check{Name: name, Status: run.StatusFail,
			Detail: "opa not found, policy.rego needs it (" + opaInstall + ")"}
	}
	for _, line := range strings.Split(string(out), "\n") {
		if v, ok := strings.CutPrefix(strings.TrimSpace(line), "Version:"); ok {
			return run.Check{Name: name, Status: run.StatusPass, Detail: strings.TrimSpace(v)}
		}
	}
	return run.Check{Name: name, Status: run.StatusPass, Detail: "version not reported"}
}
