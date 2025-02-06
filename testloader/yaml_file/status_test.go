package yaml_file

import (
	"encoding/json"
	"testing"

	"github.com/lansfy/gonkex/models"

	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v2"
)

func TestStatusEnumUnmarshalYAML(t *testing.T) {
	tests := []struct {
		name     string
		yamlData string
		want     models.Status
		wantErr  string
	}{
		{
			name:     "missing field",
			yamlData: "",
			want:     models.StatusNone,
		},
		{
			name:     "valid status string (none)",
			yamlData: "status: \"\"",
			want:     models.StatusNone,
		},
		{
			name:     "valid status string (skipped)",
			yamlData: "status: skipped",
			want:     models.StatusSkipped,
		},
		{
			name:     "valid status string (broken)",
			yamlData: "status: broken",
			want:     models.StatusBroken,
		},
		{
			name:     "valid status string (focus)",
			yamlData: "status: focus",
			want:     models.StatusFocus,
		},
		{
			name:     "invalid status value",
			yamlData: "status: fake",
			wantErr:  "unsupported value for status: fake",
		},
		{
			name:     "wrong status type",
			yamlData: "status: {}",
			wantErr:  "wrong type for status value",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var config struct {
				Status StatusEnum `yaml:"status"`
			}

			err := yaml.Unmarshal([]byte(tt.yamlData), &config)
			if tt.wantErr != "" {
				require.Error(t, err)
				require.Equal(t, tt.wantErr, err.Error())
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.want, config.Status.value)
			}
		})
	}
}

func TestStatusUnmarshalJSON(t *testing.T) {
	tests := []struct {
		name     string
		jsonData string
		want     models.Status
		wantErr  string
	}{
		{
			name:     "missing field",
			jsonData: `{}`,
			want:     models.StatusNone,
		},
		{
			name:     "null value",
			jsonData: `{"status": null}`,
			want:     models.StatusNone,
		},
		{
			name:     "valid status string (none)",
			jsonData: `{"status": ""}`,
			want:     models.StatusNone,
		},
		{
			name:     "valid status string (skipped)",
			jsonData: `{"status": "skipped"}`,
			want:     models.StatusSkipped,
		},
		{
			name:     "valid status string (broken)",
			jsonData: `{"status": "broken"}`,
			want:     models.StatusBroken,
		},
		{
			name:     "valid status string (focus)",
			jsonData: `{"status": "focus"}`,
			want:     models.StatusFocus,
		},
		{
			name:     "invalid status value",
			jsonData: `{"status": "fake"}`,
			wantErr:  "unsupported value for status: fake",
		},
		{
			name:     "wrong status type",
			jsonData: `{"status": {}}`,
			wantErr:  "wrong type for status value",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var config struct {
				Status StatusEnum `json:"status"`
			}

			err := json.Unmarshal([]byte(tt.jsonData), &config)
			if tt.wantErr != "" {
				require.Error(t, err)
				require.Equal(t, tt.wantErr, err.Error())
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.want, config.Status.value)
			}
		})
	}
}
