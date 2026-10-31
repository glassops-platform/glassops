package cli

import (
	"time"

	"github.com/spf13/cobra"
)

// defaultEnvTarget is the target name an ephemeral environment registers as by default.
const defaultEnvTarget = "ephemeral"

type envCreateOptions struct {
	target  string
	name    string
	ttlDays int
	shape   string
}

type envDeleteOptions struct {
	target string
	name   string
}

type envSweepOptions struct {
	target string
	maxAge time.Duration
}

func newEnvCommand(app *App, g *globals) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "env",
		Short: "Create, delete and sweep ephemeral environments",
		Long: "env manages the ephemeral environment configured under environments in\n" +
			"glassops.yaml. Each subcommand logs in to the environment's hub target.",
		Args: cobra.NoArgs,
	}
	cmd.AddCommand(newEnvCreateCommand(app, g), newEnvDeleteCommand(app, g), newEnvSweepCommand(app, g))
	return cmd
}

func newEnvCreateCommand(app *App, g *globals) *cobra.Command {
	o := &envCreateOptions{}
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create an ephemeral environment",
		Long: "env create creates the ephemeral environment for --target from its hub. The login\n" +
			"material is kept in a file only the current user can read, never in a result.",
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return notImplemented("env create")
		},
	}
	f := cmd.Flags()
	f.StringVar(&o.target, "target", defaultEnvTarget, "environment name from the environments section")
	f.StringVar(&o.name, "name", "", "environment name (default glassops-<repo>-<pr>-<run> with the configured prefix)")
	f.IntVar(&o.ttlDays, "ttl-days", 0, "days the environment lives (default ttl_days, else 1)")
	f.StringVar(&o.shape, "shape", "", "source org Id for an org shape (default the configured shape)")
	return cmd
}

func newEnvDeleteCommand(app *App, g *globals) *cobra.Command {
	o := &envDeleteOptions{}
	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete an ephemeral environment",
		Long:  "env delete deletes the ephemeral environment for --target from its hub.",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return notImplemented("env delete")
		},
	}
	f := cmd.Flags()
	f.StringVar(&o.target, "target", defaultEnvTarget, "environment name from the environments section")
	f.StringVar(&o.name, "name", "", "environment name (default the one env create made)")
	return cmd
}

func newEnvSweepCommand(app *App, g *globals) *cobra.Command {
	o := &envSweepOptions{}
	cmd := &cobra.Command{
		Use:   "sweep",
		Short: "Delete old ephemeral environments",
		Long: "env sweep deletes every environment on the hub whose name starts with the\n" +
			"configured prefix and that is older than --max-age, and lists what it deleted.",
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return notImplemented("env sweep")
		},
	}
	f := cmd.Flags()
	f.StringVar(&o.target, "target", defaultEnvTarget, "environment name from the environments section")
	f.DurationVar(&o.maxAge, "max-age", 0, "delete environments older than this, such as 24h")
	return cmd
}
