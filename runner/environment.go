package runner

import (
	"fmt"
	"os"
	"strings"

	"github.com/lansfy/gonkex/variables"

	"github.com/joho/godotenv"
)

// RegisterEnvironmentVariables loads environment variables from the specified env file.
// It reads the file named fileName and sets the variables in the current process environment.
// If overload is true, existing environment variables will be overridden.
func RegisterEnvironmentVariables(fileName string, overload bool) error {
	currentEnv := map[string]bool{}
	rawEnv := os.Environ()
	for _, rawEnvLine := range rawEnv {
		key := strings.Split(rawEnvLine, "=")[0]
		currentEnv[key] = true
	}

	data, err := os.ReadFile(fileName)
	if err != nil {
		return err
	}

	content := variables.New().Substitute(string(data))
	envMap, err := godotenv.Parse(strings.NewReader(content))
	if err != nil {
		return err
	}

	for key, value := range envMap {
		if !currentEnv[key] || overload {
			err = os.Setenv(key, value)
			if err != nil {
				return fmt.Errorf("register environment variable %q: %w", key, err)
			}
		}
	}
	return nil
}
