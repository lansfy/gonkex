package endpoint

import (
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/lansfy/gonkex/mocks"

	"github.com/tidwall/match"
)

func runEndpoint(path string, services *mocks.Mocks, e Endpoint, req *http.Request) (*http.Response, error) {
	requestBytes, err := io.ReadAll(req.Body)
	if err != nil {
		return nil, err
	}

	_ = req.Body.Close()

	helper := newHelper(path, requestBytes, services)
	err = e(helper)
	if err != nil {
		var data struct {
			Error string `json:"error"`
		}
		data.Error = err.Error()
		_ = helper.SetResponseAsJson(&data)
		helper.SetStatusCode(http.StatusBadRequest)
	}
	return helper.createHTTPResponse(), nil
}

func SelectEndpoint(services *mocks.Mocks, m EndpointMap, path string, req *http.Request) (*http.Response, error) {
	path = path[len(Prefix):]
	for name, endpoint := range m {
		if !match.Match(path, name) {
			continue
		}
		return runEndpoint(path, services, endpoint, req)
	}
	available := []string{}
	for name := range m {
		available = append(available, Prefix+name)
	}
	return nil, fmt.Errorf("helper endpoint %q not found (available: %s)", Prefix+path, strings.Join(available, ","))
}
