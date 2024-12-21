package mocks

import (
	"net/http"
	"regexp"
	"testing"

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
				header: "X-Test",
				regexp: regexp.MustCompile("^test-.*$"),
			},
		},
		{
			description: "set valid header with value in regexp format MUST be successful",
			def: map[interface{}]interface{}{
				"header": "X-Test",
				"value":  "$matchRegexp(^test-.*$)",
			},
			want: &headerConstraint{
				header: "X-Test",
				regexp: regexp.MustCompile("^test-.*$"),
			},
		},
		{
			description: "missing header key MUST fail",
			def: map[interface{}]interface{}{
				"value": "test-value",
			},
			wantErr: "'header' key required",
		},
		{
			description: "invalid regexp value MUST fail",
			def: map[interface{}]interface{}{
				"header": "X-Test",
				"regexp": "[invalid",
			},
			wantErr: "error parsing regexp: missing closing ]: `[invalid`",
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
			wantErr: "request doesn't have header X-Test",
		},
		{
			description: "header value does not match",
			header:      "X-Test",
			value:       "expected-value",
			request: &http.Request{
				Header: http.Header{"X-Test": []string{"actual-value"}},
			},
			wantErr: "X-Test header value actual-value doesn't match expected expected-value",
		},
		{
			description: "header value matches regexp",
			header:      "X-Test",
			regexp:      "^test-.*$",
			request: &http.Request{
				Header: http.Header{"X-Test": []string{"test-value"}},
			},
			wantErr: "",
		},
		{
			description: "header value does not match regexp",
			header:      "X-Test",
			regexp:      "^test-.*$",
			request: &http.Request{
				Header: http.Header{"X-Test": []string{"wrong-value"}},
			},
			wantErr: "X-Test header value wrong-value doesn't match regexp ^test-.*$",
		},
		{
			description: "header value matches expected value",
			header:      "X-Test",
			value:       "test-value",
			request: &http.Request{
				Header: http.Header{"X-Test": []string{"test-value"}},
			},
			wantErr: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.description, func(t *testing.T) {
			var reCompiled *regexp.Regexp
			if tt.regexp != "" {
				reCompiled = regexp.MustCompile(tt.regexp)
			}

			c := &headerConstraint{
				header: tt.header,
				value:  tt.value,
				regexp: reCompiled,
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
