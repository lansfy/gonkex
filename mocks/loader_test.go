package mocks

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLoadHeaders(t *testing.T) {
	tests := []struct {
		description string
		input       map[interface{}]interface{}
		want        map[string]string
		wantErr     string
	}{
		{
			description: "valid headers",
			input: map[interface{}]interface{}{
				"headers": map[interface{}]interface{}{
					"Header1": "value1",
					"Header2": "value2",
				},
			},
			want: map[string]string{
				"Header1": "value1",
				"Header2": "value2",
			},
			wantErr: "",
		},
		{
			description: "headers is not a map",
			input: map[interface{}]interface{}{
				"headers": "invalid",
			},
			want:    nil,
			wantErr: "`headers` must be a map",
		},
		{
			description: "header key is not a string",
			input: map[interface{}]interface{}{
				"headers": map[interface{}]interface{}{
					123: "value",
				},
			},
			want:    nil,
			wantErr: "`headers` requires string keys",
		},
		{
			description: "header value is not a string",
			input: map[interface{}]interface{}{
				"headers": map[interface{}]interface{}{
					"key": 123,
				},
			},
			want:    nil,
			wantErr: "`headers` requires string values",
		},
		{
			description: "no headers key",
			input:       map[interface{}]interface{}{},
			want:        nil,
			wantErr:     "",
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			result, err := loadHeaders(test.input)

			if test.wantErr != "" {
				require.EqualError(t, err, test.wantErr)
			} else {
				require.NoError(t, err)
			}

			require.Equal(t, test.want, result)
		})
	}
}
