package yaml_file

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func TestVariablesToSetUnmarshalYAML(t *testing.T) {
	tests := []struct {
		name     string
		yamlData string
		want     VariablesToSet
		wantErr  string
	}{
		{
			name:     "empty yaml",
			yamlData: "",
			want:     nil,
		},
		{
			name: "plain text format - single variable",
			yamlData: `
variablesToSet:
  100: varName1`,
			want: VariablesToSet{
				100: {"varName1": ""},
			},
		},
		{
			name: "plain text format - multiple variables",
			yamlData: `
variablesToSet:
  100: varName1
  200: varName2
  300: varName3`,
			want: VariablesToSet{
				100: {"varName1": ""},
				200: {"varName2": ""},
				300: {"varName3": ""},
			},
		},
		{
			name: "json-paths format - single code with multiple variables",
			yamlData: `
variablesToSet:
  100:
    varName1: value1
    varName2: value2`,
			want: VariablesToSet{
				100: {
					"varName1": "value1",
					"varName2": "value2",
				},
			},
		},
		{
			name: "json-paths format - multiple codes",
			yamlData: `
variablesToSet:
  100:
    varName1: value1
    varName2: value2
  200:
    varName3: different.path`,
			want: VariablesToSet{
				100: {
					"varName1": "value1",
					"varName2": "value2",
				},
				200: {
					"varName3": "different.path",
				},
			},
		},
		{
			name: "json-paths format - empty paths",
			yamlData: `
variablesToSet:
  100:
    varName1: ""
    varName2: path`,
			want: VariablesToSet{
				100: {
					"varName1": "",
					"varName2": "path",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var config struct {
				VariablesToSet VariablesToSet `yaml:"variablesToSet"`
			}

			err := yaml.Unmarshal([]byte(tt.yamlData), &config)
			if tt.wantErr != "" {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.wantErr)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.want, config.VariablesToSet)
			}
		})
	}
}

func TestVariablesToSetUnmarshalJSON(t *testing.T) {
	tests := []struct {
		name     string
		jsonData string
		want     VariablesToSet
		wantErr  string
	}{
		{
			name:     "empty json",
			jsonData: `{}`,
			want:     nil,
		},
		{
			name:     "null value",
			jsonData: `{"variablesToSet": null}`,
			want:     VariablesToSet{},
		},
		{
			name:     "plain text format - single variable",
			jsonData: `{"variablesToSet": {"100": "varName1"}}`,
			want: VariablesToSet{
				100: {"varName1": ""},
			},
		},
		{
			name:     "plain text format - multiple variables",
			jsonData: `{"variablesToSet": {"100": "varName1", "200": "varName2", "300": "varName3"}}`,
			want: VariablesToSet{
				100: {"varName1": ""},
				200: {"varName2": ""},
				300: {"varName3": ""},
			},
		},
		{
			name:     "json-paths format - single code with multiple variables",
			jsonData: `{"variablesToSet": {"100": {"varName1": "value1", "varName2": "value2"}}}`,
			want: VariablesToSet{
				100: {
					"varName1": "value1",
					"varName2": "value2",
				},
			},
		},
		{
			name:     "json-paths format - multiple codes",
			jsonData: `{"variablesToSet": {"100": {"varName1": "value1", "varName2": "value2"}, "200": {"varName3": "different.path"}}}`,
			want: VariablesToSet{
				100: {
					"varName1": "value1",
					"varName2": "value2",
				},
				200: {
					"varName3": "different.path",
				},
			},
		},
		{
			name:     "json-paths format - empty paths",
			jsonData: `{"variablesToSet": {"100": {"varName1": "", "varName2": "path"}}}`,
			want: VariablesToSet{
				100: {
					"varName1": "",
					"varName2": "path",
				},
			},
		},
		{
			name:     "invalid json structure",
			jsonData: `{"variablesToSet": "invalid"}`,
			wantErr:  "cannot unmarshal string into Go struct",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var config struct {
				VariablesToSet VariablesToSet `json:"variablesToSet"`
			}

			err := json.Unmarshal([]byte(tt.jsonData), &config)
			if tt.wantErr != "" {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.wantErr)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.want, config.VariablesToSet)
			}
		})
	}
}

func TestVariablesToSetDirectUnmarshaling(t *testing.T) {
	t.Run("direct YAML unmarshaling", func(t *testing.T) {
		yamlData := `
100:
  varName1: path1
  varName2: path2
200: 
  varName3: path3`

		var variables VariablesToSet
		err := yaml.Unmarshal([]byte(yamlData), &variables)
		require.NoError(t, err)

		expected := VariablesToSet{
			100: {
				"varName1": "path1",
				"varName2": "path2",
			},
			200: {
				"varName3": "path3",
			},
		}
		require.Equal(t, expected, variables)
	})
	t.Run("direct JSON unmarshaling", func(t *testing.T) {
		jsonData := `{"100": {"varName1": "path1", "varName2": "path2"}, "200": {"varName3": "path3"}}`

		var variables VariablesToSet
		err := json.Unmarshal([]byte(jsonData), &variables)
		require.NoError(t, err)

		expected := VariablesToSet{
			100: {
				"varName1": "path1",
				"varName2": "path2",
			},
			200: {
				"varName3": "path3",
			},
		}
		require.Equal(t, expected, variables)
	})
}

func TestVariablesToSetEdgeCases(t *testing.T) {
	t.Run("zero code value", func(t *testing.T) {
		yamlData := `0: varName`

		var variables VariablesToSet
		err := yaml.Unmarshal([]byte(yamlData), &variables)
		require.NoError(t, err)

		expected := VariablesToSet{
			0: {"varName": ""},
		}
		require.Equal(t, expected, variables)
	})

	t.Run("negative code value", func(t *testing.T) {
		yamlData := `-1: varName`

		var variables VariablesToSet
		err := yaml.Unmarshal([]byte(yamlData), &variables)
		require.NoError(t, err)

		expected := VariablesToSet{
			-1: {"varName": ""},
		}
		require.Equal(t, expected, variables)
	})

	t.Run("large code value", func(t *testing.T) {
		yamlData := `
999999:
  varName: "path"`

		var variables VariablesToSet
		err := yaml.Unmarshal([]byte(yamlData), &variables)
		require.NoError(t, err)

		expected := VariablesToSet{
			999999: {"varName": "path"},
		}
		require.Equal(t, expected, variables)
	})

	t.Run("empty variable name in plain format", func(t *testing.T) {
		yamlData := `100: ""`

		var variables VariablesToSet
		err := yaml.Unmarshal([]byte(yamlData), &variables)
		require.NoError(t, err)

		expected := VariablesToSet{
			100: {"": ""},
		}
		require.Equal(t, expected, variables)
	})

	t.Run("empty variable name in json-paths format", func(t *testing.T) {
		yamlData := `
100:
  "": "path"`

		var variables VariablesToSet
		err := yaml.Unmarshal([]byte(yamlData), &variables)
		require.NoError(t, err)

		expected := VariablesToSet{
			100: {"": "path"},
		}
		require.Equal(t, expected, variables)
	})
}
