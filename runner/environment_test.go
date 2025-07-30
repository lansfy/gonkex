package runner_test

import (
	"os"
	"testing"

	"github.com/lansfy/gonkex/runner"

	"github.com/stretchr/testify/require"
)

func mustWriteFile(t *testing.T, name, content string) string {
	t.Helper()
	tmp := t.TempDir()
	path := tmp + "/" + name
	require.NoError(t, os.WriteFile(path, []byte(content), 0o644))
	return path
}

func Test_RegisterEnvironmentVariables(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		overload bool
		existing map[string]string
		expect   map[string]string
	}{
		{
			name:    "simple load",
			content: "FOO=bar\nBAZ=123",
			expect:  map[string]string{"FOO": "bar", "BAZ": "123"},
		},
		{
			name:     "no overwrite existing",
			content:  "FOO=should_not_set",
			existing: map[string]string{"FOO": "original"},
			expect:   map[string]string{"FOO": "original"},
		},
		{
			name:     "with overwrite",
			content:  "FOO=newval",
			existing: map[string]string{"FOO": "oldval"},
			overload: true,
			expect:   map[string]string{"FOO": "newval"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			envFile := mustWriteFile(t, tt.name+".env", tt.content)

			for k, v := range tt.existing {
				require.NoError(t, os.Setenv(k, v))
			}

			err := runner.RegisterEnvironmentVariables(envFile, tt.overload)
			require.NoError(t, err)

			for k, want := range tt.expect {
				require.Equal(t, want, os.Getenv(k), "env var %q", k)
			}
		})
	}
}

func Test_RegisterEnvironmentVariables_FileNotFound(t *testing.T) {
	err := runner.RegisterEnvironmentVariables("nonexistent.env", false)
	require.Error(t, err)
	require.True(t, os.IsNotExist(err), "expected file not found error, got: %v", err)
}

func Test_RegisterEnvironmentVariables_InvalidFormat(t *testing.T) {
	envFile := mustWriteFile(t, "invalid.env", "NO_EQUALS_LINE")

	err := runner.RegisterEnvironmentVariables(envFile, false)
	require.Error(t, err)
	require.Contains(t, err.Error(), "Can't separate key from value") // from godotenv
}
