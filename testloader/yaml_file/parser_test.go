package yaml_file

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseTestsWithCases(t *testing.T) {
	tests, err := parseTestDefinitionFile(DefaultFileRead, "testdata/parser.yaml")
	require.NoError(t, err)
	require.Equal(t, 2, len(tests))
}
