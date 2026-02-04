// Package services contains runtime service implementations.
package services

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// RuntimeEnvironment handles CLI installation and plugin management.
type RuntimeEnvironment struct {
	platform string
}

// NewRuntimeEnvironment creates a new RuntimeEnvironment.
func NewRuntimeEnvironment() *RuntimeEnvironment {
	return &RuntimeEnvironment{
		platform: os.Getenv("GOOS"),
	}
}

// Install and InstallPlugins methods have been removed as part of the Runtime Decoupling migration.
// The execution environment is now expected to have all necessary dependencies pre-installed.

func (r *RuntimeEnvironment) execWithAutoConfirm(command string, args []string) error {
	quotedArgs := make([]string, len(args))
	for i, arg := range args {
		quotedArgs[i] = fmt.Sprintf(`"%s"`, arg)
	}
	fullCommand := fmt.Sprintf("%s %s", command, strings.Join(quotedArgs, " "))

	var cmd *exec.Cmd
	if r.platform == "windows" {
		cmd = exec.Command("cmd", "/c", fmt.Sprintf("echo y|%s", fullCommand))
	} else {
		cmd = exec.Command("sh", "-c", fmt.Sprintf("echo y | %s", fullCommand))
	}

	return cmd.Run()
}
