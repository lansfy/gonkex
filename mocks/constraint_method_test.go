package mocks

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_loadMethodConstraint(t *testing.T) {
	tests := []struct {
		description string
		def         map[interface{}]interface{}
		want        verifier
		wantErr     string
	}{
		{
			description: "set valid method MUST be successful",
			def: map[interface{}]interface{}{
				"method": "GET",
			},
			want: &methodConstraint{name: "methodIs", method: "GET"},
		},
		{
			description: "missing method key MUST fail",
			def:         map[interface{}]interface{}{},
			wantErr:     "'method' key required",
		},
		{
			description: "set invalid method type MUST fail",
			def: map[interface{}]interface{}{
				"method": 123,
			},
			wantErr: "key 'method' has non-string value",
		},
	}

	for _, tt := range tests {
		t.Run(tt.description, func(t *testing.T) {
			got, err := loadMethodConstraint(tt.def)
			if tt.wantErr == "" {
				require.NoError(t, err)
				require.Equal(t, tt.want, got)
			} else {
				require.Error(t, err)
				require.EqualError(t, err, tt.wantErr)
				require.Nil(t, got)
			}
		})
	}
}

func Test_methodConstraint_GetName(t *testing.T) {
	c := &methodConstraint{"name1", ""}
	got := c.GetName()
	require.Equal(t, "name1", got)
}

func Test_methodConstraint_Verify(t *testing.T) {
	tests := []struct {
		description string
		method      string
		request     *http.Request
		wantErr     string
	}{
		{
			description: "method matches",
			method:      "GET",
			request:     &http.Request{Method: "GET"},
			wantErr:     "",
		},
		{
			description: "method does not match",
			method:      "POST",
			request:     &http.Request{Method: "GET"},
			wantErr:     "method does not match: expected GET, actual POST",
		},
	}

	for _, tt := range tests {
		t.Run(tt.description, func(t *testing.T) {
			c := &methodConstraint{
				name:   "methodIs",
				method: tt.method,
			}
			got := c.Verify(tt.request)
			if tt.wantErr == "" {
				require.Nil(t, got)
			} else {
				require.NotNil(t, got)
				require.EqualError(t, got[0], tt.wantErr)
			}
		})
	}
}
