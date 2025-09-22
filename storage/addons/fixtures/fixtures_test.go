package fixtures

import (
	"errors"
	"fmt"
	"sort"
	"testing"

	"github.com/lansfy/gonkex/endpoint"
	"github.com/lansfy/gonkex/runner"
)

type testLoaderImpl struct {
	known map[string]string
}

func (l *testLoaderImpl) Load(name string) (string, []byte, error) {
	content, ok := l.known[name]
	if !ok {
		return "", nil, errors.New("file not found")
	}

	l.known[name] = ""
	return name, []byte(content), nil
}

func process(h endpoint.Helper) error {
	h.SetStatusCode(200)
	h.SetResponseFormat(endpoint.FormatText)
	wrap := func(err error) error {
		if err != nil {
			h.SetResponseRaw([]byte(fmt.Sprintf("error: %v", err)))
		}
		return nil
	}

	var input struct {
		Types []string          `yaml:"types"`
		Names []string          `yaml:"names"`
		FS    map[string]string `yaml:"fs"`
	}
	err := h.GetRequest(&input, endpoint.FormatYaml)
	if err != nil {
		return wrap(err)
	}

	known := input.FS
	if known == nil {
		known = map[string]string{}
	}

	if input.Types == nil {
		input.Types = []string{"collections"}
	}

	opts := &LoadDataOpts{
		AllowedTypes: input.Types,
		CustomActions: map[string]func(string) string{
			"custom_action": func(value string) string {
				return "!!!" + value + "!!!"
			},
		},
	}

	coll, err := LoadData(&testLoaderImpl{known}, input.Names, opts)
	if err != nil {
		return wrap(err)
	}

	// sort result to preserve element order
	sort.Slice(coll, func(i, j int) bool {
		if coll[i].Type == coll[j].Type {
			return coll[i].Name < coll[j].Name
		}
		return coll[i].Type < coll[j].Type
	})

	data, err := DumpCollection(coll, len(input.Types) != 1)
	if err != nil {
		return wrap(err)
	}

	h.SetResponseRaw(data)
	return nil
}

func init() {
	runner.RegisterFlags()
}

func Test_generateTestFixtures(t *testing.T) {
	opts := &runner.RunWithTestingOpts{
		TestsDir:     "testdata/tests",
		OnFailPolicy: runner.PolicyContinue,
		HelperEndpoints: endpoint.EndpointMap{
			"process": process,
		},
	}
	runner.RunWithTesting(t, "http://localhost", opts)
}
