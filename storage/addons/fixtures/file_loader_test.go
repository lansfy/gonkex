package fixtures

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_FileLoader_Load(t *testing.T) {
	tests := []struct {
		description string
		name        string
		wantErr     string
		wantData    string
	}{
		{
			description: "file with exact name MUST be first",
			name:        "step1",
			wantData:    "content of step1",
		},
		{
			description: "file with .yml extension MUST be second",
			name:        "step2",
			wantData:    "content of step2",
		},
		{
			description: "file with .yaml extension MUST be third",
			name:        "step3",
			wantData:    "content of step3",
		},
		{
			description: "loader MUST ignore any dir with same name",
			name:        "step4",
			wantData:    "content of step4",
		},
		{
			description: "WHEN file not exists loader MUST return error",
			name:        "step5",
			wantErr:     "file not exists",
		},
	}

	for _, tt := range tests {
		t.Run(tt.description, func(t *testing.T) {
			loader := CreateFileLoader("testdata/loader")

			// first load
			file, data, err := loader.Load(tt.name)

			if tt.wantErr != "" {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.wantErr)
				return
			}

			require.NoError(t, err)
			require.Contains(t, file, "testdata/loader/"+tt.name)
			require.Equal(t, tt.wantData, string(data))

			// re-load the same file and expect empty data
			file2, data2, err2 := loader.Load(tt.name)
			require.NoError(t, err2)
			require.Equal(t, file, file2)
			require.Empty(t, data2)
		})
	}
}
