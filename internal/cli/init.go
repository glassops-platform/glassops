package cli

import (
	"bytes"
	_ "embed"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"text/template"

	"github.com/spf13/cobra"

	"github.com/glassops-platform/glassops/internal/core/config"
	"github.com/glassops-platform/glassops/internal/core/exitcode"
	"github.com/glassops-platform/glassops/internal/core/run"
)

//go:embed templates/glassops.yaml.tmpl
var configTemplate string

//go:embed templates/github-workflow.yml.tmpl
var githubWorkflowTemplate string

// githubWorkflowPath is where init writes the GitHub workflow.
const githubWorkflowPath = ".github/workflows/glassops.yml"

func newInitCommand(app *App, g *globals) *cobra.Command {
	var force bool
	cmd := &cobra.Command{
		Use:   "init",
		Short: "Detect the project and CI, then write glassops.yaml and a CI file",
		Long: "init detects the provider (Salesforce: sfdx-project.json) and the CI system,\n" +
			"then writes glassops.yaml with stub targets and the default pr and deploy\n" +
			"pipelines, plus a CI file for the detected CI. Existing files are never\n" +
			"overwritten unless --force is given.",
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return app.init(g, force)
		},
	}
	cmd.Flags().BoolVar(&force, "force", false, "overwrite existing files")
	return cmd
}

type initData struct {
	Provider string
}

func (app *App) init(g *globals, force bool) error {
	c, err := app.detectCI(g)
	if err != nil {
		return &exitcode.ExitError{Code: exitcode.Error, Err: err}
	}
	res := run.New("init", app.Now)

	p, err := app.Providers.Detect(g.dir)
	if err != nil {
		return &exitcode.ExitError{Code: exitcode.Error, Err: err}
	}
	data := initData{}
	if p != nil {
		data.Provider = p.Name()
		res.Provider = p.Name()
		res.AddCheck("provider", run.StatusPass, p.Name())
	} else {
		res.AddCheck("provider", run.StatusSkipped, "no provider detected, writing a provider-neutral config")
	}

	cfgPath := g.configPath
	if cfgPath == "" {
		cfgPath = filepath.Join(g.dir, config.FileName)
	}
	cfgBytes, err := render(configTemplate, data)
	if err != nil {
		return &exitcode.ExitError{Code: exitcode.Error, Err: err}
	}
	// What init writes must load cleanly, or the first doctor run would fail.
	if _, err := config.Parse(cfgBytes); err != nil {
		return &exitcode.ExitError{Code: exitcode.Error, Err: fmt.Errorf("generated config is invalid: %w", err)}
	}
	app.writeFile(res, "config", cfgPath, cfgBytes, force)

	ciSystem := detectCIFiles(g.dir, c.Name())
	res.Facts["ci"] = ciSystem
	switch ciSystem {
	case "github":
		wf, err := render(githubWorkflowTemplate, data)
		if err != nil {
			return &exitcode.ExitError{Code: exitcode.Error, Err: err}
		}
		app.writeFile(res, "ci.github", filepath.Join(g.dir, filepath.FromSlash(githubWorkflowPath)), wf, force)
	case "":
		res.AddCheck("ci", run.StatusSkipped, "no CI detected, so no CI file written")
	default:
		res.AddCheck("ci", run.StatusSkipped, ciSystem+" is detected but not supported yet, so no CI file written")
	}

	return app.finish(res, c, g)
}

// detectCIFiles looks at the project files first, then at the CI the command runs in.
func detectCIFiles(dir, running string) string {
	markers := []struct{ path, name string }{
		{".github", "github"},
		{"bitbucket-pipelines.yml", "bitbucket"},
		{".gitlab-ci.yml", "gitlab"},
		{"azure-pipelines.yml", "azure"},
	}
	for _, m := range markers {
		if _, err := os.Stat(filepath.Join(dir, m.path)); err == nil {
			return m.name
		}
	}
	if running != "local" {
		return running
	}
	return ""
}

// writeFile writes data unless the file exists and force is off. Outcomes become checks:
// refusing to overwrite is a failure, so a script notices that nothing was written.
func (app *App) writeFile(res *run.Result, check, path string, data []byte, force bool) {
	_, err := os.Stat(path)
	exists := err == nil
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		res.AddCheck(check, run.StatusError, err.Error())
		return
	}
	if exists && !force {
		res.AddCheck(check, run.StatusFail, path+" already exists (use --force to overwrite)")
		return
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		res.AddCheck(check, run.StatusError, err.Error())
		return
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		res.AddCheck(check, run.StatusError, err.Error())
		return
	}
	res.Artifacts = append(res.Artifacts, path)
	verb := "wrote"
	if exists {
		verb = "overwrote"
	}
	res.AddCheck(check, run.StatusPass, verb+" "+path)
}

func render(tmpl string, data any) ([]byte, error) {
	t, err := template.New("t").Option("missingkey=error").Parse(tmpl)
	if err != nil {
		return nil, err
	}
	var b bytes.Buffer
	if err := t.Execute(&b, data); err != nil {
		return nil, err
	}
	return b.Bytes(), nil
}
