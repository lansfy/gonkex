package endpoint

import (
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/lansfy/gonkex/mocks"

	"github.com/tidwall/match"
)

func runEndpoint(e Endpoint, path string, req *http.Request, services *mocks.Mocks, meta MetaProvider) (*http.Response, error) {
	requestBytes, err := io.ReadAll(req.Body)
	if err != nil {
		return nil, err
	}

	_ = req.Body.Close()

	helper := newHelper(path, requestBytes, services, meta)
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

func SelectEndpoint(m EndpointMap, path string, req *http.Request,
	services *mocks.Mocks, meta MetaProvider) (*http.Response, error) {
	path = path[len(Prefix):]
	for name, endpoint := range m {
		if !match.Match(path, name) {
			continue
		}
		return runEndpoint(endpoint, path, req, services, meta)
	}
	available := []string{}
	for name := range m {
		available = append(available, Prefix+name)
	}
	return nil, fmt.Errorf("helper endpoint %q not found (available: %s)", Prefix+path, strings.Join(available, ","))
}
