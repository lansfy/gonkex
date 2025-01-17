package runner

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/lansfy/gonkex/endpoint"
	"github.com/lansfy/gonkex/models"
	"github.com/lansfy/gonkex/testloader/yaml_file"

	"github.com/stretchr/testify/require"
)

func testServerRedirect() *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/redirect-url", http.StatusFound)
	}))
}

func Test_dontFollowRedirects(t *testing.T) {
	srv := testServerRedirect()
	defer srv.Close()

	RunWithTesting(t, srv.URL, &RunWithTestingOpts{
		TestsDir: "testdata/dont-follow-redirects",
	})
}

type failEndpoint struct {
	expectedError string
	pattern       string
	index         int
}

func (e *failEndpoint) Run(h endpoint.Helper) error {
	pattern := strings.Split(h.GetPath(), "/")[1]
	if e.pattern != pattern {
		e.pattern = pattern
		e.index = -1
	}
	e.index++
	if e.index >= len(e.pattern) {
		return fmt.Errorf("end of iteration")
	}
	if e.pattern[e.index] == '1' {
		return nil
	}
	return fmt.Errorf("fake error")
}

func (e *failEndpoint) Handler(t models.TestInterface, f TestExecutor) error {
	e.expectedError = t.GetDescription()
	return defaultTestHandler(t, f)
}

func Test_retries(t *testing.T) {
	endpoint.Prefix = "/test."

	testCases := []string{
		"success.yaml",
		"failure1.yaml",
		"failure2.yaml",
		"failure3.yaml",
		"failure4.yaml",
	}

	for _, file := range testCases {
		t.Run(file, func(t *testing.T) {
			yamlLoader := yaml_file.NewLoader("testdata/retry/" + file)

			e := &failEndpoint{}

			runner := New(
				yamlLoader,
				&RunnerOpts{
					HelperEndpoints: endpoint.EndpointMap{
						"run/*": e.Run,
					},
					TestHandler: e.Handler,
				},
			)
			err := runner.Run()
			if e.expectedError != "" {
				require.Error(t, err)
				require.Contains(t, err.Error(), e.expectedError)
			} else {
				require.NoError(t, err)
				require.Equal(t, len(e.pattern)-1, e.index)
			}
		})
	}
}
