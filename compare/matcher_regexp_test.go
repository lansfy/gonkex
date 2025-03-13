package compare

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_RegexpMatcher_MatchValues(t *testing.T) {
	tests := []struct {
		description string
		matcher     *regexpMatcher
		actual      interface{}
		wantErr     string
	}{
		{
			description: "valid regex match",
			matcher:     &regexpMatcher{data: "^test$"},
			actual:      "test",
		},
		{
			description: "valid regex, but no match",
			matcher:     &regexpMatcher{data: "^test$"},
			actual:      "not-equal",
			wantErr:     "'error-prefix': value does not match regexp:\n     expected: $matchRegexp(^test$)\n       actual: not-equal",
		},
		{
			description: "invalid regex",
			matcher:     &regexpMatcher{data: "(unclosed"},
			actual:      "test",
			wantErr:     "'error-prefix': cannot compile regexp:\n     expected: <nil>\n       actual: missing closing ): `(unclosed`",
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
