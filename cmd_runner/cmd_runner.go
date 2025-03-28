//go:build !windows
// +build !windows

package cmd_runner

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"syscall"
	"time"
)

func CmdRun(scriptPath string, timeout time.Duration) error {
	// by default timeout should be 3s
	if timeout <= 0 {
		timeout = 3 * time.Second
	}
	cmd := exec.Command(strings.TrimRight(scriptPath, "\n"))
	cmd.Env = os.Environ()

	// Set up a process group which will be killed later
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}

	if err := cmd.Start(); err != nil {
		return err
	}

	done := make(chan error, 1)
	go func() {
		done <- cmd.Wait()
	}()

	select {
	case <-time.After(timeout):

		// Get process group which we want to kill
		pgid, err := syscall.Getpgid(cmd.Process.Pid)
		if err != nil {
			return err
		}
		// Send kill to process group
		if err := syscall.Kill(-pgid, 15); err != nil {
			return err
		}
		_, _ = fmt.Printf("Process killed as timeout (%s) reached\n", timeout)
	case err := <-done:
		if err != nil {
			return fmt.Errorf("process finished with error = %v", err)
		}
		_, _ = fmt.Print("Process finished successfully")
	}

	// Print log
	scanner := bufio.NewScanner(stdout)
	for scanner.Scan() {
		m := scanner.Text()
		_, _ = fmt.Println(m)
	}

	return nil
}
