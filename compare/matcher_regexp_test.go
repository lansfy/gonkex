package compare

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_RegexpMatcher_MatchValues(t *testing.T) {
	tests := []struct {
		description string
		matcher     Matcher
		actual      interface{}
		wantErr     string
	}{
		{
			description: "valid regex match",
			matcher:     StringAsMatcher("$matchRegexp(^test$)"),
			actual:      "test",
		},
		{
			description: "valid regex, but no match",
			matcher:     StringAsMatcher("$matchRegexp(^test$)"),
			actual:      "not-equal",
			wantErr:     "value does not match regexp:\n     expected: $matchRegexp(^test$)\n       actual: not-equal",
		},
		{
			description: "invalid regex",
			matcher:     StringAsMatcher("$matchRegexp((unclosed)"),
			actual:      "test",
			wantErr:     "cannot compile regexp:\n     expected: <nil>\n       actual: missing closing ): `(unclosed`",
		},
	}

	for _, tt := range tests {
		t.Run(tt.description, func(t *testing.T) {
			require.NotNil(t, tt.matcher)
			err := tt.matcher.MatchValues(tt.actual)
			if tt.wantErr == "" {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
				require.Equal(t, tt.wantErr, err.Error())
			}
		})
	}
}
