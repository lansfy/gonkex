package checker_test

import (
	"encoding/json"
	"errors"
	"fmt"
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

type docStorage struct {
	response []json.RawMessage
}

func (s *docStorage) GetType() string {
	return "dummy"
}

func (s *docStorage) LoadFixtures(location string, names []string) error {
	return nil
}

func (s *docStorage) ExecuteQuery(query string) ([]json.RawMessage, error) {
	if s.response == nil {
		return nil, errors.New("fake error")
	}
	resp := s.response
	s.response = nil
	return resp, nil
}

func (s *docStorage) setDBResponse(h endpoint.Helper) error {
	h.SetStatusCode(200)
	wrap := func(err error) error {
		if err != nil {
			err = fmt.Errorf("error: %w", err)
			h.SetResponseAsBytes([]byte(err.Error()))
		}
		return nil
	}
	data := []json.RawMessage{}
	err := h.GetRequestAsJson(&data)
	if err != nil {
		return wrap(err)
	}
	s.response = data

	return nil
}

type errorChecker struct {
	t         *testing.T
	errorInfo string
	lastTest  models.TestInterface
}

func normalizeString(s string) string {
	s = strings.ReplaceAll(s, "\r\n", "\n")
	return strings.TrimSpace(s)
}

func simplifyError(err error) string {
	input := err.Error()
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

	storage := &docStorage{}
	checker := &errorChecker{
		t: t,
	}

	opts := &runner.RunnerOpts{
		Host:        "http://" + m.Service("someservice").ServerAddr(),
		Mocks:       m,
		MocksLoader: mocks.NewYamlLoader(&mocks.YamlLoaderOpts{}),
		DB:          storage,
		TestHandler: checker.Handle,
		HelperEndpoints: endpoint.EndpointMap{
			"set_db_response": storage.setDBResponse,
		},
	}

	r := runner.New(yaml_file.NewLoader("testdata"), opts)
	err = r.Run()
	require.NoError(t, err)
}
