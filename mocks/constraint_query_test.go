package mocks

import (
	"net/http"
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
			got, err := newQueryConstraint(tt.query)
			require.NoError(t, err, "newQueryConstraint() returned an unexpected error")
			require.NotNil(t, got, "newQueryConstraint() returned nil")
			require.Equal(t, tt.want, *got, "newQueryConstraint() = %v, want %v", *got, tt.want)
		})
	}
}

func Test_queryConstraint_Verify(t *testing.T) {
	constraint, err := newQueryConstraint("people=2&food=tea&food=cake")
	require.NoError(t, err, "newQueryConstraint() returned an unexpected error")

	tests := []struct {
		name       string
		query      string
		wantErrors int
	}{
		{
			name:       "expected",
			query:      "people=2&food=tea&food=cake",
		},
		{
			name:       "different order (1)",
			query:      "food=tea&food=cake&people=2",
		},
		{
			name:       "different order (2)",
			query:      "food=cake&food=tea&people=2",
		},
		{
			name:       "different order (3)",
			query:      "people=2&food=cake&food=tea",
		},
		{
			name:       "unexpected keys are ignored",
			query:      "food=cake&food=tea&people=2&one-more=person",
		},
		{
			name:       "unexpected value",
			query:      "food=cake&food=beer&people=3",
			wantErrors: 2,
		},
		{
			name:       "key is missing",
			query:      "food=cake&food=tea",
			wantErrors: 1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := newTestRequest(tt.query)
			gotErrors := constraint.Verify(req)
			require.Len(t, gotErrors, tt.wantErrors, "unexpected amount of errors. Errors: '%v'", gotErrors)
		})
	}
}

func newTestRequest(query string) *http.Request {
	r, _ := http.NewRequest("GET", "http://localhost/?"+query, http.NoBody)
	return r
}
