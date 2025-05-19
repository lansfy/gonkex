package compare

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_extractArgs(t *testing.T) {
	defaultParams := map[string]string{
		"param1": "default1",
		"param2": "default2",
	}
	tests := []struct {
		description string
		input       string
		wantParams  map[string]string
		wantErr     string
	}{
		{
			description: "only value, no additional params",
			input:       "somevalue",
			wantParams:  defaultParams,
		},
		{
			description: "value with one additional param",
			input:       "somevalue,param1=aaa",
			wantParams: map[string]string{
				"param1": "aaa",
				"param2": "default2",
			},
		},
		{
			description: "value with two additional params",
			input:       "somevalue,param2=bbb,param1=aaa",
			wantParams: map[string]string{
				"param1": "aaa",
				"param2": "bbb",
			},
		},
		{
			description: "value with two additional params (and spaces in name)",
			input:       "somevalue, param2 =bbb, param1=aaa",
			wantParams: map[string]string{
				"param1": "aaa",
				"param2": "bbb",
			},
		},
		{
			description: "invalid parameter format (no '=') not allowed",
			input:       "somevalue,bar",
			wantErr:     "parameter 'bar': invalid parameter format",
		},
		{
			description: "empty parameter name not allowed",
			input:       "somevalue,=value",
			wantErr:     "parameter '=value': empty parameter name",
		},
		{
			description: "unknown name in parameters",
			input:       "somevalue,baz=value",
			wantErr:     "parameter 'baz=value': unknown parameter name",
		},
		{
			description: "duplicate key in param",
			input:       "somevalue,param1=value1,param1=value2",
			wantErr:     "parameter 'param1=value2': duplicate parameter name",
		},
	}

	for _, tt := range tests {
		t.Run(tt.description, func(t *testing.T) {
			gotBase, gotParams, err := extractArgs(tt.input, defaultParams)

			if tt.wantErr != "" {
				require.Error(t, err)
				require.EqualError(t, err, tt.wantErr)
				require.Equal(t, "", gotBase)
			} else {
				require.NoError(t, err)
				require.Equal(t, "somevalue", gotBase)
				require.Equal(t, tt.wantParams, gotParams)
			}
		})
	}
}
