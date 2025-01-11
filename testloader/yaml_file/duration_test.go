package yaml_file

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v2"
)

func TestDurationUnmarshalYAML(t *testing.T) {
	tests := []struct {
		name     string
		yamlData string
		want     time.Duration
		wantErr  string
	}{
		{
			name:     "valid duration string",
			yamlData: "timeout: 10s",
			want:     10 * time.Second,
		},
		{
			name:     "valid integer (seconds)",
			yamlData: "timeout: 30",
			want:     30 * time.Second,
		},
		{
			name:     "valid float duration",
			yamlData: `timeout: 0.02`,
			want:     20 * time.Millisecond,
		},
		{
			name:     "invalid duration string",
			yamlData: "timeout: invalid",
			want:     0,
			wantErr:  "invalid duration string: time: invalid duration \"invalid\"",
		},
		{
			name:     "negative integer",
			yamlData: "timeout: -5",
			wantErr:  "invalid duration value: cannot be negative",
		},
		{
			name:     "wrong value type",
			yamlData: "timeout: {}",
			wantErr:  "invalid duration value: must be an integer (seconds) or a duration string",
		},
		{
			name:     "missing field",
			yamlData: "",
			want:     0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var config struct {
				Timeout Duration `yaml:"timeout"`
			}

			err := yaml.Unmarshal([]byte(tt.yamlData), &config)
			if tt.wantErr != "" {
				require.Error(t, err)
				require.Equal(t, tt.wantErr, err.Error())
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.want, config.Timeout.Duration)
			}
		})
	}
}

func TestDurationUnmarshalJSON(t *testing.T) {
	tests := []struct {
		name     string
		jsonData string
		want     time.Duration
		wantErr  string
	}{
		{
			name:     "valid duration string",
			jsonData: `{"timeout": "10s"}`,
			want:     10 * time.Second,
		},
		{
			name:     "valid float duration",
			jsonData: `{"timeout": 0.55}`,
			want:     550 * time.Millisecond,
		},
		{
			name:     "valid integer (seconds)",
			jsonData: `{"timeout": 30}`,
			want:     30 * time.Second,
		},
		{
			name:     "invalid duration string",
			jsonData: `{"timeout": "invalid"}`,
			wantErr:  `invalid duration string: time: invalid duration "invalid"`,
		},
		{
			name:     "negative integer",
			jsonData: `{"timeout": -5}`,
			wantErr:  "invalid duration value: cannot be negative",
		},
		{
			name:     "empty value",
			jsonData: `{"timeout": ""}`,
			wantErr:  `invalid duration string: time: invalid duration ""`,
		},
		{
			name:     "wrong type",
			jsonData: `{"timeout": {}}`,
			wantErr:  `invalid duration value: must be an integer (seconds) or a duration string`,
		},
		{
			name:     "null value",
			jsonData: `{"timeout": null}`,
			want:     0,
		},
		{
			name:     "missing field",
			jsonData: `{}`,
			want:     0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var config struct {
				Timeout Duration `json:"timeout"`
			}

			err := json.Unmarshal([]byte(tt.jsonData), &config)
			if tt.wantErr != "" {
				require.Error(t, err)
				require.Equal(t, tt.wantErr, err.Error())
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.want, config.Timeout.Duration)
			}
		})
	}
}
