//go:build windows

package cli

import "os/exec"

// setProcessGroup does nothing on Windows. Only the command itself is stopped there.
func setProcessGroup(*exec.Cmd) {}

// killProcessGroup kills the command itself if it is still running.
func killProcessGroup(cmd *exec.Cmd) error {
	if cmd.Process == nil || cmd.ProcessState != nil {
		return nil
	}
	return cmd.Process.Kill()
}
