//go:build windows

package cmd_runner

import (
	"os/exec"
)

func beforeStart(cmd *exec.Cmd) {
}

func killProcess(cmd *exec.Cmd) error {
	return cmd.Process.Kill()
}
