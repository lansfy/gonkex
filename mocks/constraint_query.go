package mocks

import (
	"fmt"
	"net/http"
	"net/url"
	"reflect"
	"sort"
	"strings"
)

func loadQueryConstraint(def map[interface{}]interface{}) (verifier, error) {
	query, err := getRequiredStringKey(def, "expectedQuery", false)
	if err != nil {
		return nil, err
	}
	return newQueryConstraint(query)
}

func newQueryConstraint(query string) (*queryConstraint, error) {
	// user may begin his query with '?', just omit it in this case
	query = strings.TrimPrefix(query, "?")
	pq, err := url.ParseQuery(query)
	if err != nil {
		return nil, err
	}

	return &queryConstraint{expectedQuery: pq}, nil
}

type queryConstraint struct {
	expectedQuery url.Values
}

func (c *queryConstraint) GetName() string {
	return "queryMatches"
}

func (c *queryConstraint) Verify(r *http.Request) (errors []error) {
	gotQuery := r.URL.Query()
	for key, want := range c.expectedQuery {
		got, ok := gotQuery[key]
		if !ok {
			errors = append(errors, fmt.Errorf("'%s' parameter is missing in expQuery", key))
			continue
		}

		sort.Strings(got)
		sort.Strings(want)
		if !reflect.DeepEqual(got, want) {
			errors = append(errors, fmt.Errorf(
				"'%s' parameters are not equal.\n Got: %s \n Want: %s", key, got, want,
			))
		}
	}

	return errors
}
