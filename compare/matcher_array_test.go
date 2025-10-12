package compare

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_matchArray(t *testing.T) {
	tests := []struct {
		name     string
		expected string
		actual   string
		params   Params
		wantErr  string
	}{
		{
			name:     "$matchArray(pattern) works",
			expected: `["$matchArray(pattern)", "$matchRegexp(^[0-9]+$)"]`,
			actual:   `["123", "456", "7", "8", "9"]`,
		},
		{
			name:     "$matchArray(pattern) works on empty array",
			expected: `["$matchArray(pattern)", "$matchRegexp(^[0-9]+$)"]`,
			actual:   `[]`,
		},
		{
			name:     "$matchArray(subset+pattern) works",
			expected: `["$matchArray(subset+pattern)", "a", "b", "c", "$matchRegexp(^[0-9]+$)"]`,
			actual:   `["a", "b", "c", "5", "6"]`,
		},
		{
			name:     "$matchArray(subset+pattern) works WHEN no extra elements",
			expected: `["$matchArray(subset+pattern)", "a", "b", "c", "$matchRegexp(^[0-9]+$)"]`,
			actual:   `["a", "b", "c"]`,
		},
		{
			name:     "$matchArray(pattern+subset) works",
			expected: `["$matchArray(pattern+subset)", "$matchRegexp(^[0-9]+$)", "a", "b", "c"]`,
			actual:   `["5", "6", "a", "b", "c"]`,
		},
		{
			name:     "$matchArray(pattern+subset) works WHEN no extra elements",
			expected: `["$matchArray(pattern+subset)", "$matchRegexp(^[0-9]+$)", "a", "b", "c"]`,
			actual:   `["a", "b", "c"]`,
		},
		{
			name:     "WHEN $matchArray has unknown mode MUST fail with error",
			expected: `["$matchArray(errorhere)", ["$matchRegexp(^[0-4]+$)", "a"]]`,
			actual:   `[]`,
			wantErr:  makeErrorString("$", "parse '$matchArray': unknown mode", "pattern / pattern+subset / subset+pattern", "errorhere"),
		},
		{
			name:     "WHEN first element in array is $matchArray(pattern) next element MUST treat as template for all elements in this array",
			expected: `["$matchArray(pattern)", ["$matchRegexp(^[0-4]+$)", "a"]]`,
			actual:   `[["12", "a"], ["34", "a"], ["03", "a"]]`,
		},
		{
			name:     "WHEN use $matchArray(pattern) and one element of array does not match the pattern, the check MUST fail",
			expected: `["$matchArray(pattern)", ["$matchRegexp(^[0-4]+$)", "a"]]`,
			actual:   `[["12", "a"], ["34", "a"], ["45", "a"]]`,
			wantErr:  makeErrorString("$[2][0]", "value does not match regexp", "$matchRegexp(^[0-4]+$)", "45"),
		},
		{
			name:     "WHEN use $matchArray(subset+pattern) and one element of array does not match the pattern, the check MUST fail",
			expected: `["$matchArray(subset+pattern)", "a", "b", "c", "$matchRegexp(^[0-9]+$)"]`,
			actual:   `["a", "b", "c", "d", "5"]`,
			wantErr:  makeErrorString("$[3]", "value does not match regexp", "$matchRegexp(^[0-9]+$)", "d"),
		},
		{
			name:     "WHEN use $matchArray(subset+pattern) and header of array does not match the subset, the check MUST fail",
			expected: `["$matchArray(subset+pattern)", "a", "b", "b", "$matchRegexp(^[0-9]+$)"]`,
			actual:   `["a", "b", "c", "5", "6"]`,
			wantErr:  makeErrorString("$[2]", "values do not match", "b", "c"),
		},
		{
			name:     "WHEN use $matchArray(pattern+subset) and one element of array does not match the pattern, the check MUST fail",
			expected: `["$matchArray(pattern+subset)", "$matchRegexp(^[0-9]+$)", "a", "b", "c"]`,
			actual:   `["d", "5", "a", "b", "c"]`,
			wantErr:  makeErrorString("$[0]", "value does not match regexp", "$matchRegexp(^[0-9]+$)", "d"),
		},
		{
			name:     "WHEN use $matchArray(pattern+subset) and footer of array does not match the subset, the check MUST fail",
			expected: `["$matchArray(pattern+subset)", "$matchRegexp(^[0-9]+$)", "b", "b", "c"]`,
			actual:   `["5", "6", "a", "b", "c"]`,
			wantErr:  makeErrorString("$[2]", "values do not match", "b", "a"),
		},
		{
			name:     "WHEN use $matchArray(pattern) and didn't provide pattern, the check MUST fail",
			expected: `["$matchArray(pattern)"]`,
			actual:   `[["12", "a"], ["34", "a"]]`,
			wantErr:  "path '$': array with $matchArray(pattern) must have one pattern element",
		},
		{
			name:     "WHEN use $matchArray(pattern+subset) and didn't provide pattern or subset, the check MUST fail",
			expected: `["$matchArray(pattern+subset)", "aaaa"]`,
			actual:   `[]`,
			wantErr:  "path '$': array with $matchArray(pattern+subset) must have pattern and additional elements",
		},
		{
			name:     "WHEN use $matchArray(subset+pattern) and didn't provide pattern or subset, the check MUST fail",
			expected: `["$matchArray(subset+pattern)", "aaaa"]`,
			actual:   `[]`,
			wantErr:  "path '$': array with $matchArray(subset+pattern) must have pattern and additional elements",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var json1, json2 interface{}
			err := json.Unmarshal([]byte(tt.expected), &json1)
			require.NoError(t, err)
			err = json.Unmarshal([]byte(tt.actual), &json2)
			require.NoError(t, err)

			errors := Compare(json1, json2, tt.params)
			if tt.wantErr == "" {
				require.Empty(t, errors)
			} else {
				require.Len(t, errors, 1)
				require.Equal(t, tt.wantErr, errors[0].Error())
			}
		})
	}
}
