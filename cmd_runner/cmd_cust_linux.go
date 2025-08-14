//go:build linux || darwin

package cmd_runner

import (
	"os/exec"
	"syscall"
)

func beforeStart(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
}

func killProcess(cmd *exec.Cmd) error {
	// Unix-like systems: Kill the process group
	pgid, err := syscall.Getpgid(cmd.Process.Pid)
	if err != nil {
		return err
	}
	return syscall.Kill(-pgid, 15)
}
