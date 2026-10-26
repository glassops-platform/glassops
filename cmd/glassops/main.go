// Command glassops is the GlassOps CLI.
package main

import (
	"os"

	"github.com/glassops-platform/glassops/internal/cli"
)

// version is set at build time with -ldflags "-X main.version=...".
var version = "dev"

func main() {
	os.Exit(cli.Execute(cli.NewApp(version), os.Args[1:]))
}
