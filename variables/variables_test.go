package variables

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_SetAndGetVariables(t *testing.T) {
	vs := New()
	vs.Set("foo", "bar")
	vs.Set("baz", "qux")

	tests := []struct {
		description string
		key         string
		want        string
	}{
		{
			description: "existing variable",
			key:         "foo",
			want:        "bar",
		},
		{
			description: "non-existing variable",
			key:         "unknown",
			want:        "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.description, func(t *testing.T) {
			got, ok := vs.get(tt.key)

			require.Equal(t, tt.want != "", ok)
			require.Equal(t, tt.want, got)
		})
	}
}

func Test_Substitute(t *testing.T) {
	vs := New()
	vs.Set("name", "John")
	vs.Set("city", "New York")

	os.Setenv("ENV_VAR", "12345")
	defer os.Unsetenv("ENV_VAR")

	tests := []struct {
		description string
		input       string
		want        string
	}{
		{
			description: "substitute existing variables",
			input:       "Hello {{ $name }} from {{ $city }}!",
			want:        "Hello John from New York!",
		},
		{
			description: "substitute with environment variable",
			input:       "Environment variable: {{ $ENV_VAR }}",
			want:        "Environment variable: 12345",
		},
		{
			description: "substitute non-existing variable",
			input:       "This is {{ $unknown_var }}",
			want:        "This is {{ $unknown_var }}",
		},
	}

	for _, tt := range tests {
		t.Run(tt.description, func(t *testing.T) {
			got := vs.Substitute(tt.input)
			require.Equal(t, tt.want, got)
		})
	}
}

func Test_Len(t *testing.T) {
	vs := New()
	require.Equal(t, 0, vs.Len())

	vs.Set("key1", "value1")
	require.Equal(t, 1, vs.Len())

	vs.Set("key2", "value2")
	require.Equal(t, 2, vs.Len())

	vs.Set("key2", "value3")
	require.Equal(t, 2, vs.Len())
}

func Test_Merge(t *testing.T) {
	vs1 := New()
	vs1.Set("key1", "value1")
	vs1.Set("key3", "value3")

	vs2 := map[string]string{
		"key1": "overridden",
		"key2": "value2",
	}

	vs1.Merge(vs2)

	key1Value, _ := vs1.get("key1")
	key2Value, _ := vs1.get("key2")
	key3Value, _ := vs1.get("key3")

	require.Equal(t, "overridden", key1Value)
	require.Equal(t, "value2", key2Value)
	require.Equal(t, "value3", key3Value)
}
