package sqldb_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/lansfy/gonkex/endpoint"
	"github.com/lansfy/gonkex/runner"
	"github.com/lansfy/gonkex/storage/addons/sqldb"
)

type testLoaderImpl struct {
	known map[string]string
}

func (l *testLoaderImpl) Load(name string) (string, []byte, error) {
	content, ok := l.known[name]
	if !ok {
		return "", nil, errors.New("file not exists")
	}

	l.known[name] = ""
	return name, []byte(content), nil
}

func process(h endpoint.Helper) error {
	h.SetStatusCode(200)
	h.SetContentType("application/text")
	wrap := func(err error) error {
		if err != nil {
			_ = h.SetResponseAsBytes([]byte(fmt.Sprintf("error: %v", err)))
		}
		return nil
	}

	var input struct {
		Names []string          `yaml:"names"`
		FS    map[string]string `yaml:"fs"`
	}
	err := h.GetRequestAsYaml(&input)
	if err != nil {
		return wrap(err)
	}

	known := input.FS
	if known == nil {
		known = map[string]string{}
	}

	data, err := sqldb.ConvertToTestFixtures(&testLoaderImpl{known}, input.Names)
	if err != nil {
		return wrap(err)
	}

	return h.SetResponseAsBytes(data)
}

func init() {
	runner.RegisterFlags()
}

func Test_generateTestFixtures(t *testing.T) {
	opts := &runner.RunWithTestingOpts{
		TestsDir:     "testdata/fixtures_tests",
		OnFailPolicy: runner.PolicyContinue,
		HelperEndpoints: endpoint.EndpointMap{
			"process": process,
		},
	}
	runner.RunWithTesting(t, "http://localhost", opts)
}
