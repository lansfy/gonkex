package mocks

import (
	"net/http"
	"regexp"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_newPathConstraint(t *testing.T) {
	tests := []struct {
		description string
		def         map[interface{}]interface{}
		want        verifier
		wantErr     string
	}{
		{
			description: "set valid path MUST be successful",
			def: map[interface{}]interface{}{
				"path": "/test",
			},
			want: &pathConstraint{
				path: "/test",
			},
		},
		{
			description: "set valid regexp MUST be successful",
			def: map[interface{}]interface{}{
				"regexp": "^/test[0-9]*$",
			},
			want: &pathConstraint{
				regexp: regexp.MustCompile("^/test[0-9]*$"),
			},
		},
		{
			description: "set valid regexp via path field MUST be successful",
			def: map[interface{}]interface{}{
				"path": "$matchRegexp(^/test[0-9]*$)",
			},
			want: &pathConstraint{
				regexp: regexp.MustCompile("^/test[0-9]*$"),
			},
		},
		{
			description: "set empty path and empty regexp MUST be successful",
			def: map[interface{}]interface{}{
				"path": "",
			},
			want: &pathConstraint{
				path:   "",
				regexp: nil,
			},
		},
		{
			description: "set path with wrong type MUST fail",
			def: map[interface{}]interface{}{
				"path": 42,
			},
			wantErr: "key 'path' has non-string value",
		},
		{
			description: "set empty regexp MUST fail",
			def: map[interface{}]interface{}{
				"regexp": "",
			},
			wantErr: "'regexp' value can't be empty",
		},
		{
			description: "set invalid regexp MUST fail",
			def: map[interface{}]interface{}{
				"regexp": "[",
			},
			wantErr: "error parsing regexp: missing closing ]: `[`",
		},
	}

	for _, tt := range tests {
		t.Run(tt.description, func(t *testing.T) {
			got, err := loadPathConstraint(tt.def)
			if tt.wantErr != "" {
				require.Error(t, err)
				require.Equal(t, err.Error(), tt.wantErr)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.want, got)
			}
		})
	}
}

func Test_pathConstraint_GetName(t *testing.T) {
	c := &pathConstraint{}
	got := c.GetName()
	require.Equal(t, "pathMatches", got)
}

func Test_pathConstraint_Verify(t *testing.T) {
	tests := []struct {
		description string
		path        string
		re          string
		reqPath     string
		wantErr     string
	}{
		{
			description: "path matches exactly",
			path:        "/test",
			reqPath:     "/test",
			wantErr:     "",
		},
		{
			description: "path does not match",
			path:        "/test",
			reqPath:     "/mismatch",
			wantErr:     "url 'path' does not match expected:\n     expected: /test\n       actual: /mismatch",
		},
		{
			description: "regexp matches",
			re:          "^/test[0-9]*$",
			reqPath:     "/test123",
			wantErr:     "",
		},
		{
			description: "regexp does not match",
			re:          "^/test[0-9]*$",
			reqPath:     "/mismatch",
			wantErr:     "url 'path' does not match expected regexp:\n     expected: ^/test[0-9]*$\n       actual: /mismatch",
		},
	}

	for _, tt := range tests {
		t.Run(tt.description, func(t *testing.T) {
			checker, err := newPathConstraint(tt.path, tt.re)
			require.NoError(t, err)

			req, err := http.NewRequest(http.MethodGet, tt.reqPath, nil)
			require.NoError(t, err)

			got := checker.Verify(req)
			if tt.wantErr == "" {
				require.Nil(t, got)
			} else {
				require.Len(t, got, 1)
				require.EqualError(t, got[0], tt.wantErr)
			}
		})
	}
}
