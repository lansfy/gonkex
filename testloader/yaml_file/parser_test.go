package yaml_file

import (
	"testing"
)

func TestParseTestsWithCases(t *testing.T) {
	tests, err := parseTestDefinitionFile("testdata/parser.yaml")
	if err != nil {
		t.Error(err)
	}
	if len(tests) != 2 {
		t.Errorf("wait len(tests) == 2, got len(tests) == %d", len(tests))
	}
}
