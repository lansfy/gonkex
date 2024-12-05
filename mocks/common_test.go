package mocks

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_GetRequiredStringKey(t *testing.T) {
	inputMap := map[interface{}]interface{}{
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
			wantErr:      "`absentKey` key required",
		},
		{
			description:  "key exists but value is not a string",
			key:          "nonStringKey",
			allowedEmpty: true,
			wantErr:      "`nonStringKey` must be string",
		},
		{
			description:  "empty value not allowed",
			key:          "emptyKey",
			allowedEmpty: false,
			wantErr:      "`emptyKey` value can't be empty",
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

func Test_GetOptionalStringKey(t *testing.T) {
	inputMap := map[interface{}]interface{}{
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
			wantErr:      "`nonStringKey` must be string",
		},
		{
			description:  "empty value not allowed",
			key:          "emptyKey",
			allowedEmpty: false,
			wantErr:      "`emptyKey` value can't be empty",
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

func Test_GetOptionalIntKey(t *testing.T) {
	inputMap := map[interface{}]interface{}{
		"key":       42,
		"nonIntKey": "aaaa",
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
			description:  "key does not exist, default value returned",
			key:          "absentKey",
			defaultValue: 99,
			want:         99,
		},
		{
			description:  "key exists but value is not an integer",
			key:          "nonIntKey",
			defaultValue: 99,
			want:         0,
			wantErr:      "`nonIntKey` must be integer",
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
