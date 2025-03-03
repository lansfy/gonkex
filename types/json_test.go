package types

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_JSON_GetName(t *testing.T) {
	b := &jsonBodyType{}
	require.Equal(t, "JSON", b.GetName())
}

func Test_JSON_IsSupportedContentType(t *testing.T) {
	b := &jsonBodyType{}
	tests := []struct {
		contentType string
		want        bool
	}{
		{"application/json", true},
		{"text/json", true},
		{"application/yaml", false},
	}

	for _, tt := range tests {
		t.Run(tt.contentType, func(t *testing.T) {
			require.Equal(t, tt.want, b.IsSupportedContentType(tt.contentType))
		})
	}
}

func Test_JSON_Decode(t *testing.T) {
	b := &jsonBodyType{}
	tests := []struct {
		body    string
		want    interface{}
		wantErr string
	}{
		{
			body: `{"key1": "value1", "key2": "value2"}`,
			want: map[string]interface{}{
				"key1": "value1",
				"key2": "value2",
			},
		},
		{
			body:    "invalid_json",
			wantErr: "json: invalid character 'i' looking for beginning of value",
		},
	}

	for _, tt := range tests {
		t.Run(tt.body, func(t *testing.T) {
			got, err := b.Decode(tt.body)
			if tt.wantErr != "" {
				require.Error(t, err)
				require.Equal(t, tt.wantErr, err.Error())
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.want, got)
			}
		})
	}
}

func Test_JSON_ExtractResponseValue(t *testing.T) {
	b := &jsonBodyType{}
	tests := []struct {
		body    string
		path    string
		want    string
		wantErr string
	}{

		{
			body: `{"key1": "value1", "key2": "value2"}`,
			path: "key2",
			want: "value2",
		},
		{
			body: `{"nested": {"key": "value"}}`,
			path: "nested.key",
			want: "value",
		},
		{
			body:    `{"key1": "value1", "key2": "value2"}`,
			path:    "missing",
			wantErr: "path '$.missing' does not exist in service response",
		},
		{
			body:    "invalid_json",
			path:    "key",
			wantErr: "json: invalid character 'i' looking for beginning of value",
		},
	}

	for _, tt := range tests {
		t.Run(tt.body, func(t *testing.T) {
			got, err := b.ExtractResponseValue(tt.body, tt.path)
			if tt.wantErr != "" {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.wantErr)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.want, got)
			}
		})
	}
}
