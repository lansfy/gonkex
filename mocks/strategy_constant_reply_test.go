package mocks

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v2"
)

func Test_ConstantReply_HandleRequest(t *testing.T) {
	tests := []struct {
		description string
		content     string
		wantStatus  int
		wantHeaders map[string]string
		wantBody    string
	}{
		{
			description: "simple response",
			content: `
statusCode: 200
headers:
  Content-Type: text/plain
body: "Hello, World!"`,
			wantStatus:  http.StatusOK,
			wantHeaders: map[string]string{"Content-Type": "text/plain"},
			wantBody:    "Hello, World!",
		},
		{
			description: "empty body with headers",
			content: `
statusCode: 204
headers:
  SomeHeader: SomeValue
body: ""`,
			wantStatus:  http.StatusNoContent,
			wantHeaders: map[string]string{"SomeHeader": "SomeValue"},
			wantBody:    "",
		},
		{
			description: "multiple headers",
			content: `
statusCode: 202
headers:
  Header1: Value1
  Header2: Value2
body: "Multi-header test"`,
			wantStatus:  http.StatusAccepted,
			wantHeaders: map[string]string{"Header1": "Value1", "Header2": "Value2"},
			wantBody:    "Multi-header test",
		},
	}

	loader := &loaderImpl{}

	for _, tt := range tests {
		t.Run(tt.description, func(t *testing.T) {
			// Create the reply strategy
			var def map[interface{}]interface{}
			err := yaml.Unmarshal([]byte(tt.content), &def)
			require.NoError(t, err)

			reply, err := loader.loadConstantStrategy(def)
			require.NoError(t, err)

			// Mock request and response
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			rec := httptest.NewRecorder()

			// Call the HandleRequest method
			reply.HandleRequest(rec, req)

			// Assert the response body
			require.Equal(t, tt.wantBody, rec.Body.String())

			// Assert the status code
			require.Equal(t, tt.wantStatus, rec.Code)

			// Assert headers
			for key, value := range tt.wantHeaders {
				require.Equal(t, value, rec.Header().Get(key))
			}
		})
	}
}

func Test_ConstantReplyStrategy(t *testing.T) {
	strategy := NewConstantReplyWithCode([]byte("somebodycontent"), 200, nil)
	m := NewServiceMock("mytest", NewDefinition("$", nil, strategy, CallsNoConstraint))
	err := m.StartServer()
	require.NoError(t, err)

	for i := 0; i < 10; i++ {
		res, err := http.Get("http://" + m.ServerAddr() + "/bar")
		require.NoError(t, err)
		require.Equal(t, 200, res.StatusCode)

		body, _ := io.ReadAll(res.Body)
		res.Body.Close()
		require.Equal(t, "somebodycontent", string(body))
	}
}
