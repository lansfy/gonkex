package allure

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/lansfy/gonkex/models"
	"github.com/lansfy/gonkex/testloader/yaml_file"

	"github.com/stretchr/testify/require"
)

var testResult map[string]string

func resetValues() {
	testResult = map[string]string{}
	writeFile = func(name string, data []byte, perm os.FileMode) error {
		testResult[name] = string(data)
		return nil
	}
	mkDir = func(name string, perm os.FileMode) error {
		return nil
	}
	t := time.Unix(0, 1000000)
	timeNow = func() time.Time {
		t = t.Add(100000 * time.Nanosecond)
		return t
	}
	counter := 1
	createUUID = func() string {
		counter++
		return fmt.Sprintf("[uuid%d]", counter)
	}
}

func checkTestResult(t *testing.T, fileCount int) {
	require.Len(t, testResult, fileCount)
	for name, value := range testResult {
		expected, err := os.ReadFile(name)
		require.NoError(t, err)
		expectedStr := strings.ReplaceAll(string(expected), "\r\n", "\n")
		value = strings.ReplaceAll(value, "\r\n", "\n")
		require.Equal(t, expectedStr, value, "content of file %s doesn't match", name)
	}
}

func TestParse_TestWithCases(t *testing.T) {
	resetValues()

	loader := yaml_file.NewLoader("testdata/testset1.yaml")
	tests, err := loader.Load()
	require.NoError(t, err)
	require.Len(t, tests, 2)

	output, err := NewOutput("testset1", "testdata/testset1")
	require.NoError(t, err)

	result := &models.Result{
		Path:                "/test/1/path",
		Query:               "?test1=query",
		RequestBody:         "body1\nbody1",
		ResponseStatusCode:  200,
		ResponseStatus:      "OK",
		ResponseContentType: "text/text",
		ResponseHeaders: map[string][]string{
			"aaa": {"bbb"},
		},
		ResponseBody: "somebody1",

		Errors: []error{
			errors.New("some error1"),
		},
		Test: tests[0],
	}

	err = output.Process(tests[0], result)
	require.NoError(t, err)

	err = output.Finalize()
	require.NoError(t, err)
	checkTestResult(t, 3)
}
