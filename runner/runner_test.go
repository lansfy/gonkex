package runner

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/lansfy/gonkex/endpoint"
	"github.com/lansfy/gonkex/mocks"
	"github.com/lansfy/gonkex/models"
	"github.com/lansfy/gonkex/output"
	"github.com/lansfy/gonkex/testloader/yaml_file"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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

type variablesServer struct {
	t          *testing.T
	counter    int
	subservice string
}

func (s *variablesServer) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	s.counter++
	requestBytes, err := io.ReadAll(r.Body)
	if err != nil {
		assert.NoError(s.t, err)
		return
	}

	_ = r.Body.Close()

	var data struct {
		Counter     int `json:"counter"`
		EvenCounter int `json:"even_counter"`
	}

	decoder := json.NewDecoder(bytes.NewBuffer(requestBytes))
	decoder.DisallowUnknownFields()
	err = decoder.Decode(&data)
	if err != nil {
		assert.NoError(s.t, err)
		return
	}

	assert.Equal(s.t, s.counter, data.Counter)
	assert.Equal(s.t, 100+s.counter/2, data.EvenCounter)

	resp, err := http.Get(fmt.Sprintf("http://%s/", s.subservice))
	if err != nil {
		assert.NoError(s.t, err)
		return
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		assert.NoError(s.t, err)
		return
	}

	assert.Equal(s.t, fmt.Sprintf("%d", s.counter), string(body))

	rw.Header().Set("Content-Type", "application/json")
	rw.WriteHeader(http.StatusOK)
	content := fmt.Sprintf(`{"counter":%d, "even_counter":%d}`, s.counter, 100+s.counter/2)
	_, err = rw.Write([]byte(content))
	assert.NoError(s.t, err)
}

func Test_variablesSubstitution(t *testing.T) {
	m := mocks.NewNop("subservice")
	err := m.Start()
	require.NoError(t, err)
	defer m.Shutdown()

	srv := httptest.NewServer(&variablesServer{
		subservice: m.Service("subservice").ServerAddr(),
		t:          t,
	})
	defer srv.Close()

	RunWithTesting(t, srv.URL, &RunWithTestingOpts{
		TestsDir: "testdata/variables/case-substitution.yaml",
		Mocks:    m,
	})
}

type statusServer struct {
	counter      int
	totalTests   int
	skippedTests int
	brokenTests  int
}

func (s *statusServer) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	rw.Header().Set("Content-Type", "application/json")
	rw.WriteHeader(http.StatusOK)
	s.counter++
	_, _ = rw.Write([]byte(fmt.Sprintf(`{"calls":%d}`, s.counter)))
}

func (s *statusServer) BeforeTest(v models.TestInterface) error {
	s.totalTests++
	if v.GetStatus() == models.StatusBroken {
		s.brokenTests++
	}
	if v.GetStatus() == models.StatusSkipped {
		s.skippedTests++
	}
	return nil
}

func (s *statusServer) Process(models.TestInterface, *models.Result) error {
	return nil
}

func Test_status(t *testing.T) {
	testCases := []struct {
		name    string
		total   int
		skipped int
		broken  int
	}{
		{"broken_one.yaml", 1, 0, 1},
		{"broken_many.yaml", 5, 0, 3},
		{"focus_one.yaml", 1, 0, 0},
		{"focus_many.yaml", 5, 4, 0},
		{"skipped_one.yaml", 1, 1, 0},
		{"skipped_many.yaml", 5, 3, 0},
		{"skipped_with_retry.yaml", 1, 1, 0},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			obj := &statusServer{}
			srv := httptest.NewServer(obj)
			defer srv.Close()

			RunWithTesting(t, srv.URL, &RunWithTestingOpts{
				TestsDir: "testdata/status/" + tt.name,
				Outputs:  []output.OutputInterface{obj},
			})
			require.Equal(t, tt.total, obj.totalTests)
			require.Equal(t, tt.skipped, obj.skippedTests)
			require.Equal(t, tt.broken, obj.brokenTests)
		})
	}
}
