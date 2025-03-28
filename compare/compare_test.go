package compare

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func makeErrorString(path, msg string, expected, actual interface{}) string {
	return fmt.Sprintf(
		"path '%s': %s:\n     expected: %v\n       actual: %v",
		path,
		msg,
		expected,
		actual,
	)
}

func makeDiffErrorString(path, diff string) string {
	return fmt.Sprintf(
		"path '%s': values do not match:\n     diff (--- expected vs +++ actual):\n%s\n",
		path,
		diff,
	)
}

func TestCompareScalarTypes(t *testing.T) {
	tests := []struct {
		name     string
		expected interface{}
		actual   interface{}
		wantErr  string
	}{
		{
			name:     "nil values MUST be equal",
			expected: nil,
			actual:   nil,
		},
		{
			name:     "nil and not-nil values MUST be not equal",
			expected: "",
			actual:   nil,
			wantErr:  makeErrorString("$", "types do not match", "string", "nil"),
		},
		{
			name:     "nil and not-nil values MUST be not equal",
			expected: nil,
			actual:   "",
			wantErr:  makeErrorString("$", "types do not match", "nil", "string"),
		},
		{
			name:     "same string value MUST be equal",
			expected: "12345",
			actual:   "12345",
		},
		{
			name:     "different strings MUST produce error",
			expected: "123",
			actual:   "12345",
			wantErr:  makeErrorString("$", "values do not match", "123", "12345"),
		},
		{
			name:     "different multi-line strings MUST produce diff as error",
			expected: "123\n12345",
			actual:   "12345",
			wantErr:  makeDiffErrorString("$", "-123\n 12345"),
		},
		{
			name:     "different multi-line strings MUST produce diff as error",
			expected: "12345",
			actual:   "12345\n123",
			wantErr:  makeDiffErrorString("$", " 12345\n+123"),
		},
		{
			name:     "same integer value MUST be equal",
			expected: 12345,
			actual:   12345,
		},
		{
			name:     "different integers MUST produce error",
			expected: 123,
			actual:   12345,
			wantErr:  makeErrorString("$", "values do not match", 123, 12345),
		},
		{
			name:     "WHEN the string matches a regular expression, the check MUST pass",
			expected: "$matchRegexp(x.+z)",
			actual:   "xyyyz",
		},
		{
			name:     "WHEN the string not matches a regular expression, the check MUST fail",
			expected: "$matchRegexp(x.+z)",
			actual:   "ayyyb",
			wantErr:  makeErrorString("$", "value does not match regexp", "$matchRegexp(x.+z)", "ayyyb"),
		},
		{
			name:     "WHEN the integer number matches a regular expression, the check MUST pass",
			expected: "$matchRegexp(^[0-5]+$)",
			actual:   543210,
		},
		{
			name:     "WHEN the integer number not matches a regular expression, the check MUST fail",
			expected: "$matchRegexp(^[0-5]+$)",
			actual:   12367,
			wantErr:  makeErrorString("$", "value does not match regexp", "$matchRegexp(^[0-5]+$)", "12367"),
		},
		{
			name:     "WHEN the float number matches a regular expression, the check MUST pass",
			expected: "$matchRegexp(^[0-9]+\\.2.*$)",
			actual:   1.234,
		},
		{
			name:     "WHEN the integer number not matches a regular expression, the check MUST fail",
			expected: "$matchRegexp(^[0-9]+\\.3.*$)",
			actual:   1.23,
			wantErr:  makeErrorString("$", "value does not match regexp", "$matchRegexp(^[0-9]+\\.3.*$)", "1.23"),
		},
		{
			name:     "WHEN condition has invalid regular expression, the check MUST fail with error",
			expected: "$matchRegexp((?x))",
			actual:   2,
			wantErr:  makeErrorString("$", "cannot compile regexp", nil, "invalid or unsupported Perl syntax: `(?x`"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errors := Compare(tt.expected, tt.actual, Params{})
			if tt.wantErr == "" {
				require.Empty(t, errors)
			} else {
				require.Len(t, errors, 1)
				require.Equal(t, tt.wantErr, errors[0].Error())
			}
		})
	}
}

//go:embed testdata/complex_data_1.json
var complexJson1 string

//go:embed testdata/complex_data_2.json
var complexJson2 string

