package endpoint

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/lansfy/gonkex/mocks"

	"gopkg.in/yaml.v2"
)

type MetaProvider interface {
	GetMeta(key string) interface{}
}

type helperImpl struct {
	path          string
	requestBytes  []byte
	responseBytes []byte
	responseCode  int
	contentType   string
	services      *mocks.Mocks
	provider      MetaProvider
}

func newHelper(path string, requestBytes []byte, services *mocks.Mocks, provider MetaProvider) *helperImpl {
	return &helperImpl{
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

// GetPath returns request path without Prefix
func (h *helperImpl) GetPath() string {
	return h.path
}

// GetRequestAsJson unmarshals the request bytes into the provided object.
func (h *helperImpl) GetRequestAsJson(v interface{}) error {
	decoder := json.NewDecoder(bytes.NewBuffer(h.requestBytes))
	decoder.DisallowUnknownFields()
	err := decoder.Decode(v)
	if err != nil {
		return internalError("GetRequestAsJson", err)
	}
	return nil
}

// GetRequestAsYaml unmarshals the request bytes into the provided object as YAML.
func (h *helperImpl) GetRequestAsYaml(v interface{}) error {
	err := yaml.UnmarshalStrict(h.requestBytes, v)
	if err != nil {
		return internalError("GetRequestAsYaml", err)
	}
	return nil
}

// GetRequestAsBytes returns the raw request bytes.
func (h *helperImpl) GetRequestAsBytes() ([]byte, error) {
	return h.requestBytes, nil
}

// GetMockAddr returns address of mock with specified name
func (h *helperImpl) GetMockAddr(name string) string {
	if h.services == nil {
		panic(fmt.Sprintf("mock with name %q not exists", name))
	}
	return "http://" + h.services.Service(name).ServerAddr()
}

func (h *helperImpl) GetMeta(key string) interface{} {
	return h.provider.GetMeta(key)
}

// SetResponseAsJson marshals the provided object into JSON and stores it as the response.
func (h *helperImpl) SetResponseAsJson(response interface{}) error {
	b, err := json.Marshal(response)
	if err != nil {
		return internalError("SetResponseAsJson", err)
	}
	h.responseBytes = b
	return nil
}

// SetResponseAsYaml marshals the provided object into YAML and stores it as the response.
func (h *helperImpl) SetResponseAsYaml(response interface{}) error {
	b, err := yaml.Marshal(response)
	if err != nil {
		return internalError("SetResponseAsYaml", err)
	}
	h.contentType = "application/yaml"
	h.responseBytes = b
	return nil
}

// SetResponseAsBytes sets the raw response bytes.
func (h *helperImpl) SetResponseAsBytes(response []byte) error {
	h.responseBytes = response
	return nil
}

// SetStatusCode sets the HTTP response status code.
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
		Body:          io.NopCloser(bytes.NewReader(h.responseBytes)),
		Header:        make(http.Header),
		ContentLength: int64(len(h.responseBytes)),
	}

	response.Header.Set("Content-Type", h.contentType)
	return response
}
