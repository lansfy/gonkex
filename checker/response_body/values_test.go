package response_body

import (
	"testing"

	"github.com/lansfy/gonkex/models"

	"github.com/stretchr/testify/require"
)

func Test_ExtractValuesFromJSONResponse(t *testing.T) {
	defaultBody := `{"key1": "value1", "key2": "value2"}`
	tests := []struct {
		description string
		varsToSet   map[string]string
		body        string
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
			body: defaultBody,
			want: map[string]string{
				"var1":         "value1",
				"var2":         "value2",
				"wholeBodyVar": defaultBody,
			},
		},
		{
			description: "json body with missing path",
			varsToSet: map[string]string{
				"var1": "key1",
				"var2": "missingKey",
			},
			body:    defaultBody,
			wantErr: "variable 'var2': path '$.missingKey' does not exist in service response",
		},
		{
			description: "empty json body with path",
			varsToSet: map[string]string{
				"var1": "key1",
			},
			body:    "",
			wantErr: "variable 'var1': paths not supported for empty body",
		},
		{
			description: "json body with valid paths and optional prefix",
			varsToSet: map[string]string{
				"var1":         "body:key1",
				"var2":         "body: key2 ",
				"wholeBodyVar": "body:",
			},
			body: defaultBody,
			want: map[string]string{
				"var1":         "value1",
				"var2":         "value2",
				"wholeBodyVar": defaultBody,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.description, func(t *testing.T) {
			result := &models.Result{
				ResponseBody:        tt.body,
				ResponseContentType: "json",
			}
			got, err := ExtractValues(tt.varsToSet, result)

			if tt.wantErr != "" {
				require.Equal(t, 1, len(err))
				require.Error(t, err[0])
				require.EqualError(t, err[0], tt.wantErr)
			} else {
				require.Equal(t, 0, len(err), "not exected error: %v", err)
				require.Equal(t, tt.want, got)
			}
		})
	}
}

func Test_ExtractValuesFromXMLResponse(t *testing.T) {
	defaultBody := `<?xml version="1.0" encoding="UTF-8"?>
		<Items>
			<Item>
				<Name>name1</Name>
				<Value>value1</Value>
			</Item>
			<Item>
				<Name>name2</Name>
				<Value>value2</Value>
			</Item>
		</Items>`
	tests := []struct {
		description string
		varsToSet   map[string]string
		body        string
		want        map[string]string
		wantErr     string
	}{
		{
			description: "xml body with valid paths",
			varsToSet: map[string]string{
				"var1":         "Items.Item.#(Value==\"value1\").Name",
				"var2":         "Items.Item.#(Name==\"name2\").Value",
				"wholeBodyVar": "",
			},
			body: defaultBody,
			want: map[string]string{
				"var1":         "name1",
				"var2":         "value2",
				"wholeBodyVar": defaultBody,
			},
		},
		{
			description: "xml body with missing path",
			varsToSet: map[string]string{
				"var1": "Items.Item.#(Value==\"value1\").Name",
				"var2": "missingKey",
			},
			body:    defaultBody,
			wantErr: "variable 'var2': path '$.missingKey' does not exist in service response",
		},
		{
			description: "empty xml body with path",
			varsToSet: map[string]string{
				"var1": "key1",
			},
			body:    "",
			wantErr: "variable 'var1': paths not supported for empty body",
		},
		{
			description: "xml body with valid paths and optional prefix",
			varsToSet: map[string]string{
				"var1":         "body:Items.Item.0.Name",
				"var2":         "body: Items.Item.1.Value ",
				"wholeBodyVar": "body:",
			},
			body: defaultBody,
			want: map[string]string{
				"var1":         "name1",
				"var2":         "value2",
				"wholeBodyVar": defaultBody,
			},
		},
		{
			description: "invalid xml body",
			varsToSet: map[string]string{
				"var1":         "body:Items.Item.0.Name",
				"wholeBodyVar": "body:",
			},
			body: "<Items>",
			want: map[string]string{
				"var1":         "name1",
				"var2":         "value2",
				"wholeBodyVar": defaultBody,
			},
			wantErr: "variable 'var1': invalid XML in response: XML syntax error on line 1: unexpected EOF",
		},
		{
			description: "invalid xml body with whole body variables",
			varsToSet: map[string]string{
				"wholeBodyVar": "",
			},
			body: "<Items>",
			want: map[string]string{
				"wholeBodyVar": "<Items>",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.description, func(t *testing.T) {
			result := &models.Result{
				ResponseBody:        tt.body,
				ResponseContentType: "xml",
			}
			got, err := ExtractValues(tt.varsToSet, result)

			if tt.wantErr != "" {
				require.Equal(t, 1, len(err))
				require.Error(t, err[0])
				require.EqualError(t, err[0], tt.wantErr)
			} else {
				require.Equal(t, 0, len(err), "not exected error: %v", err)
				require.Equal(t, tt.want, got)
			}
		})
	}
}

func Test_ExtractValuesFromPlainResponse(t *testing.T) {
	defaultBody := "plain text value"
	tests := []struct {
		description string
		varsToSet   map[string]string
		body        string
		want        map[string]string
		wantErr     string
	}{
		{
			description: "plain text body with valid variable",
			varsToSet: map[string]string{
				"var1": "",
			},
			body: defaultBody,
			want: map[string]string{
				"var1": defaultBody,
			},
		},
		{
			description: "plain text body with non-empty path",
			varsToSet: map[string]string{
				"var1": "some.path",
			},
			body:    defaultBody,
			wantErr: "variable 'var1': paths not supported for plain text body",
		},
		{
			description: "empty text body with path",
			varsToSet: map[string]string{
				"var1": "some.path",
			},
			body:    "",
			wantErr: "variable 'var1': paths not supported for empty body",
		},
	}

	for _, tt := range tests {
		t.Run(tt.description, func(t *testing.T) {
			result := &models.Result{
				ResponseBody:        tt.body,
				ResponseContentType: "text",
			}

			got, err := ExtractValues(tt.varsToSet, result)

			if tt.wantErr != "" {
				require.Equal(t, 1, len(err))
				require.Error(t, err[0])
				require.EqualError(t, err[0], tt.wantErr)
			} else {
				require.Equal(t, 0, len(err), "not exected error: %v", err)
				require.Equal(t, tt.want, got)
			}
		})
	}
}

func Test_ExtractValuesFromHeaders(t *testing.T) {
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
			got, err := ExtractValues(tt.varsToSet, result)

			if tt.wantErr != "" {
				require.Equal(t, 1, len(err))
				require.Error(t, err[0])
				require.EqualError(t, err[0], tt.wantErr)
			} else {
				require.Equal(t, 0, len(err))
				require.Equal(t, tt.want, got)
			}
		})
	}
}

func Test_ExtractValuesFromCookie(t *testing.T) {
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
			wantErr: "variable 'var2': 'Set-Cookie' header does not include expected cookie wrong_name",
		},
	}

	for _, tt := range tests {
		t.Run(tt.description, func(t *testing.T) {
			result := &models.Result{
				ResponseHeaders: headers,
			}
			got, err := ExtractValues(tt.varsToSet, result)

			if tt.wantErr != "" {
				require.Equal(t, 1, len(err))
				require.Error(t, err[0])
				require.EqualError(t, err[0], tt.wantErr)
			} else {
				require.Equal(t, 0, len(err))
				require.Equal(t, tt.want, got)
			}
		})
	}
}
