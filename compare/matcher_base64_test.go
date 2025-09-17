package compare

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_base64Matcher_MatchValues(t *testing.T) {
	tests := []struct {
		description string
		matcher     Matcher
		actual      interface{}
		wantErr     string
	}{
		{
			description: "valid base64 match",
			matcher:     StringAsMatcher("$matchBase64(somevalue)"),
			actual:      "c29tZXZhbHVl", // encoded "somevalue"
		},
		{
			description: "valid base64, but no match",
			matcher:     StringAsMatcher("$matchBase64(somevalue)"),
			actual:      "b3RoZXJ2YWx1ZQ==", // encoded "othervalue"
			wantErr:     "'error-prefix': base64 decoded value does not match:\n     expected: somevalue\n       actual: othervalue",
		},
		{
			description: "invalid base64",
			matcher:     StringAsMatcher("$matchBase64(somevalue)"),
			actual:      "inva$$id",
			wantErr:     "'error-prefix': cannot make base64 decode:\n     expected: <nil>\n       actual: illegal base64 data at input byte 4",
		},
	}

	for _, tt := range tests {
		t.Run(tt.description, func(t *testing.T) {
			require.NotNil(t, tt.matcher)
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
