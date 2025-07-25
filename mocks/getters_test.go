package mocks

import (
	"testing"

	"github.com/lansfy/gonkex/compare"

	"github.com/stretchr/testify/require"
)

func Test_getRequiredStringKey(t *testing.T) {
	inputMap := map[string]interface{}{
		"key":          "value",
		"nonStringKey": 123,
		"emptyKey":     "",
	}
	tests := []struct {
		description  string
		key          string
		allowedEmpty bool
		want         string
		wantErr      string
	}{
		{
			description:  "key exists and value is valid string",
			key:          "key",
			allowedEmpty: true,
			want:         "value",
		},
		{
			description:  "empty value allowed",
			key:          "emptyKey",
			allowedEmpty: true,
			want:         "",
		},
		{
			description:  "key does not exist",
			key:          "absentKey",
			allowedEmpty: true,
			wantErr:      "'absentKey' key required",
		},
		{
			description:  "key exists but value is not a string",
			key:          "nonStringKey",
			allowedEmpty: true,
			wantErr:      "key 'nonStringKey' has non-string value",
		},
		{
			description:  "empty value not allowed",
			key:          "emptyKey",
			allowedEmpty: false,
			wantErr:      "'emptyKey' value can't be empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.description, func(t *testing.T) {
			got, err := getRequiredStringKey(inputMap, tt.key, tt.allowedEmpty)
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

func Test_getOptionalStringKey(t *testing.T) {
	inputMap := map[string]interface{}{
		"key":          "value",
		"nonStringKey": 123,
		"emptyKey":     "",
	}
	tests := []struct {
		description  string
		key          string
		allowedEmpty bool
		want         string
		wantErr      string
	}{
		{
			description:  "key exists and value is valid string",
			key:          "key",
			allowedEmpty: true,
			want:         "value",
		},
		{
			description:  "key does not exist",
			key:          "absentKey",
			allowedEmpty: true,
			want:         "",
		},
		{
			description:  "empty value allowed",
			key:          "emptyKey",
			allowedEmpty: true,
			want:         "",
		},
		{
			description:  "key exists but value is not a string",
			key:          "nonStringKey",
			allowedEmpty: true,
			wantErr:      "key 'nonStringKey' has non-string value",
		},
		{
			description:  "empty value not allowed",
			key:          "emptyKey",
			allowedEmpty: false,
			wantErr:      "'emptyKey' value can't be empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.description, func(t *testing.T) {
			got, err := getOptionalStringKey(inputMap, tt.key, tt.allowedEmpty)
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

func Test_getOptionalIntKey(t *testing.T) {
	inputMap := map[string]interface{}{
		"key":         42,
		"stringKey":   "84",
		"nonIntKey":   "aaaa",
		"negativeKey": "-100",
		"nilKey":      nil,
	}

	tests := []struct {
		description  string
		key          string
		defaultValue int
		want         int
		wantErr      string
	}{
		{
			description:  "key exists and value is valid integer",
			key:          "key",
			defaultValue: 0,
			want:         42,
		},
		{
			description:  "key exists and value is string with valid integer",
			key:          "stringKey",
			defaultValue: 0,
			want:         84,
		},
		{
			description:  "key does not exist, default value returned",
			key:          "absentKey",
			defaultValue: 99,
			want:         99,
		},
		{
			description:  "key exists but value can't be converted to integer",
			key:          "nonIntKey",
			defaultValue: 99,
			want:         0,
			wantErr:      "value for key 'nonIntKey' cannot be converted to integer",
		},
		{
			description:  "key exists but value is a negative integer",
			key:          "negativeKey",
			defaultValue: 99,
			want:         0,
			wantErr:      "value for the key 'negativeKey' cannot be negative",
		},
		{
			description:  "key exists but value is a negative integer",
			key:          "nilKey",
			defaultValue: 99,
			want:         0,
			wantErr:      "value for key 'nilKey' cannot be converted to integer",
		},
	}

	for _, tt := range tests {
		t.Run(tt.description, func(t *testing.T) {
			got, err := getOptionalIntKey(inputMap, tt.key, tt.defaultValue)
			if tt.wantErr != "" {
				require.Error(t, err)
				require.Equal(t, tt.wantErr, err.Error())
				require.Zero(t, got)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.want, got)
			}
		})
	}
}

func Test_readCompareParams(t *testing.T) {
	tests := []struct {
		description string
		input       map[string]interface{}
		want        *compare.Params
		wantErr     string
	}{
		{
			description: "default params when no 'comparisonParams' key",
			input: map[string]interface{}{
				"someOtherKey": true,
			},
			want: &compare.Params{
				IgnoreArraysOrdering: true,
			},
		},
		{
			description: "valid 'comparisonParams' values (1)",
			input: map[string]interface{}{
				"comparisonParams": map[string]interface{}{
					"ignoreValues":         true,
					"ignoreArraysOrdering": false,
					"disallowExtraFields":  true,
				},
			},
			want: &compare.Params{
				IgnoreValues:         true,
				IgnoreArraysOrdering: false,
				DisallowExtraFields:  true,
			},
		},
		{
			description: "valid 'comparisonParams' values (2)",
			input: map[string]interface{}{
				"comparisonParams": map[string]interface{}{
					"ignoreValues":         false,
					"ignoreArraysOrdering": true,
					"disallowExtraFields":  false,
				},
			},
			want: &compare.Params{
				IgnoreValues:         false,
				IgnoreArraysOrdering: true,
				DisallowExtraFields:  false,
			},
		},
		{
			description: "valid 'comparisonParams' values (3)",
			input: map[string]interface{}{
				"comparisonParams": map[string]interface{}{
					"ignoreValues":         true,
					"ignoreArraysOrdering": true,
					"disallowExtraFields":  false,
				},
			},
			want: &compare.Params{
				IgnoreValues:         true,
				IgnoreArraysOrdering: true,
				DisallowExtraFields:  false,
			},
		},
		{
			description: "non-map 'comparisonParams' value",
			input: map[string]interface{}{
				"comparisonParams": "invalidType",
			},
			wantErr: "section 'comparisonParams': must be a map",
		},
		{
			description: "non-bool value in 'comparisonParams'",
			input: map[string]interface{}{
				"comparisonParams": map[string]interface{}{
					"ignoreValues": "notBool",
				},
			},
			wantErr: "section 'comparisonParams': key 'ignoreValues' has non-bool value",
		},
		{
			description: "unexpected key in 'comparisonParams'",
			input: map[string]interface{}{
				"comparisonParams": map[string]interface{}{
					"someKey": true,
				},
			},
			wantErr: "section 'comparisonParams': unexpected key 'someKey' (allowed only [ignoreValues ignoreArraysOrdering disallowExtraFields])",
		},
	}

	for _, tt := range tests {
		t.Run(tt.description, func(t *testing.T) {
			actual, err := readCompareParams(tt.input)
			if tt.wantErr != "" {
				require.Error(t, err)
				require.EqualError(t, err, tt.wantErr)
			} else {
				require.NoError(t, err)
				require.Equal(t, *tt.want, actual)
			}
		})
	}
}

func Test_loadHeaders(t *testing.T) {
	tests := []struct {
		description string
		input       map[string]interface{}
		want        map[string]string
		wantErr     string
	}{
		{
			description: "valid headers",
			input: map[string]interface{}{
				"headers": map[string]interface{}{
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
			input: map[string]interface{}{
				"headers": "invalid",
			},
			want:    nil,
			wantErr: "map under 'headers' key is required",
		},
		{
			description: "header value is not a string",
			input: map[string]interface{}{
				"headers": map[string]interface{}{
					"key": 123,
				},
			},
			want:    nil,
			wantErr: "'headers' requires string values",
		},
		{
			description: "no headers key",
			input:       map[string]interface{}{},
			want:        nil,
			wantErr:     "",
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			result, err := loadHeaders(test.input)

			if test.wantErr != "" {
				require.Error(t, err)
				require.EqualError(t, err, test.wantErr)
			} else {
				require.NoError(t, err)
			}

			require.Equal(t, test.want, result)
		})
	}
}
