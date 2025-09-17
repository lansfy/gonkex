package endpoint

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/lansfy/gonkex/mocks"

	"gopkg.in/yaml.v3"
)

type MetaProvider interface {
	GetMeta(key string) interface{}
}

type helperImpl struct {
	headers       map[string][]string
	path          string
	requestBytes  []byte
	responseBytes []byte
	responseCode  int
	contentType   string
	services      *mocks.Mocks
	provider      MetaProvider
}

func newHelper(headers map[string][]string, path string, requestBytes []byte,
	services *mocks.Mocks, provider MetaProvider) *helperImpl {
	return &helperImpl{
		headers:      headers,
		path:         path,
		requestBytes: requestBytes,
		responseCode: http.StatusNoContent,
		contentType:  "application/json",
		services:     services,
		provider:     provider,
	}
}

func internalError(name string, err error) error {
	return fmt.Errorf("internal: %s: %w", name, err)
}

func (h *helperImpl) GetHeaders() map[string][]string {
	return h.headers
}

func (h *helperImpl) GetPath() string {
	return h.path
}

func (h *helperImpl) GetRequestAsJson(v interface{}) error {
	decoder := json.NewDecoder(bytes.NewBuffer(h.requestBytes))
	decoder.DisallowUnknownFields()
	err := decoder.Decode(v)
	if err != nil {
		return internalError("GetRequestAsJson", err)
	}
	return nil
}

func (h *helperImpl) GetRequestAsYaml(v interface{}) error {
	decoder := yaml.NewDecoder(bytes.NewBuffer(h.requestBytes))
	decoder.KnownFields(true)
	err := decoder.Decode(v)
	if err != nil {
		return internalError("GetRequestAsYaml", err)
	}
	return nil
}

func (h *helperImpl) GetRequestAsBytes() ([]byte, error) {
	return h.requestBytes, nil
}

func (h *helperImpl) GetMocksTransport() http.RoundTripper {
	return h.services
}

func (h *helperImpl) GetMockAddr(name string) string {
	if h.services == nil {
		panic(fmt.Sprintf("mock with name %q not exists", name))
	}
	return "http://" + h.services.Service(name).ServerAddr()
}

func (h *helperImpl) GetMeta(key string) interface{} {
	return h.provider.GetMeta(key)
}

func (h *helperImpl) SetResponseAsJson(response interface{}) error {
	b, err := json.Marshal(response)
	if err != nil {
		return internalError("SetResponseAsJson", err)
	}
	h.responseBytes = b
	return nil
}

func (h *helperImpl) SetResponseAsYaml(response interface{}) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = internalError("SetResponseAsYaml", fmt.Errorf("%v", r))
		}
	}()

	var out []byte
	out, err = yaml.Marshal(response)
	if err != nil {
		return internalError("SetResponseAsYaml", err)
	}
	h.contentType = "application/yaml"
	h.responseBytes = out
	return err
}

func (h *helperImpl) SetResponseAsBytes(response []byte) {
	h.responseBytes = response
}

func (h *helperImpl) SetContentType(contentType string) {
	h.contentType = contentType
}

func (h *helperImpl) SetStatusCode(code int) {
	h.responseCode = code
}

func (h *helperImpl) createHTTPResponse() *http.Response {
	if h.responseBytes == nil {
		h.responseBytes = []byte{}
	}

	if h.responseCode == http.StatusNoContent && len(h.responseBytes) > 0 {
		h.responseCode = http.StatusOK
	}

	// Create an HTTP response
	response := &http.Response{
		StatusCode:    h.responseCode,
		Status:        fmt.Sprintf("%d %s", h.responseCode, http.StatusText(h.responseCode)),
		Body:          io.NopCloser(bytes.NewReader(h.responseBytes)),
		Header:        make(http.Header),
		ContentLength: int64(len(h.responseBytes)),
	}

	response.Header.Set("Content-Type", h.contentType)
	return response
}
