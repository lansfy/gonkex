package variables

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFromResponse(t *testing.T) {
	tests := []struct {
		description string
		varsToSet   map[string]string
		body        string
		isJSON      bool
		want        map[string]string
		wantErr     string
	}{
		{
			description: "json body with valid paths",
			varsToSet: map[string]string{
				"var1": "key1",
				"var2": "key2",
			},
			body:   `{"key1": "value1", "key2": "value2"}`,
			isJSON: true,
			want: map[string]string{
				"var1": "value1",
				"var2": "value2",
			},
		},
		{
			description: "json body with missing path",
			varsToSet: map[string]string{
				"var1": "key1",
				"var2": "missingKey",
			},
			body:    `{"key1": "value1"}`,
			isJSON:  true,
			wantErr: "path 'missingKey' does not exist in given json",
		},
		{
			description: "plain text body with valid variable",
			varsToSet: map[string]string{
				"var1": "unusedPath",
			},
			body:   "plain text value",
			isJSON: false,
			want: map[string]string{
				"var1": "plain text value",
			},
		},
		{
			description: "plain text body with multiple variables",
			varsToSet: map[string]string{
				"var1": "unusedPath",
				"var2": "unusedPath2",
			},
			body:    "plain text value",
			isJSON:  false,
			wantErr: "count of variables for plain-text response should be 1, 2 given",
		},
	}

	for _, tt := range tests {
		t.Run(tt.description, func(t *testing.T) {
			got, err := FromResponse(tt.varsToSet, tt.body, tt.isJSON)

			if tt.wantErr != "" {
				require.Error(t, err)
				require.EqualError(t, err, tt.wantErr)
			} else {
				require.NoError(t, err)
				wantVars := New()
				wantVars.Load(tt.want)
				require.Equal(t, wantVars, got)
			}
		})
	}
}
