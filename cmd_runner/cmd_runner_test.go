package cmd_runner

import (
	"os"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestCmdRun_Success(t *testing.T) {
	// Create a simple command that should succeed
	cmd := selectCommandLine("cmd /c echo hello", "echo hello")
	output, err := CmdRun(cmd, 0)
	require.NoError(t, err)
	require.Equal(t, "hello\n", normalize(output))
}

func TestCmdRun_DefaultTimeout(t *testing.T) {
	// Test with timeout <= 0 to verify default timeout is used
	cmd := selectCommandLine("cmd /c echo hello", "echo hello")
	output, err := CmdRun(cmd, 0)
	require.NoError(t, err)
	require.Equal(t, "hello\n", normalize(output))

	output, err = CmdRun(cmd, -1)
	require.NoError(t, err)
	require.Equal(t, "hello\n", normalize(output))
}

func TestCmdRun_Timeout(t *testing.T) {
	// Create a command that will run longer than the timeout
	cmd := createTestScript(t, "long_running")
	start := time.Now()
	_, err := CmdRun(cmd, 1*time.Second) // 1 second timeout
	duration := time.Since(start)

	require.Error(t, err)
	require.ErrorContains(t, err, "process killed as timeout (1s) reached")
	require.LessOrEqual(t, duration, 2*time.Second, "command took too long")
}

func TestCmdRun_InvalidCommand(t *testing.T) {
	// Test with a command that doesn't exist
	output, err := CmdRun("nonexistent_command_12345", 5*time.Second)
	require.Error(t, err)
	require.Equal(t, "", output)
}

func TestCmdRun_CommandWithNewline(t *testing.T) {
	// Test that trailing newlines are trimmed
	cmd := selectCommandLine("cmd /c echo hello\n", "echo hello\n")
	output, err := CmdRun(cmd, 5*time.Second)
	require.NoError(t, err)
	require.Equal(t, "hello\n", normalize(output))
}

func TestCmdRun_EmptyCommand(t *testing.T) {
	// Test with empty command
	output, err := CmdRun("", 5*time.Second)
	require.Error(t, err)
	require.ErrorContains(t, err, "empty command provided")
	require.Equal(t, "", output)
}

func TestCmdRun_CommandWithOutput(t *testing.T) {
	// Test a command that produces output
	cmd := selectCommandLine("cmd /c echo line1&& echo line2", "printf 'line1\\nline2\\n'")
	output, err := CmdRun(cmd, 5*time.Second)
	require.NoError(t, err)
	require.Equal(t, "line1\nline2\n", normalize(output))
}

func TestCmdRun_CommandWithError(t *testing.T) {
	// Test a command that exits with non-zero status
	cmd := createTestScript(t, "fails")
	output, err := CmdRun(cmd, 5*time.Second)
	require.Error(t, err)
	require.ErrorContains(t, err, "process finished with error: exit status")
	require.Equal(t, "Script executed\n", normalize(output))
}

func TestCmdRun_ScriptFile(t *testing.T) {
	// Test running a script file
	cmd := createTestScript(t, "simple")
	output, err := CmdRun(cmd, 5*time.Second)
	require.NoError(t, err)
	require.Equal(t, "Script executed\n", normalize(output))
}

func TestCmdRun_LongRunningScript(t *testing.T) {
	cmd := createTestScript(t, "long_running")

	start := time.Now()
	output, err := CmdRun(cmd, 5*time.Second) // Should complete within timeout
	duration := time.Since(start)
	require.NoError(t, err)

	// Should take at least 2 seconds but not much more
	require.LessOrEqual(t, 1*time.Second, duration, "script completed too quickly")
	require.Equal(t, "Starting\nFinished\n", normalize(output))
}

// Test concurrent execution
func TestCmdRun_Concurrent(t *testing.T) {
	cmd := selectCommandLine("cmd /c echo concurrent", "echo concurrent")

	const numGoroutines = 10
	errors := make(chan error, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func() {
			_, err := CmdRun(cmd, 5*time.Second)
			errors <- err
		}()
	}

	for i := 0; i < numGoroutines; i++ {
		if err := <-errors; err != nil {
			t.Fatalf("Concurrent execution failed: %v", err)
		}
	}
}

func normalize(s string) string {
	return strings.ReplaceAll(s, "\r\n", "\n")
}

func selectCommandLine(cmdWin, cmdLin string) string {
	if runtime.GOOS == "windows" {
		return cmdWin
	}
	return cmdLin
}

// Helper function to generate path to script with specified prefix
func createTestScript(t *testing.T, name string) string {
	t.Helper()

	if runtime.GOOS == "windows" {
		name += ".bat"
	} else {
		name += ".sh"
	}

	name = "testdata/" + name

	if runtime.GOOS != "windows" {
		err := os.Chmod(name, 0o755)
		require.NoError(t, err, "failed to make script executable")
		name = "bash " + name
	}

	return name
}
