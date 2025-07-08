package mocks_test

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"regexp"
	"strings"
	"testing"
	"time"

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
	t         *testing.T
	errorInfo string
	lastTest  models.TestInterface
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
	c.lastTest = t
	c.errorInfo = fmt.Sprintf("test %q (%s) failed", t.GetName(), t.GetFileName())

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

	assert.Equal(c.t, c.getExpected(), normalizeString(content), c.errorInfo)
	return false, nil
}

func (c *errorChecker) CheckRequest(mockName string, req *http.Request, resp *http.Response) []error {
	assert.Equal(c.t, mockName, "someservice", c.errorInfo)
	if strings.Contains(c.lastTest.GetFileName(), "drop_request") {
		// we don't check body for drop_request strategy
		return nil
	}
	assert.NotNil(c.t, resp, c.errorInfo)

	bodyBytes, err := io.ReadAll(resp.Body)
	require.NoError(c.t, err, c.errorInfo)

	if !strings.Contains(c.getExpected(), "unhandled request to mock") &&
		!strings.Contains(c.lastTest.GetFileName(), "nop") {
		assert.Contains(c.t, string(bodyBytes), "result", c.errorInfo)
	}
	return nil
}

func (c *errorChecker) getExpected() string {
	expected := c.lastTest.GetMeta("expected")
	if expected == nil {
		return ""
	}
	return normalizeString(expected.(string))
}

func Test_Declarative(t *testing.T) {
	m := mocks.NewNop("someservice")
	err := m.Start()
	require.NoError(t, err)
	defer m.Shutdown()

	checker := &errorChecker{
		t: t,
	}
	m.SetCheckers([]mocks.CheckerInterface{checker, checker})

	opts := &runner.RunnerOpts{
		Host:        "http://" + m.Service("someservice").ServerAddr(),
		Mocks:       m,
		MocksLoader: mocks.NewYamlLoader(&mocks.YamlLoaderOpts{}),
		TestHandler: checker.Handle,
	}

	r := runner.New(yaml_file.NewLoader("testdata"), opts)
	err = r.Run()
	require.NoError(t, err)
}

func Test_MocksWithPort(t *testing.T) {
	m := mocks.NewNop("someservice")
	err := m.Start()
	require.NoError(t, err)
	defer m.Shutdown()

	time.Sleep(100 * time.Millisecond)

	require.NotNil(t, m.Service("someservice"))
	addr := m.Service("someservice").ServerAddr()

	_, port, err := net.SplitHostPort(addr)
	require.NoError(t, err)

	m = mocks.NewNop("someservice:" + port)
	err = m.Start()
	require.ErrorContains(t, err, fmt.Sprintf("listen tcp 127.0.0.1:%s: bind:", port))

	require.NotNil(t, m.Service("someservice"))
	require.Panics(t, func() {
		m.Service("someservice").ServerAddr()
	})
}
