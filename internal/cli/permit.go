package cli

import (
	"errors"
	"fmt"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/glassops-platform/glassops/internal/core/config"
	"github.com/glassops-platform/glassops/internal/core/permit"
	"github.com/glassops-platform/glassops/internal/core/run"
	"github.com/glassops-platform/glassops/internal/providers"
)

type permitVerifyOptions struct {
	target     string
	permitPath string
	planPath   string
}

func newPermitCommand(app *App, g *globals) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "permit",
		Short: "Work with PERMITs",
		Args:  cobra.NoArgs,
	}
	cmd.AddCommand(newPermitVerifyCommand(app, g))
	return cmd
}

func newPermitVerifyCommand(app *App, g *globals) *cobra.Command {
	o := &permitVerifyOptions{}
	cmd := &cobra.Command{
		Use:   "verify",
		Short: "Check that a PERMIT allows this deploy",
		Long: "permit verify checks the PERMIT's signature and that it allows a deploy of this\n" +
			"plan, at this commit, with this config, to this target, before it expires. It\n" +
			"exits 0 when the PERMIT holds and 1 when it does not, so any deploy tool can honour\n" +
			"a PERMIT. It exits 2 when GLASSOPS_PERMIT_KEY is not set or usable, or when the\n" +
			"PERMIT, the plan or the config cannot be read. It needs the key and never logs in\n" +
			"to the target.",
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return app.permitVerify(g, o)
		},
	}
	f := cmd.Flags()
	f.StringVar(&o.target, "target", "", "target name from glassops.yaml (required)")
	f.StringVar(&o.permitPath, "permit", "", "PERMIT file (default <dir>/.glassops/permit.json)")
	f.StringVar(&o.planPath, "plan", "", "change set file (default <dir>/.glassops/plan.json)")
	_ = cmd.MarkFlagRequired("target")
	return cmd
}

// permitVerify checks the PERMIT against what this run would deploy. A PERMIT that is not
// valid is a fail (exit 1). Anything that stops the check from running, such as a missing
// key or file, is an error (exit 2). Missing evidence never passes.
func (app *App) permitVerify(g *globals, o *permitVerifyOptions) error {
	c, err := app.detectCI(g)
	if err != nil {
		return err
	}
	res := run.New("permit-verify", app.Now)
	res.Target = o.target
	stop := func(check string, err error) error {
		res.AddCheck(check, run.StatusError, err.Error())
		return app.finish(res, c, g)
	}

	// The key comes first, so a missing key is reported before anything else is read.
	keyText, ok := app.LookupEnv(permit.KeyEnv)
	if !ok || keyText == "" {
		return stop("key", fmt.Errorf("%s is not set", permit.KeyEnv))
	}
	key, err := permit.ParseKey(keyText)
	if err != nil {
		return stop("key", err)
	}
	signer := permit.NewHMAC(key)
	res.Facts["permit.key_id"] = signer.(permit.KeyIdentifier).KeyID()

	proj, err := app.loadProject(g)
	if err != nil {
		return stop("config", err)
	}
	res.Provider = proj.provider.Name()
	loaded, err := config.Load(g.dir, g.configPath, proj.cfg.Provider)
	if err != nil {
		return stop("config", err)
	}
	t, err := app.target(proj.cfg, o.target)
	if err != nil {
		return stop("target", err)
	}

	planPath := o.planPath
	if planPath == "" {
		planPath = filepath.Join(g.dir, ".glassops", "plan.json")
	}
	cs, err := providers.ReadChangeSet(planPath)
	if err != nil {
		return stop("plan", err)
	}
	commit := c.Context().SHA
	if commit == "" {
		return stop("commit", errors.New("the CI gave no commit to check the PERMIT against"))
	}

	permitPath := o.permitPath
	if permitPath == "" {
		permitPath = filepath.Join(g.dir, ".glassops", "permit.json")
	}
	p, err := permit.Read(permitPath)
	switch {
	case errors.Is(err, permit.ErrUnreadable):
		res.AddCheck("permit", run.StatusFail, err.Error())
		return app.finish(res, c, g)
	case err != nil:
		return stop("permit", err)
	}
	res.Facts["permit.id"] = p.ID
	res.Facts["permit.expires_at"] = p.ExpiresAt.UTC()

	canon := proj.provider.CanonicalID
	err = permit.Verify(p, signer, permit.Expect{
		TargetName:   o.target,
		TargetID:     canon(t.ID),
		Commit:       commit,
		PlanSHA256:   cs.SHA256,
		ConfigSHA256: loaded.SHA256(),
		Now:          app.Now(),
		CanonicalID:  canon,
	})
	switch {
	case err == nil:
		res.AddCheck("permit", run.StatusPass, fmt.Sprintf("permit %s allows this deploy to %s until %s", p.ID, o.target, p.ExpiresAt.UTC().Format("2006-01-02T15:04:05Z")))
	case isPermitRefusal(err):
		res.AddCheck("permit", run.StatusFail, err.Error())
	default:
		return stop("permit", err)
	}
	return app.finish(res, c, g)
}

func isPermitRefusal(err error) bool {
	return errors.Is(err, permit.ErrSignature) || errors.Is(err, permit.ErrDenied) ||
		errors.Is(err, permit.ErrExpired) || errors.Is(err, permit.ErrMismatch)
}

// permitKeyWarning reports whether the permit key is present where it should not be: in a
// pull request run. It checks presence only and never reads the value into any output.
func (app *App) permitKeyWarning(event string) (string, bool) {
	if event != "pull_request" && event != "pull_request_target" {
		return "", false
	}
	if v, ok := app.LookupEnv(permit.KeyEnv); !ok || v == "" {
		return "", false
	}
	return permit.KeyEnv + " is set in a " + event + " run. The key should exist only in the protected deploy job, because a pull request run can sign a PERMIT with it", true
}
