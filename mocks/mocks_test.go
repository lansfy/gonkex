package mocks_test

import (
	"bytes"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"regexp"
	"strings"
	"testing"

	"github.com/lansfy/gonkex/endpoint"
	"github.com/lansfy/gonkex/mocks"
	"github.com/lansfy/gonkex/models"
	"github.com/lansfy/gonkex/runner"
	"github.com/lansfy/gonkex/testloader/yaml_file"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func init() {
	runner.RegisterFlags()
}

var (
	// This regex matches :[port]/ after 127.0.0.1
	portRegexp = regexp.MustCompile(`127\.0\.0\.1:\d+`)

	// This regex matches :[port]/ after 127.0.0.1
	httpBoundaryRegexp = regexp.MustCompile(`[0-9a-f]{60}`)
)

type errorChecker struct {
	t         *testing.T
	errorInfo string
	lastTest  models.TestInterface
	wasError  bool
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
		input = input[:idx]+", request was..."
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

	if !assert.Equal(c.t, c.getExpected(), normalizeString(content), c.errorInfo) {
		c.wasError = true
	}
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
		!strings.Contains(c.getExpected(), " template:") &&
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

type mockCheckerWrap struct {
	checker mocks.CheckerInterface
}

func (m *mockCheckerWrap) CheckRequest(mockName string, req *http.Request, resp *http.Response) []error {
	return m.checker.CheckRequest(mockName, req, resp)
}

type item struct {
	RequestURL string `json:"request_url"`
	Response   string `json:"response_body"`
}

func multiRequest(h endpoint.Helper) error {
	h.SetStatusCode(200)
	h.SetResponseFormat(endpoint.FormatText)
	wrap := func(err error) error {
		if err != nil {
			err = fmt.Errorf("error: %w", err)
			h.SetResponseRaw([]byte(err.Error()))
		}
		return nil
	}
	data := []item{}
	err := h.GetRequest(&data, endpoint.FormatJson)
	if err != nil {
		return wrap(err)
	}

	client := &http.Client{
		Transport: h.GetMocksTransport(),
	}

	for i := range data {
		resp, err := client.Get("http://someservice" + data[i].RequestURL)
		if err != nil {
			return wrap(err)
		}
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return wrap(err)
		}
		resp.Body.Close()
		if string(body) != data[i].Response {
			return wrap(fmt.Errorf("request #%d: not expected response %q", i, string(body)))
		}
	}
	return nil
}

type customClient struct{}

func (c *customClient) Do(req *http.Request) (*http.Response, error) {
	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	if req.Body == nil {
		return client.Do(req)
	}

	// Read the original body
	bodyBytes, err := io.ReadAll(req.Body)
	if err != nil {
		return nil, fmt.Errorf("internal error: read body: %w", err)
	}
	_ = req.Body.Close()

	modified := httpBoundaryRegexp.ReplaceAllString(string(bodyBytes), "<http-boundary-header>")
	modified = strings.ReplaceAll(modified, "\r\n", "\n")

	req.Body = io.NopCloser(bytes.NewBufferString(modified))
	req.ContentLength = int64(len(modified))

	return client.Do(req)
}

type errorGenerator struct {
	id int
}

func (m *errorGenerator) CheckRequest(mockName string, req *http.Request, resp *http.Response) []error {
	var errs []error
	if req.URL.Path == "/test/checkers" {
		errs = append(errs, fmt.Errorf("called mock checker #%d", m.id))
	}
	return errs
}

type checkerRemover struct {
	m     *mocks.Mocks
	known []mocks.CheckerInterface
}

func (c *checkerRemover) Handler(h endpoint.Helper) error {
	var req struct {
		ID int `json:"id"`
	}
	err := h.GetRequest(&req, endpoint.FormatJson)
	if err != nil {
		return err
	}
	c.m.UnregisterChecker(c.known[req.ID])
	return nil
}

func Test_Declarative(t *testing.T) {
	m := mocks.NewNop("someservice")
	err := m.Start()
	require.NoError(t, err)
	defer m.Shutdown()

	checker := &errorChecker{
		t: t,
	}
	m.RegisterChecker(checker)
	// second item for body preserve test, but it should have other address
	// because RegisterChecker doesn't allow to add same checker twice
	m.RegisterChecker(&mockCheckerWrap{checker})

	g0 := &errorGenerator{0}
	g1 := &errorGenerator{1}
	g2 := &errorGenerator{2}
	m.RegisterChecker(g0)
	m.RegisterChecker(g1)
	m.RegisterChecker(g2)

	remover := &checkerRemover{
		m:     m,
		known: []mocks.CheckerInterface{g0, g1, g2},
	}

	opts := &runner.RunnerOpts{
		Host:        "http://" + m.Service("someservice").ServerAddr(),
		Mocks:       m,
		MocksLoader: mocks.NewYamlLoader(&mocks.YamlLoaderOpts{}),
		TestHandler: checker.Handle,
		HelperEndpoints: endpoint.EndpointMap{
			"multi_request":  multiRequest,
			"remove_checker": remover.Handler,
		},
		CustomClient: &customClient{},
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

func TestRegisterEnvironmentVariables(t *testing.T) {
	m := mocks.NewNop("service1", "service2")
	err := m.Start()
	require.NoError(t, err)
	defer m.Shutdown()

	err = m.RegisterEnvironmentVariables("TEST_")
	require.NoError(t, err)

	require.Equal(t, m.Service("service1").ServerAddr(), os.Getenv("TEST_SERVICE1"))
	require.Equal(t, m.Service("service2").ServerAddr(), os.Getenv("TEST_SERVICE2"))
}
