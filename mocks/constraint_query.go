package mocks

import (
	"fmt"
	"net/http"
	"net/url"
	"sort"
	"strings"

	"github.com/lansfy/gonkex/colorize"
)

func loadQueryConstraint(def map[string]interface{}) (verifier, error) {
	return makeQueryConstraint(def, newQueryConstraint)
}

func loadQueryRegexpConstraint(def map[string]interface{}) (verifier, error) {
	return makeQueryConstraint(def, newQueryRegexpConstraint)
}

func makeQueryConstraint(def map[string]interface{},
	creator func(query string) (*queryConstraint, error)) (verifier, error) {
	key := "expectedQuery" // backward compatibility key name
	if !hasKey(def, key) {
		key = "query"
	}
	query, err := getRequiredStringKey(def, key, false)
	if err != nil {
		return nil, err
	}
	return creator(query)
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

func newQueryConstraint(query string) (*queryConstraint, error) {
	// user may begin his query with '?', just omit it in this case
	query = strings.TrimPrefix(query, "?")
	pq, err := url.ParseQuery(query)
	if err != nil {
		return nil, err
	}

	return &queryConstraint{
		name:          "queryMatches",
		expectedQuery: pq,
	}, nil
}

type queryConstraint struct {
	name          string
	expectedQuery map[string][]string
}

func (c *queryConstraint) GetName() string {
	return c.name
}

func (c *queryConstraint) Verify(r *http.Request) []error {
	expectedKeys := []string{}
	for key := range c.expectedQuery {
		expectedKeys = append(expectedKeys, key)
	}
	sort.Strings(expectedKeys)

	errors := []error{}
	gotQuery := r.URL.Query()
	for _, key := range expectedKeys {
		got, ok := gotQuery[key]
		if !ok {
			errors = append(errors, fmt.Errorf("'%s' parameter is missing in request query", key))
			continue
		}
		expected := c.expectedQuery[key]

		if len(expected) != len(got) {
			sort.Strings(expected)
			sort.Strings(got)
			errors = append(errors, colorize.NewNotEqualError(
				"number of values for parameter %s is not equal to expected:",
				key, got, expected))
			continue
		}

		errors = append(errors, compareValues("parameter %s", key, expected, got)...)
	}
	return errors
}
