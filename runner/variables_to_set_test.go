package runner

import (
	"testing"

	"github.com/lansfy/gonkex/models"

	"github.com/stretchr/testify/require"
)

func Test_extractVariablesFromJSONResponse(t *testing.T) {
	body := `{"key1": "value1", "key2": "value2"}`
	tests := []struct {
		description string
		varsToSet   map[string]string
		want        map[string]string
		wantErr     string
	}{
		{
			description: "json body with valid paths",
			varsToSet: map[string]string{
				"var1":         "key1",
				"var2":         "key2",
				"wholeBodyVar": "",
			},
			want: map[string]string{
				"var1":         "value1",
				"var2":         "value2",
				"wholeBodyVar": body,
			},
		},
		{
			description: "json body with missing path",
			varsToSet: map[string]string{
				"var1": "key1",
				"var2": "missingKey",
			},
			wantErr: "variable 'var2': path missingKey does not exist in service response",
		},
		{
			description: "json body with valid paths and optional prefix",
			varsToSet: map[string]string{
				"var1":         "body:key1",
				"var2":         "body: key2 ",
				"wholeBodyVar": "body:",
			},
			want: map[string]string{
				"var1":         "value1",
				"var2":         "value2",
				"wholeBodyVar": body,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.description, func(t *testing.T) {
			result := &models.Result{
				ResponseBody:        body,
				ResponseContentType: "json",
			}
			got, err := extractVariablesFromResponse(tt.varsToSet, result)

			if tt.wantErr != "" {
				require.Error(t, err)
				require.EqualError(t, err, tt.wantErr)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.want, got)
			}
		})
	}
}

func Test_extractVariablesFromPlainResponse(t *testing.T) {
	body := "plain text value"
	tests := []struct {
		description string
		varsToSet   map[string]string
		want        map[string]string
		wantErr     string
	}{
		{
			description: "plain text body with valid variable",
			varsToSet: map[string]string{
				"var1": "",
			},
			want: map[string]string{
				"var1": body,
			},
		},
		{
			description: "plain text body with non-empty path",
			varsToSet: map[string]string{
				"var1": "some.path",
			},
			wantErr: "variable 'var1': paths not supported for plain text body",
		},
	}

	for _, tt := range tests {
		t.Run(tt.description, func(t *testing.T) {
			result := &models.Result{
				ResponseBody:        body,
				ResponseContentType: "text",
			}

			got, err := extractVariablesFromResponse(tt.varsToSet, result)

			if tt.wantErr != "" {
				require.Error(t, err)
				require.EqualError(t, err, tt.wantErr)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.want, got)
			}
		})
	}
}

func Test_extractVariablesFromHeaders(t *testing.T) {
	headers := map[string][]string{
		"Test-Header-1": {"aaa", "bbb"},
		"Test-Header-2": {"ccc"},
	}

	tests := []struct {
		description string
		varsToSet   map[string]string
		want        map[string]string
		wantErr     string
	}{
		{
			description: "variables with valid header name",
			varsToSet: map[string]string{
				"var1": "header:Test-Header-1",
				"var2": "header: Test-Header-2 ",
			},
			want: map[string]string{
				"var1": "aaa",
				"var2": "ccc",
			},
		},
		{
			description: "variables with unknown header",
			varsToSet: map[string]string{
				"var1": "header:Test-Header-1",
				"var3": "header:Wrong-Header",
			},
			wantErr: "variable 'var3': response does not include expected header 'Wrong-Header'",
		},
		{
			description: "variables with unknown prefix",
			varsToSet: map[string]string{
				"var1": "header:Test-Header-1",
				"var3": "wrong-prefix:Test-Header-1",
			},
			wantErr: "variable 'var3': unexpected path prefix 'wrong-prefix' (allowed only [body header cookie])",
		},
	}

	for _, tt := range tests {
		t.Run(tt.description, func(t *testing.T) {
			result := &models.Result{
				ResponseHeaders: headers,
			}
			got, err := extractVariablesFromResponse(tt.varsToSet, result)

			if tt.wantErr != "" {
				require.Error(t, err)
				require.EqualError(t, err, tt.wantErr)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.want, got)
			}
		})
	}
}

func Test_extractVariablesFromCookie(t *testing.T) {
	headers := map[string][]string{
		"Set-Cookie": {
			"session_id=abc123; Path=/; HttpOnly",
			"user", // ignored, because has wrong format
			"user=JohnDoe; Path=/; Secure",
		},
	}

	tests := []struct {
		description string
		varsToSet   map[string]string
		want        map[string]string
		wantErr     string
	}{
		{
			description: "variables with valid cookie name",
			varsToSet: map[string]string{
				"var1": "cookie:session_id",
				"var2": "cookie: user ",
			},
			want: map[string]string{
				"var1": "abc123",
				"var2": "JohnDoe",
			},
		},
		{
			description: "variables with unknown cookie",
			varsToSet: map[string]string{
				"var1": "cookie:session_id",
				"var2": "cookie:wrong_name",
			},
			wantErr: "variable 'var2': Set-Cookie header does not include expected cookie 'wrong_name'",
		},
	}

	for _, tt := range tests {
		t.Run(tt.description, func(t *testing.T) {
			result := &models.Result{
				ResponseHeaders: headers,
			}
			got, err := extractVariablesFromResponse(tt.varsToSet, result)

			if tt.wantErr != "" {
				require.Error(t, err)
				require.EqualError(t, err, tt.wantErr)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.want, got)
			}
		})
	}
}
