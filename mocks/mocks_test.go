package mocks_test

import (
	"fmt"
	"regexp"
	"strings"
	"testing"

	"github.com/lansfy/gonkex/mocks"
	"github.com/lansfy/gonkex/models"
	"github.com/lansfy/gonkex/runner"
	"github.com/lansfy/gonkex/testloader/yaml_file"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// This regex matches :[port]/ after 127.0.0.1
var portRegexp = regexp.MustCompile(`127\.0\.0\.1:\d+`)

type errorChecker struct {
	t *testing.T
}

func normalizeString(s string) string {
	s = portRegexp.ReplaceAllString(s, "127.0.0.1:80")
	s = strings.ReplaceAll(s, "\r\n", "\n")
	s = strings.ReplaceAll(s, "The system cannot find the file specified.", "no such file or directory")
	return strings.TrimSpace(s)
}

func simplifyError(err error) string {
	input := err.Error()

	// Cut everything from ", request was"
	if idx := strings.Index(input, ", request was"); idx != -1 {
		input = input[:idx]
	}

	return normalizeString(input)
}

func (c *errorChecker) Handle(t models.TestInterface, f runner.TestExecutor) (bool, error) {
	content := ""
	result, err := f(t)
	if err != nil {
		content = simplifyError(err)
	}

	if content == "" {
		for idx, err := range result.Errors {
			content += fmt.Sprintf("%d) %s\n", idx+1, simplifyError(err))
		}
	}

	content = normalizeString(content)

	expectedStr := ""
	expected := t.GetMeta("expected")
	if expected != nil {
		expectedStr = normalizeString(expected.(string))
	}

	assert.Equal(c.t, expectedStr, content, "test %q (%s) failed", t.GetName(), t.GetFileName())
	return false, nil
}

func Test_Constraints(t *testing.T) {
	m := mocks.NewNop("someservice")
	err := m.Start()
	require.NoError(t, err)
	defer m.Shutdown()

	checker := &errorChecker{t}

	opts := &runner.RunnerOpts{
		Host:        "http://" + m.Service("someservice").ServerAddr(),
		Mocks:       m,
		MocksLoader: mocks.NewYamlLoader(&mocks.YamlLoaderOpts{}),
		TestHandler: checker.Handle,
	}

	r := runner.New(yaml_file.NewLoader("testdata/constraints"), opts)
	err = r.Run()
	require.NoError(t, err)
}

func Test_Strategy(t *testing.T) {
	m := mocks.NewNop("someservice")
	err := m.Start()
	require.NoError(t, err)
	defer m.Shutdown()

	checker := &errorChecker{t}

	opts := &runner.RunnerOpts{
		Host:        "http://" + m.Service("someservice").ServerAddr(),
		Mocks:       m,
		MocksLoader: mocks.NewYamlLoader(&mocks.YamlLoaderOpts{}),
		TestHandler: checker.Handle,
	}

	r := runner.New(yaml_file.NewLoader("testdata/strategy"), opts)
	err = r.Run()
	require.NoError(t, err)
}
