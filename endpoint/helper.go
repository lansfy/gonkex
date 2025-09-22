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
	format        Format
}

func newHelper(headers map[string][]string, path string, requestBytes []byte,
	services *mocks.Mocks, provider MetaProvider) *helperImpl {
	return &helperImpl{
		headers:      headers,
		path:         path,
		requestBytes: requestBytes,
		responseCode: http.StatusNoContent,
		services:     services,
		provider:     provider,
		format:       FormatJson,
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

func (h *helperImpl) GetRequest(v interface{}, format Format) error {
	switch format {
	case FormatYaml:
		return h.getRequestAsYaml(v)
	default:
		return h.getRequestAsJson(v)
	}
}

func (h *helperImpl) GetRequestRaw() []byte {
	return h.requestBytes
}

func (h *helperImpl) SetResponseFormat(format Format) {
	h.format = format
}

func (h *helperImpl) SetResponse(v interface{}) error {
	switch h.format {
	case FormatYaml:
		return h.setResponseAsYaml(v)
	default:
		return h.setResponseAsJson(v)
	}
}

func (h *helperImpl) SetResponseRaw(response []byte) {
	h.responseBytes = response
}

func (h *helperImpl) getRequestAsJson(v interface{}) error {
	decoder := json.NewDecoder(bytes.NewBuffer(h.requestBytes))
	decoder.DisallowUnknownFields()
	err := decoder.Decode(v)
	if err != nil {
		return internalError("GetRequestAsJson", err)
	}
	return nil
}

func (h *helperImpl) getRequestAsYaml(v interface{}) error {
	decoder := yaml.NewDecoder(bytes.NewBuffer(h.requestBytes))
	decoder.KnownFields(true)
	err := decoder.Decode(v)
	if err != nil {
		return internalError("GetRequestAsYaml", err)
	}
	return nil
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

func (h *helperImpl) setResponseAsJson(response interface{}) error {
	b, err := json.Marshal(response)
	if err != nil {
		return internalError("SetResponseAsJson", err)
	}
	h.responseBytes = b
	return nil
}

func (h *helperImpl) setResponseAsYaml(response interface{}) (err error) {
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

	response.Header.Set("Content-Type", h.getContentType(false))
	return response
}

func (h *helperImpl) getContentType(override bool) string {
	if h.contentType != "" && !override {
		return h.contentType
	}
	switch h.format {
	case FormatText:
		return "text/plain"
	case FormatYaml:
		return "application/yaml"
	default:
		return "application/json"
	}
}

func (h *helperImpl) setErrorResponse(err error) {
	h.contentType = h.getContentType(true)
	h.SetStatusCode(http.StatusBadRequest)

	var data struct {
		Error string `json:"error" yaml:"error"`
	}
	data.Error = err.Error()
	switch h.format {
	case FormatText:
		h.SetResponseRaw([]byte(fmt.Sprintf("error: %s", err)))
	case FormatYaml:
		_ = h.setResponseAsYaml(data)
	default:
		_ = h.setResponseAsJson(data)
	}
}
