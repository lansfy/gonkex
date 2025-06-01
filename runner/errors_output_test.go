package runner

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"regexp"
	"strings"
	"testing"

	"github.com/lansfy/gonkex/mocks"
	"github.com/lansfy/gonkex/models"
	"github.com/lansfy/gonkex/output/terminal"
	"github.com/lansfy/gonkex/testloader/yaml_file"
	"github.com/lansfy/gonkex/variables"

	"github.com/stretchr/testify/require"
)

const showOnScreen = false // get output on screen to debug colors
var dateRegexp = regexp.MustCompile("(Mon|Tue|Wed|Thu|Fri|Sat|Sun), ([0-3][0-9]) (Jan|Feb|Mar|Apr|May|Jun|Jul|Aug|Sep|Oct|Nov|Dec) ([0-9]{4}) ([01][0-9]|2[0-3])(:[0-5][0-9]){2} GMT")

type fakeStorage struct{}

func (f *fakeStorage) GetType() string {
	return "dummy"
}

func (f *fakeStorage) LoadFixtures(location string, names []string) error {
	return nil
}

func (f *fakeStorage) ExecuteQuery(query string) ([]json.RawMessage, error) {
	return []json.RawMessage{
		json.RawMessage("{\"field1\":\"value1\"}"),
		json.RawMessage("{\"field2\":123}"),
	}, nil
}

func initErrorServer() {
	http.HandleFunc("/text", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Add("Content-Type", "text/plain")
		_, _ = w.Write([]byte("1234"))
	})
	http.HandleFunc("/json", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Add("Content-Type", "application/json")
		_, _ = w.Write([]byte("{\"somefield\":123}"))
	})
}

func normalize(s string) string {
	s = strings.ReplaceAll(s, "\r\n", "\n")
	s = dateRegexp.ReplaceAllString(s, "Sat, 1 Dec 2024 00:00:00 GMT")
	return s
}

func testHandler(test models.TestInterface, executeTest TestExecutor) (bool, error) {
	_, err := executeTest(test)
	if err != nil && !isTestWasSkipped(err) {
		return false, err
	}
	return false, nil
}

func Test_Error_Examples(t *testing.T) {
	initErrorServer()
	server := httptest.NewServer(nil)

	for caseID := 1; caseID <= 7; caseID++ {
		t.Run(fmt.Sprintf("case%d", caseID), func(t *testing.T) {
			expected, err := os.ReadFile(fmt.Sprintf("testdata/errors-example/case%d_output.txt", caseID))
			require.NoError(t, err)

			m := mocks.NewNop("subservice")
			err = m.Start()
			require.NoError(t, err)
			defer m.Shutdown()

			yamlLoader := yaml_file.NewLoader(fmt.Sprintf("testdata/errors-example/case%d.yaml", caseID))
			r := New(
				yamlLoader,
				&RunnerOpts{
					Host:         server.URL,
					Variables:    variables.New(),
					Mocks:        m,
					MocksLoader:  mocks.NewYamlLoader(nil),
					DB:           &fakeStorage{},
					TestHandler:  testHandler,
					OnFailPolicy: PolicyContinue,
				},
			)

			buf := &strings.Builder{}
			opts := &terminal.OutputOpts{}
			if !showOnScreen {
				opts.Policy = terminal.PolicyForceNoColor
				opts.CustomWriter = buf
			}

			output := terminal.NewOutput(opts)
			r.AddOutput(output)

			err = r.Run()
			require.NoError(t, err)

			if !showOnScreen {
				require.Equal(t, normalize(string(expected)), normalize(buf.String()))
			}
		})
	}
	require.Equal(t, false, showOnScreen)
}
