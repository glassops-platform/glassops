package cli

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/glassops-platform/glassops/internal/core/config"
	"github.com/glassops-platform/glassops/internal/core/run"
)

// MinJavaMajor is the oldest Java the Code Analyzer runs on.
const MinJavaMajor = 11

// analyzerPlugin is the sf plugin that provides "sf code-analyzer".
const analyzerPlugin = "@salesforce/plugin-code-analyzer"

func init() {
	doctorExtras = append(doctorExtras, doctorAnalyze)
}

// doctorAnalyze checks what analyze needs when analyze.enabled is true: Java 11 or newer and
// the Code Analyzer plugin. It is silent when analyze is not switched on in the config, so a
// project that never analyzes is not asked for a JRE.
func doctorAnalyze(app *App, cfg *config.Config) []run.Check {
	if cfg.Provider != "salesforce" || cfg.Analyze.Enabled == nil || !*cfg.Analyze.Enabled {
		return nil
	}
	return []run.Check{app.javaCheck(), app.analyzerPluginCheck()}
}

// javaVersion matches the first line of "java --version", such as "openjdk 17.0.9 2023-10-17".
var javaVersion = regexp.MustCompile(`^\S+\s+"?((\d+)(?:\.\d+)*)`)

// javaCheck reads "java --version", which prints to stdout from Java 9 on. Java 8 does not
// know that flag and prints its version to stderr, which the runner does not return, so a
// failed "--version" followed by a working "-version" is reported as Java 8 or older.
func (app *App) javaCheck() run.Check {
	out, err := app.Run(".", "java", "--version")
	if err != nil {
		if _, err8 := app.Run(".", "java", "-version"); err8 == nil {
			return run.Check{Name: "tool.java", Status: run.StatusFail,
				Detail: fmt.Sprintf("Java 8 or older found, the Code Analyzer needs Java %d or newer", MinJavaMajor)}
		}
		return run.Check{Name: "tool.java", Status: run.StatusFail,
			Detail: fmt.Sprintf("java not found, the Code Analyzer needs Java %d or newer", MinJavaMajor)}
	}
	first := strings.TrimSpace(strings.SplitN(string(out), "\n", 2)[0])
	m := javaVersion.FindStringSubmatch(first)
	if m == nil {
		return run.Check{Name: "tool.java", Status: run.StatusFail, Detail: "could not read the Java version from: " + first}
	}
	if major, _ := strconv.Atoi(m[2]); major < MinJavaMajor {
		return run.Check{Name: "tool.java", Status: run.StatusFail,
			Detail: fmt.Sprintf("Java %s found, the Code Analyzer needs Java %d or newer", m[1], MinJavaMajor)}
	}
	return run.Check{Name: "tool.java", Status: run.StatusPass, Detail: first}
}

type sfPlugin struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

// analyzerPluginCheck looks for the Code Analyzer plugin in "sf plugins --json".
func (app *App) analyzerPluginCheck() run.Check {
	name := "tool.code-analyzer"
	out, err := app.Run(".", "sf", "plugins", "--json")
	if err != nil {
		return run.Check{Name: name, Status: run.StatusFail, Detail: "could not list sf plugins: " + err.Error()}
	}
	var list []sfPlugin
	if err := json.Unmarshal(out, &list); err != nil {
		var wrapped struct {
			Result []sfPlugin `json:"result"`
		}
		if err := json.Unmarshal(out, &wrapped); err != nil {
			return run.Check{Name: name, Status: run.StatusFail, Detail: "could not read sf plugins output: " + err.Error()}
		}
		list = wrapped.Result
	}
	for _, p := range list {
		if p.Name == analyzerPlugin {
			return run.Check{Name: name, Status: run.StatusPass, Detail: p.Version}
		}
	}
	return run.Check{Name: name, Status: run.StatusFail,
		Detail: "the Code Analyzer plugin is not installed (sf plugins install " + analyzerPlugin + ")"}
}
