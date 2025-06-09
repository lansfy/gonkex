package sqldb

import (
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_TestFixturesGeneration(t *testing.T) {
	tests := []struct {
		name       string
		inputFiles []string
		outputFile string
	}{
		{
			name:       "one file fixture",
			inputFiles: []string{"sql.yaml"},
			outputFile: "result.sql.yaml",
		},
		{
			name:       "MUST resolve refs",
			inputFiles: []string{"sql_refs.yaml"},
			outputFile: "result.sql_refs.yaml",
		},
		{
			name:       "MUST support extend rows",
			inputFiles: []string{"sql_extend.yaml"},
			outputFile: "result.sql_extend.yaml",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := convertToTestFixtures("testdata", tt.inputFiles)
			require.NoError(t, err)

			expected, err := os.ReadFile("testdata/" + tt.outputFile)
			require.NoError(t, err)
			expectedStr := strings.ReplaceAll(string(expected), "\r\n", "\n")
			actualStr := strings.ReplaceAll(string(data), "\r\n", "\n")
			require.Equal(t, expectedStr, actualStr, "generated fixture file doesn't match")
		})
	}
}
