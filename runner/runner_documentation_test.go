package runner

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/lansfy/gonkex/mocks"

	"github.com/stretchr/testify/require"
)

type docStorage struct{}

func (f *docStorage) GetType() string {
	return "dummy"
}

func (f *docStorage) LoadFixtures(location string, names []string) error {
	return nil
}

func (f *docStorage) ExecuteQuery(query string) ([]json.RawMessage, error) {
	if query == "SELECT id, name FROM testing_tools WHERE id=42" {
		return []json.RawMessage{
			json.RawMessage(`{"id": 42, "name": "golang"}`),
		}, nil
	}
	return nil, fmt.Errorf("wrong request to DB received: %q", query)
}

func Test_Documentation_Examples(t *testing.T) {
	m := mocks.NewNop("testservice")
	err := m.Start()
	require.NoError(t, err)
	defer m.Shutdown()

	RunWithTesting(t, "http://"+m.Service("testservice").ServerAddr(), &RunWithTestingOpts{
		TestsDir: "testdata/documentation",
		DB:       &docStorage{},
		Mocks:    m,
	})
}
