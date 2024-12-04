package mocks

import (
	"net/http"
	"net/url"
	"reflect"
	"testing"
)

func Test_newQueryRegexpConstraint(t *testing.T) {
	tests := []struct {
		name  string
		query string
		want  queryRegexpConstraint
	}{
		{
			name:  "simple expQuery",
			query: "a=1&b=2&a=3",
			want:  queryRegexpConstraint{expectedQuery: map[string][]string{"a": {"1", "3"}, "b": {"2"}}},
		},
		{
			name:  "expQuery written with '?'",
			query: "?a=1&b=2&a=3",
			want:  queryRegexpConstraint{expectedQuery: map[string][]string{"a": {"1", "3"}, "b": {"2"}}},
		},
		{
			name:  "expQuery contains multiple '?'",
			query: "?a=1&b=?&a=3",
			want:  queryRegexpConstraint{expectedQuery: map[string][]string{"a": {"1", "3"}, "b": {"?"}}},
		},
		{
			name:  "expQuery contains 'matchRegexp'",
			query: "a=1&b=$matchRegexp(\\d+)&a=3",
			want:  queryRegexpConstraint{expectedQuery: map[string][]string{"a": {"1", "3"}, "b": {"$matchRegexp(\\d+)"}}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := newQueryRegexpConstraint(tt.query)
			if err != nil {
				t.Errorf("newQueryRegexpConstraint() error = %v", err)
				return
			}
			if got == nil {
				t.Fatalf("unexpected. got nil instead of queryRegexpConstraint")
			}
			if !reflect.DeepEqual(*got, tt.want) {
				t.Errorf("newQueryRegexpConstraint() = %v, want %v", *got, tt.want)
			}
		})
	}
}

func Test_queryRegexpConstraint_Verify(t *testing.T) {
	tests := []struct {
		name       string
		expQuery   url.Values
		req        *http.Request
		wantErrors int
	}{
		{
			name:       "expected",
			expQuery:   map[string][]string{"food": {"cake", "tea"}, "people": {"2"}},
			req:        newTestRequest("food=cake&food=tea&people=2"),
			wantErrors: 0,
		},
		{
			name:       "expected but different order",
			expQuery:   map[string][]string{"food": {"cake", "tea"}, "people": {"2"}},
			req:        newTestRequest("food=tea&food=cake&people=2"),
			wantErrors: 0,
		},
		{
			name:       "unexpected value",
			expQuery:   map[string][]string{"food": {"cake", "tea"}, "people": {"2"}},
			req:        newTestRequest("food=cake&food=beer&people=3"),
			wantErrors: 2,
		},
		{
			name:       "key is missing",
			expQuery:   map[string][]string{"food": {"cake", "tea"}, "people": {"2"}},
			req:        newTestRequest("food=cake&food=tea"),
			wantErrors: 1,
		},
		{
			name:       "unexpected keys are ignored is missing",
			expQuery:   map[string][]string{"food": {"cake", "tea"}, "people": {"2"}},
			req:        newTestRequest("food=cake&food=tea&people=2&one-more=person"),
			wantErrors: 0,
		},
		{
			name:       "regexp in expected query",
			expQuery:   map[string][]string{"food": {"cake", "$matchRegexp(\\w+)"}, "people": {"$matchRegexp(\\d+)"}},
			req:        newTestRequest("food=cake&food=tea&people=2675"),
			wantErrors: 0,
		},
		{
			name:       "expected and actual parameters have different lengths",
			expQuery:   map[string][]string{"food": {"cake", "tea"}, "people": {"2"}},
			req:        newTestRequest("food=cake&food=tea&food=coffee&people=2"),
			wantErrors: 1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &queryRegexpConstraint{
				expectedQuery: tt.expQuery,
			}
			if gotErrors := c.Verify(tt.req); len(gotErrors) != tt.wantErrors {
				t.Errorf("unexpected amount of errors. Got %v, want %v. Errors are: '%v'",
					len(gotErrors), tt.wantErrors, gotErrors,
				)
			}
		})
	}
}
