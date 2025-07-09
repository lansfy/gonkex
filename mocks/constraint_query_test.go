package mocks

import (
	"net/url"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_newQueryConstraint(t *testing.T) {
	tests := []struct {
		name  string
		query string
		want  queryConstraint
	}{
		{
			name:  "simple expQuery",
			query: "a=1&b=2&a=3",
			want:  queryConstraint{expectedQuery: url.Values{"a": {"1", "3"}, "b": {"2"}}},
		},
		{
			name:  "expQuery written with '?'",
			query: "?a=1&b=2&a=3",
			want:  queryConstraint{expectedQuery: url.Values{"a": {"1", "3"}, "b": {"2"}}},
		},
		{
			name:  "expQuery contains multiple '?'",
			query: "?a=1&b=?&a=3",
			want:  queryConstraint{expectedQuery: url.Values{"a": {"1", "3"}, "b": {"?"}}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.want.name = "queryMatches"
			got, err := newQueryConstraint(tt.query)
			require.NoError(t, err, "newQueryConstraint() returned an unexpected error")
			require.NotNil(t, got, "newQueryConstraint() returned nil")
			require.Equal(t, tt.want, *got, "newQueryConstraint() = %v, want %v", *got, tt.want)
		})
	}
}
