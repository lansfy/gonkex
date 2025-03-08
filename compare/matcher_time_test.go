package compare

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_TimeMatcher_MatchValues(t *testing.T) {
	tests := []struct {
		description string
		matcher     *timeMatcher
		actual      interface{}
		wantErr     string
	}{
		{
			description: "valid time match",
			matcher:     &timeMatcher{data: "%Y-%m-%d"},
			actual:      "2023-12-25",
		},
		{
			description: "invalid time format",
			matcher:     &timeMatcher{data: "%Y-%m-%d"},
			actual:      "12-25-2023",
			wantErr:     "'error-prefix': time does not match the template:\n     expected: $matchTime(%Y-%m-%d)\n       actual: 12-25-2023",
		},
	}

	for _, tt := range tests {
		t.Run(tt.description, func(t *testing.T) {
			err := tt.matcher.MatchValues("%s:", "error-prefix", tt.actual)
			if tt.wantErr == "" {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
				require.Equal(t, tt.wantErr, err.Error())
			}
		})
	}
}

func Test_convertPythonToGoFormat(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "basic date-time format",
			input: "%Y-%m-%d %H:%M:%S",
			want:  "2006-01-02 15:04:05",
		},
		{
			name:  "month and day",
			input: "%B %d, %Y",
			want:  "January 02, 2006",
		},
		{
			name:  "12-hour format with AM/PM",
			input: "%I:%M %p",
			want:  "03:04 PM",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := convertPythonToGoFormat(tt.input)
			require.Equal(t, tt.want, got)
		})
	}
}
