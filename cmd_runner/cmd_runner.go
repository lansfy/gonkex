package cmd_runner

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/google/shlex"
	"github.com/lansfy/gonkex/models"
)

func CmdRun(scriptPath string, timeout time.Duration) (string, error) {
	// by default timeout should be 3s
	if timeout <= 0 {
		timeout = 3 * time.Second
	}

	scriptPath = strings.TrimRight(scriptPath, "\n")
	scriptPath = strings.TrimSpace(scriptPath)
	if runtime.GOOS == "windows" {
		scriptPath = strings.ReplaceAll(scriptPath, "\\", "\\\\")
	}
	if scriptPath == "" {
		return "", fmt.Errorf("empty command provided")
	}

	args, err := shlex.Split(scriptPath)
	if err != nil {
		return "", err
	}

	args[0] = filepath.FromSlash(args[0])

	cmd := exec.Command(args[0], args[1:]...)
	cmd.Env = os.Environ()

	beforeStart(cmd)

	stdout := &bytes.Buffer{}
	cmd.Stdout = stdout
	cmd.Stderr = stdout

	if err := cmd.Start(); err != nil {
		return stdout.String(), err
	}

	done := make(chan error, 1)
	go func() {
		done <- cmd.Wait()
	}()

	select {
	case <-time.After(timeout):
		if err := killProcess(cmd); err != nil {
			return stdout.String(), err
		}
		return stdout.String(), fmt.Errorf("process killed as timeout (%s) reached", timeout)
	case err := <-done:
		if err != nil {
			return stdout.String(), fmt.Errorf("process finished with error: %w", err)
		}
	}
	return stdout.String(), nil
}

func ExecuteScript(script models.Script) (string, error) {
	return CmdRun(script.CmdLine(), script.Timeout())
}
