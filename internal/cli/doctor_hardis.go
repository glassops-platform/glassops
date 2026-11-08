package cli

import (
	"github.com/glassops-platform/glassops/internal/core/config"
	"github.com/glassops-platform/glassops/internal/core/run"
	"github.com/glassops-platform/glassops/internal/providers/salesforce"
)

func init() {
	doctorExtras = append(doctorExtras, doctorHardis)
}

// doctorHardis checks that sfdx-hardis is installed at salesforce.hardis.min_version or later
// when salesforce.engine is hardis. It never installs anything. It is silent for the native
// engine and for other providers.
func doctorHardis(app *App, cfg *config.Config) []run.Check {
	if cfg.Salesforce.EngineName() != config.EngineHardis {
		return nil
	}
	if cfg.Provider != "" && cfg.Provider != "salesforce" {
		return nil
	}
	name := "tool.sfdx-hardis"
	dir := projectDirOf(cfg)
	if dir == "" {
		dir = "."
	}
	out, err := app.Run(dir, "sf", "plugins", "--json")
	if err != nil {
		return []run.Check{{Name: name, Status: run.StatusFail,
			Detail: "could not list sf plugins, engine hardis needs sf with sfdx-hardis: " + salesforce.HardisInstallSteps}}
	}
	got, err := salesforce.CheckHardisPlugins(out, cfg.Salesforce.Hardis.MinVersion)
	if err != nil {
		return []run.Check{{Name: name, Status: run.StatusFail, Detail: err.Error()}}
	}
	if got.Warning != "" {
		return []run.Check{{Name: name, Status: run.StatusPass, Detail: "warning: " + got.Warning}}
	}
	return []run.Check{{Name: name, Status: run.StatusPass, Detail: got.Version}}
}
