package mocks

import (
	"fmt"
	"strings"
)

func loadQueryRegexpConstraint(def map[interface{}]interface{}) (verifier, error) {
	query, err := getRequiredStringKey(def, "expectedQuery", false)
	if err != nil {
		return nil, err
	}
	return newQueryRegexpConstraint(query)
}

func newQueryRegexpConstraint(query string) (*queryConstraint, error) {
	// user may begin his query with '?', just omit it in this case
	query = strings.TrimPrefix(query, "?")

	expectedQuery := map[string][]string{}
	for _, rawParam := range strings.Split(query, "&") {
		parts := strings.Split(rawParam, "=")
		if len(parts) != 2 {
			return nil, fmt.Errorf("failed to parse query '%s'", rawParam)
		}

		_, ok := expectedQuery[parts[0]]
		if !ok {
			expectedQuery[parts[0]] = []string{}
		}
		expectedQuery[parts[0]] = append(expectedQuery[parts[0]], parts[1])
	}

	return &queryConstraint{"queryMatchesRegexp", expectedQuery}, nil
}