func Test_compareJson(t *testing.T) {
	tests := []struct {
		name     string
		expected string
		actual   string
		params   Params
		wantErr  string
	}{
		{
			name:     "equal array MUST be equal",
			expected: `["1", "2"]`,
			actual:   `["1", "2"]`,
		},
		{
			name:     "empty arrays MUST be equal",
			expected: `[]`,
			actual:   `[]`,
		},
		{
			name:     "array of string with same element, but different order MUST be equal IF compare with IgnoreArraysOrdering",
			expected: `["1", "2"]`,
			actual:   `["2", "1"]`,
			params: Params{
				IgnoreArraysOrdering: true,
			},
		},
		{
			name:     "arrays with different length MUST produce error with length",
			expected: `["1", "2", "3"]`,
			actual:   `["1", "2"]`,
			wantErr:  makeErrorString("$", "array lengths do not match", 3, 2),
		},
		{
			name:     "arrays with one different element MUST produce error with this element",
			expected: `["1", "2"]`,
			actual:   `["1", "3"]`,
			wantErr:  makeErrorString("$[1]", "values do not match", 2, 3),
		},
		{
			name:     "equal nested arrays MUST be equal",
			expected: `[["1", "2"], ["3", "4"]]`,
			actual:   `[["1", "2"], ["3", "4"]]`,
		},
		{
			name:     "nested arrays with one different element MUST produce error with this element",
			expected: `[["1", "2"], ["3", "4"]]`,
			actual:   `[["1", "2"], ["3", "5"]]`,
			wantErr:  makeErrorString("$[1][1]", "values do not match", 4, 5),
		},
		{
			name:     "arrays MUST support comparison elements with regexp",
			expected: `["2", "$matchRegexp(^x.+z$)"]`,
			actual:   `["2", "xyyyz"]`,
		},
		{
			name:     "arrays MUST support comparison elements with regexp for integer value",
			expected: `["2", "$matchRegexp(^[0-9]+$)"]`,
			actual:   `["2", 123]`,
		},
		{
			name:     "WHEN one of element in array doesn't match regular expression, the check MUST fail",
			expected: `["2", "$matchRegexp(x.+z)"]`,
			actual:   `["2", "ayyyb"]`,
			wantErr:  makeErrorString("$[1]", "value does not match regexp", "$matchRegexp(x.+z)", "ayyyb"),
		},
		{
			name:     "equal maps MUST be equal",
			expected: `{"a": "1", "b": "2"}`,
			actual:   `{"b": "2", "a": "1"}`,
		},
		{
			name:     "maps MUST support comparison elements with regexp",
			expected: `{"a": "1", "b": "$matchRegexp(^x.+z$)"}`,
			actual:   `{"a": "1", "b": "xyyyz"}`,
		},
		{
			name:     "WHEN one of element in map doesn't match regular expression, the check MUST fail",
			expected: `{"a": "1", "b": "$matchRegexp(x.+z)"}`,
			actual:   `{"a": "1", "b": "ayyyb"}`,
			wantErr:  makeErrorString("$.b", "value does not match regexp", "$matchRegexp(x.+z)", "ayyyb"),
		},
		{
			name:     "maps with extra item MUST be equal",
			expected: `{"a": "1", "b": "2"}`,
			actual:   `{"b": "2", "a": "1", "c": "3"}`,
		},
		{
			name:     "WHEN actual map has extra fields AND DisallowExtraFields enabled, the check MUST fail",
			expected: `{"a": "1", "b": "2"}`,
			actual:   `{"b": "2", "a": "1", "c": "3"}`,
			params:   Params{DisallowExtraFields: true},
			wantErr:  makeErrorString("$", "map lengths do not match", 2, 3),
		},
		{
			name:     "WHEN actual map has unexpected field, the check MUST fail",
			expected: `{"a": "1", "b": "2"}`,
			actual:   `{"a": "1", "c": "2"}`,
			wantErr:  makeErrorString("$", "key is missing", "b", "<missing>"),
		},
		{
			name:     "WHEN actual map has field with different value, the check MUST fail",
			expected: `{"a": "1", "b": "2"}`,
			actual:   `{"a": "1", "b": "3"}`,
			wantErr:  makeErrorString("$.b", "values do not match", 2, 3),
		},
		{
			name:     "compare of two equal maps MUST be success",
			expected: `{"a": {"i": "3", "j": "4"}, "b": {"k": "5", "l": "6"}}`,
			actual:   `{"a": {"i": "3", "j": "4"}, "b": {"k": "5", "l": "6"}}`,
		},
		{
			name:     "WHEN actual map doesn't have key, the check MUST fail",
			expected: `{"a": {"i": "3", "j": "4"}, "b": {"k": "5", "l": "6"}}`,
			actual:   `{"a": {"i": "3", "j": "4"}, "b": {"l": "6"}}`,
			wantErr:  makeErrorString("$.b", "key is missing", "k", "<missing>"),
		},
		{
			name:     "WHEN actual map has key with different value, the check MUST fail",
			expected: `{"a": {"i": "3", "j": "4"}, "b": {"k": "5", "l": "6"}}`,
			actual:   `{"a": {"i": "3", "j": "4"}, "b": {"k": "5", "l": "7"}}`,
			wantErr:  makeErrorString("$.b.l", "values do not match", 6, 7),
		},
		{
			name:     "equal scalars MUST be equal",
			expected: `1`,
			actual:   `1`,
		},
		{
			name:     "different scalars MUST produce error",
			expected: `1`,
			actual:   `2`,
			wantErr:  makeErrorString("$", "values do not match", 1, 2),
		},
		{
			name: "TestCompareEqualArraysWithIgnoreArraysOrdering",
			expected: `{
				"data": [
					{"name": "n111"},
					{"name": "n222"},
					{"name": "n333"}
				]}`,
			actual: `{
				"data": [
					{"message": "m555", "name": "n333"},
					{"message": "m777", "name": "n111"},
					{"message": "m999","name": "n222"}
				]}`,
			params: Params{IgnoreArraysOrdering: true},
		},
		{
			name:     "test success complex json comparison",
			expected: complexJson1,
			actual:   complexJson1,
		},
		{
			name:     "test failed complex json comparison",
			expected: complexJson1,
			actual:   complexJson2,
			wantErr: makeErrorString(
				"$.paths./api/get-delivery-info.get.parameters[2].$ref",
				"values do not match",
				"#/parameters/profile_id",
				"#/parameters/profile_id2",
			),
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
			wantErr:  makeErrorString("$", "unknown $matchArray mode", "pattern / pattern+subset / subset+pattern", "errorhere"),
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
			wantErr:  "path '$': array with $matchArray(pattern) must pattern element",
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

func TestCompareArraysFewErrors(t *testing.T) {
	array1 := []string{"1", "2", "3"}
	array2 := []string{"1", "3", "4"}
	errors := Compare(array1, array2, Params{})
	assert.Len(t, errors, 2)
}

func TestCompareMapsWithFewErrors(t *testing.T) {
	array1 := map[string]string{"a": "1", "b": "2", "c": "5"}
	array2 := map[string]string{"a": "1", "b": "3", "d": "4"}
	errors := Compare(array1, array2, Params{})
	assert.Len(t, errors, 2)
}
