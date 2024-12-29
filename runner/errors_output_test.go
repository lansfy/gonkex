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

	"github.com/lansfy/gonkex/checker/response_body"
	"github.com/lansfy/gonkex/checker/response_db"
	"github.com/lansfy/gonkex/checker/response_header"
	"github.com/lansfy/gonkex/output/terminal"
	"github.com/lansfy/gonkex/testloader/yaml_file"
	"github.com/lansfy/gonkex/variables"

	"github.com/stretchr/testify/require"
)

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

func Test_Error_Examples(t *testing.T) {
	initErrorServer()
	server := httptest.NewServer(nil)

	for caseID := 1; caseID <= 3; caseID++ {
		t.Run(fmt.Sprintf("case%d", caseID), func(t *testing.T) {
			expected, err := os.ReadFile(fmt.Sprintf("testdata/errors-example/case%d_output.txt", caseID))
			require.NoError(t, err)

			testHandler := NewConsoleHandler()
			yamlLoader := yaml_file.NewLoader(fmt.Sprintf("testdata/errors-example/case%d.yaml", caseID))
			r := New(
				&Config{
					Host:      server.URL,
					Variables: variables.New(),
				},
				yamlLoader,
				testHandler.HandleTest,
			)

			buf := &strings.Builder{}
			output := terminal.NewOutput(&terminal.OutputOpts{
				Policy:       terminal.PolicyForceNoColor,
				CustomWriter: buf,
			})
			r.AddOutput(output)
			r.AddCheckers(response_body.NewChecker())
			r.AddCheckers(response_header.NewChecker())
			r.AddCheckers(response_db.NewChecker(&fakeStorage{}))

			err = r.Run()
			require.NoError(t, err)

			require.Equal(t, normalize(string(expected)), normalize(buf.String()))
		})
	}
}
