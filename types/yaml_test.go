package types

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_YAML_GetName(t *testing.T) {
	b := &yamlBodyType{}
	require.Equal(t, "YAML", b.GetName())
}

func Test_YAML_IsSupportedContentType(t *testing.T) {
	b := &yamlBodyType{}
	tests := []struct {
		contentType string
		want        bool
	}{
		{"application/x-yaml", true},
		{"text/yaml", true},
		{"application/json", false},
	}

	for _, tt := range tests {
		t.Run(tt.contentType, func(t *testing.T) {
			require.Equal(t, tt.want, b.IsSupportedContentType(tt.contentType))
		})
	}
}

func Test_YAML_Decode(t *testing.T) {
	b := &yamlBodyType{}
	tests := []struct {
		body    string
		want    interface{}
		wantErr string
	}{
		{
			body: "key1: value1\nkey2: value2",
			want: map[string]interface{}{
				"key1": "value1",
				"key2": "value2",
			},
		},
		{
			body:    "invalid_yaml: [unclosed",
			wantErr: "yaml: line 1: did not find expected ',' or ']'",
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

func Test_YAML_ExtractResponseValue(t *testing.T) {
	b := &yamlBodyType{}
	tests := []struct {
		body    string
		path    string
		want    string
		wantErr string
	}{
		{
			body: "key1: value1\nkey2: value2",
			path: "key2",
			want: "value2",
		},
		{
			body: "nested:\n  key: value",
			path: "nested.key",
			want: "value",
		},
		{
			body:    "key1: value1\nkey2: value2",
			path:    "missing",
			wantErr: "path '$.missing' does not exist in service response",
		},
		{
			body:    "invalid_yaml: [unclosed",
			path:    "key",
			wantErr: "yaml: line 1: did not find expected ',' or ']'",
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
