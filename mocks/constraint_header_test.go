package mocks

import (
	"net/http"
	"testing"

	"github.com/lansfy/gonkex/compare"

	"github.com/stretchr/testify/require"
)

func Test_loadHeaderConstraint(t *testing.T) {
	tests := []struct {
		description string
		def         map[interface{}]interface{}
		want        verifier
		wantErr     string
	}{
		{
			description: "set valid header with value MUST be successful",
			def: map[interface{}]interface{}{
				"header": "X-Test",
				"value":  "test-value",
			},
			want: &headerConstraint{
				header: "X-Test",
				value:  "test-value",
			},
		},
		{
			description: "set valid header with regexp value MUST be successful",
			def: map[interface{}]interface{}{
				"header": "X-Test",
				"regexp": "^test-.*$",
			},
			want: &headerConstraint{
				header:  "X-Test",
				matcher: compare.StringAsMatcher("$matchRegexp(^test-.*$)"),
			},
		},
		{
			description: "set valid header with value in regexp format MUST be successful",
			def: map[interface{}]interface{}{
				"header": "X-Test",
				"value":  "$matchRegexp(^test-.*$)",
			},
			want: &headerConstraint{
				header:  "X-Test",
				matcher: compare.StringAsMatcher("$matchRegexp(^test-.*$)"),
			},
		},
		{
			description: "missing header key MUST fail",
			def: map[interface{}]interface{}{
				"value": "test-value",
			},
			wantErr: "'header' key required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.description, func(t *testing.T) {
			got, err := loadHeaderConstraint(tt.def)
			if tt.wantErr == "" {
				require.NoError(t, err)
				require.Equal(t, tt.want, got)
			} else {
				require.Error(t, err)
				require.EqualError(t, err, tt.wantErr)
				require.Nil(t, got)
			}
		})
	}
}

func Test_headerConstraint_GetName(t *testing.T) {
	c := &headerConstraint{}
	require.Equal(t, "headerIs", c.GetName())
}

func Test_headerConstraint_Verify(t *testing.T) {
	tests := []struct {
		description string
		header      string
		value       string
		regexp      string
		request     *http.Request
		wantErr     string
	}{
		{
			description: "header missing",
			header:      "X-Test",
			request: &http.Request{
				Header: http.Header{},
			},
			wantErr: "request does not have header 'X-Test'",
		},
		{
			description: "header value does not match",
			header:      "X-Test",
			value:       "expected-value",
			request: &http.Request{
				Header: http.Header{"X-Test": []string{"actual-value"}},
			},
			wantErr: "'X-Test' header value does not match:\n     expected: expected-value\n       actual: actual-value",
		},
		{
			description: "header value matches regexp",
			header:      "X-Test",
			regexp:      "^test-.*$",
			request: &http.Request{
				Header: http.Header{"X-Test": []string{"test-value"}},
			},
		},
		{
			description: "header value does not match regexp",
			header:      "X-Test",
			regexp:      "^test-.*$",
			request: &http.Request{
				Header: http.Header{"X-Test": []string{"wrong-value"}},
			},
			wantErr: "'X-Test' header: value does not match regexp:\n     expected: $matchRegexp(^test-.*$)\n       actual: wrong-value",
		},
		{
			description: "header value matches expected value",
			header:      "X-Test",
			value:       "test-value",
			request: &http.Request{
				Header: http.Header{"X-Test": []string{"test-value"}},
			},
		},
		{
			description: "invalid regexp value MUST fail",
			header:      "X-Test",
			regexp:      "[invalid",
			request: &http.Request{
				Header: http.Header{"X-Test": []string{"test-value"}},
			},
			wantErr: "'X-Test' header: cannot compile regexp:\n     expected: <nil>\n       actual: missing closing ]: `[invalid`",
		},
	}

	for _, tt := range tests {
		t.Run(tt.description, func(t *testing.T) {
			var matcher compare.Matcher
			if tt.regexp != "" {
				matcher = compare.StringAsMatcher(compare.MatchRegexpWrap(tt.regexp))
			}

			c := &headerConstraint{
				header:  tt.header,
				value:   tt.value,
				matcher: matcher,
			}

			got := c.Verify(tt.request)
			if tt.wantErr == "" {
				require.Nil(t, got)
			} else {
				require.NotNil(t, got)
				require.EqualError(t, got[0], tt.wantErr)
			}
		})
	}
}
