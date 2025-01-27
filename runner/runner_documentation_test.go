package runner

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
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

func docServer() {
	http.HandleFunc("/test/vars-from-response-currently-running-test", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Add("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"result_id": "1", "query_result": [[42, "golang"], [2, "gonkex"]]}`))
	})
	http.HandleFunc("/test/vars-usage/some-value", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "application/json")

		if r.Method != "GET" {
			_, _ = w.Write([]byte("wrong Method received"))
			return
		}

		if r.URL.Query()["param"][0] != "some-value" {
			_, _ = w.Write([]byte("wrong query received"))
		}

		if r.Header.Get("Header1") != "some-value" {
			_, _ = w.Write([]byte("wrong header received"))
		}

		body, err := io.ReadAll(r.Body)
		if err != nil {
			_, _ = w.Write([]byte(err.Error()))
			return
		}

		bodyStr := strings.ReplaceAll(strings.ReplaceAll(string(body), " ", ""), "\n", "")
		if bodyStr != `{"reqParam":"some-value"}` {
			_, _ = w.Write([]byte(fmt.Sprintf("received wrong body: %s", string(body))))
			return
		}

		_, _ = w.Write([]byte(`{"data":"some-value"}`))
	})
}

func Test_Documentation_Examples(t *testing.T) {
	docServer()

	srv := httptest.NewServer(nil)
	defer srv.Close()

	RunWithTesting(t, srv.URL, &RunWithTestingOpts{
		TestsDir: "testdata/documentation",
		DB:       &docStorage{},
	})
}
