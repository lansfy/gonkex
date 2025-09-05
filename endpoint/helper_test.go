package endpoint

import (
	"encoding/json"
	"io"
	"net/http"
	"testing"

	"github.com/lansfy/gonkex/mocks"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func newTestHelper(reqBytes []byte, provider MetaProvider, srv *mocks.Mocks) *helperImpl {
	return newHelper(
		map[string][]string{"X-Test": {"1"}},
		"/some/path",
		reqBytes,
		srv,
		provider,
	)
}

func TestHelper_GetHeaders(t *testing.T) {
	h := newTestHelper(nil, nil, nil)
	require.Equal(t, map[string][]string{"X-Test": {"1"}}, h.GetHeaders())
}

func TestHelper_GetPath(t *testing.T) {
	h := newTestHelper(nil, nil, nil)
	require.Equal(t, "/some/path", h.GetPath())
}

var input1 = map[string]any{
	"key1": "value1",
	"key2": "value2",
}

func TestHelper_GetRequestAsJson(t *testing.T) {
	reqJSON, err := json.Marshal(input1)
	require.NoError(t, err)

	h := newTestHelper(reqJSON, nil, nil)

	var out map[string]any
	err = h.GetRequestAsJson(&out)
	require.NoError(t, err)
	require.Equal(t, input1, out)
}

func TestHelper_GetRequestAsJson_Error(t *testing.T) {
	h := newTestHelper([]byte("{\"corrupted\":\"json"), nil, nil)

	var out map[string]any
	err := h.GetRequestAsJson(&out)
	require.EqualError(t, err, "internal: GetRequestAsJson: unexpected EOF")
}

func TestHelper_GetRequestAsYaml(t *testing.T) {
	reqYAML, err := yaml.Marshal(input1)
	require.NoError(t, err)

	h := newTestHelper(reqYAML, nil, nil)

	var out map[string]any
	err = h.GetRequestAsYaml(&out)
	require.NoError(t, err)
	require.Equal(t, input1, out)
}

func TestHelper_GetRequestAsYaml_Error(t *testing.T) {
	h := newTestHelper([]byte("key1: value1\n  key2: with wrong offset"), nil, nil)

	var out map[string]any
	err := h.GetRequestAsYaml(&out)
	require.EqualError(t, err, "internal: GetRequestAsYaml: yaml: line 2: mapping values are not allowed in this context")
}

func TestHelper_GetRequestAsBytes(t *testing.T) {
	input := []byte("some-bytes")
	h := newTestHelper(input, nil, nil)

	out, err := h.GetRequestAsBytes()
	require.NoError(t, err)
	require.Equal(t, input, out)
}

func TestHelper_GetMocksTransport(t *testing.T) {
	m := mocks.New()
	h := newTestHelper(nil, nil, m)
	require.Equal(t, m, h.GetMocksTransport())
}

func TestHelper_GetMockAddr(t *testing.T) {
	m := mocks.NewNop("backend")
	err := m.Start()
	require.NoError(t, err)
	defer m.Shutdown()

	h := newTestHelper(nil, nil, m)
	require.Equal(t, "http://"+m.Service("backend").ServerAddr(), h.GetMockAddr("backend"))
}

type meta struct{}

func (m *meta) GetMeta(key string) interface{} {
	switch key {
	case "key1":
		return "value1"
	case "key2":
		return input1
	default:
		return nil
	}
}

func TestHelper_GetMeta(t *testing.T) {
	h := newTestHelper(nil, &meta{}, nil)
	require.Equal(t, "value1", h.GetMeta("key1"))
	require.Equal(t, input1, h.GetMeta("key2"))
	require.Nil(t, h.GetMeta("key3"))
}

var input2 = map[string]string{
	"key1": "value1",
	"key2": "value2",
}

func TestHelper_SetResponseAsJson(t *testing.T) {
	h := newTestHelper(nil, nil, nil)

	err := h.SetResponseAsJson(input2)
	require.NoError(t, err)

	var decoded map[string]string
	err = json.Unmarshal(h.responseBytes, &decoded)
	require.NoError(t, err)
	require.Equal(t, input2, decoded)
}

func TestHelper_SetResponseAsJson_Error(t *testing.T) {
	type Invalid struct {
		Data chan int // channels cannot be marshaled
	}
	invalidMap := Invalid{Data: make(chan int)}

	h := newTestHelper(nil, nil, nil)
	err := h.SetResponseAsJson(invalidMap)
	require.EqualError(t, err, "internal: SetResponseAsJson: json: unsupported type: chan int")
}

func TestHelper_SetResponseAsYaml(t *testing.T) {
	h := newTestHelper(nil, nil, nil)

	err := h.SetResponseAsYaml(input2)
	require.NoError(t, err)

	var decoded map[string]string
	err = yaml.Unmarshal(h.responseBytes, &decoded)
	require.NoError(t, err)
	require.Equal(t, input2, decoded)
	require.Equal(t, "application/yaml", h.contentType)
}

func TestHelper_SetResponseAsYaml_Error(t *testing.T) {
	type Invalid struct {
		Fn func() // functions can't be marshaled
	}
	invalidMap := Invalid{Fn: func() {}}

	h := newTestHelper(nil, nil, nil)
	err := h.SetResponseAsYaml(invalidMap)
	require.EqualError(t, err, "internal: SetResponseAsYaml: cannot marshal type: func()")
}

func TestHelper_SetResponseAsBytes(t *testing.T) {
	h := newTestHelper(nil, nil, nil)
	bytesIn := []byte("raw-response")

	h.SetResponseAsBytes(bytesIn)
	require.Equal(t, bytesIn, h.responseBytes)
}

func TestHelper_SetStatusCode(t *testing.T) {
	h := newTestHelper(nil, nil, nil)
	h.SetStatusCode(http.StatusTeapot)
	require.Equal(t, http.StatusTeapot, h.responseCode)
}

func TestHelper_createHTTPResponse_Default(t *testing.T) {
	h := newTestHelper(nil, nil, nil)
	resp := h.createHTTPResponse()
	defer resp.Body.Close()

	require.Equal(t, http.StatusNoContent, resp.StatusCode)
	require.Equal(t, "application/json", resp.Header.Get("Content-Type"))
	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	require.Empty(t, body)
}

func TestHelper_createHTTPResponse_OverridesStatusCode(t *testing.T) {
	h := newTestHelper(nil, nil, nil)
	h.SetResponseAsBytes([]byte("non-empty"))
	resp := h.createHTTPResponse()
	defer resp.Body.Close()

	require.Equal(t, http.StatusOK, resp.StatusCode)
	body, _ := io.ReadAll(resp.Body)
	require.Equal(t, "non-empty", string(body))
	require.Equal(t, "application/json", resp.Header.Get("Content-Type"))
}
