package mocks

import (
	"net/http"

	"github.com/lansfy/gonkex/colorize"
	"github.com/lansfy/gonkex/compare"
)

func loadHeaderConstraint(def map[string]interface{}) (verifier, error) {
	header, err := getRequiredStringKey(def, "header", false)
	if err != nil {
		return nil, err
	}

	valueStr, err := getOptionalStringKey(def, "value", true)
	if err != nil {
		return nil, err
	}

	regexpStr, err := getOptionalStringKey(def, "regexp", false)
	if err != nil {
		return nil, err
	}

	if regexpStr != "" {
		valueStr = compare.MatchRegexpWrap(regexpStr)
	}

	return newHeaderConstraint(header, valueStr), nil
}

func newHeaderConstraint(header, value string) verifier {
	return &headerConstraint{
		header: header,
		value:  value,
	}
}

type headerConstraint struct {
	header string
	value  string
}

func (c *headerConstraint) GetName() string {
	return "headerIs"
}

func (c *headerConstraint) Verify(r *http.Request) []error {
	value := getHeader(r, c.header)
	if value == "" {
		return []error{colorize.NewEntityError("request does not have header %s", c.header)}
	}

	return compareValues("header %s", c.header, c.value, value)
}

func compareValues(pattern, entity string, expected, actual interface{}) []error {
	errs := compare.Compare(expected, actual, compare.Params{
		IgnoreArraysOrdering: true,
	})
	for idx := range errs {
		errs[idx] = colorize.NewEntityError(pattern, entity).WithSubError(
			colorize.RemovePathComponent(errs[idx]))
	}
	return errs
}
