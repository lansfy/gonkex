package mocks

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/lansfy/gonkex/compare"
)

func loadQueryRegexpConstraint(def map[interface{}]interface{}) (verifier, error) {
	query, err := getRequiredStringKey(def, "expectedQuery", false)
	if err != nil {
		return nil, err
	}
	return newQueryRegexpConstraint(query)
}

func newQueryRegexpConstraint(query string) (*queryRegexpConstraint, error) {
	// user may begin his query with '?', just omit it in this case
	query = strings.TrimPrefix(query, "?")

	rawParams := strings.Split(query, "&")

	expectedQuery := map[string][]string{}
	for _, rawParam := range rawParams {
		parts := strings.Split(rawParam, "=")
		if len(parts) != 2 {
			return nil, fmt.Errorf("error parsing query: got %d parts, expected 2", len(parts))
		}

		_, ok := expectedQuery[parts[0]]
		if !ok {
			expectedQuery[parts[0]] = make([]string, 0)
		}
		expectedQuery[parts[0]] = append(expectedQuery[parts[0]], parts[1])
	}

	return &queryRegexpConstraint{expectedQuery}, nil
}

type queryRegexpConstraint struct {
	expectedQuery map[string][]string
}

func (c *queryRegexpConstraint) Verify(r *http.Request) (errors []error) {
	gotQuery := r.URL.Query()
	for key, want := range c.expectedQuery {
		got, ok := gotQuery[key]
		if !ok {
			errors = append(errors, fmt.Errorf("'%s' parameter is missing in expQuery", key))
			continue
		}

		if ok, err := compare.Query(want, got); err != nil {
			errors = append(errors, fmt.Errorf(
				"'%s' parameters comparison failed. \n %s'", key, err.Error(),
			))
		} else if !ok {
			errors = append(errors, fmt.Errorf(
				"'%s' parameters are not equal.\n Got: %s \n Want: %s", key, got, want,
			))
		}
	}

	return errors
}
