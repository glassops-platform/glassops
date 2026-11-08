package cli

import (
	"strings"

	"github.com/glassops-platform/glassops/internal/core/config"
	"github.com/glassops-platform/glassops/internal/core/run"
	"github.com/glassops-platform/glassops/internal/providers/salesforce"
)

func init() {
	doctorExtras = append(doctorExtras, doctorHardis)
}

// doctorHardis runs when salesforce.engine is hardis. tool.sfdx-hardis checks that sfdx-hardis
// is installed at salesforce.hardis.min_version or later, and never installs anything.
// salesforce.hardis.commands fails when the project's sfdx-hardis config defines deploy
// commands, which the engine refuses. It is silent for the native engine and other providers.
func doctorHardis(app *App, cfg *config.Config) []run.Check {
	if cfg.Salesforce.EngineName() != config.EngineHardis {
		return nil
	}
	if cfg.Provider != "" && cfg.Provider != "salesforce" {
		return nil
	}
	dir := projectDirOf(cfg)
	if dir == "" {
		dir = "."
	}
	return []run.Check{hardisToolCheck(app, cfg, dir), hardisCommandsCheck(dir)}
}

func hardisToolCheck(app *App, cfg *config.Config, dir string) run.Check {
	name := "tool.sfdx-hardis"
	out, err := app.Run(dir, "sf", "plugins", "--json")
	if err != nil {
		return run.Check{Name: name, Status: run.StatusFail,
			Detail: "could not list sf plugins, engine hardis needs sf with sfdx-hardis: " + salesforce.HardisInstallSteps}
	}
	got, err := salesforce.CheckHardisPlugins(out, cfg.Salesforce.Hardis.MinVersion)
	if err != nil {
		return run.Check{Name: name, Status: run.StatusFail, Detail: err.Error()}
	}
	if got.Warning != "" {
		return run.Check{Name: name, Status: run.StatusPass, Detail: "warning: " + got.Warning}
	}
	return run.Check{Name: name, Status: run.StatusPass, Detail: got.Version}
}

func hardisCommandsCheck(dir string) run.Check {
	name := "salesforce.hardis.commands"
	problems, err := salesforce.HardisCommandProblems(dir)
	switch {
	case err != nil:
		return run.Check{Name: name, Status: run.StatusError, Detail: err.Error()}
	case len(problems) > 0:
		return run.Check{Name: name, Status: run.StatusFail, Detail: strings.Join(problems, ". ")}
	}
	return run.Check{Name: name, Status: run.StatusPass, Detail: "no deploy commands in the sfdx-hardis config"}
}
