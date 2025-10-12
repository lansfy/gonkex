package compare

import (
	"testing"

	"github.com/lansfy/gonkex/colorize"

	"github.com/stretchr/testify/require"
)

type matcherTest struct {
	description string
	matcher     interface{}
	actual      interface{}
	wantErr     string
}

func processTests(t *testing.T, tests []matcherTest, params Params) {
	t.Helper()
	for _, tt := range tests {
		t.Run(tt.description, func(t *testing.T) {
			require.NotNil(t, CreateMatcher(tt.matcher))
			expected := map[string]interface{}{
				"data": tt.matcher,
			}
			actual := map[string]interface{}{
				"data": tt.actual,
			}
			errs := Compare(expected, actual, params)
			if tt.wantErr == "" {
				require.Empty(t, errs)
			} else {
				require.Equal(t, 1, len(errs))
				require.Equal(t, tt.wantErr, colorize.RemovePathComponent(errs[0]).Error())
			}
		})
	}
}
