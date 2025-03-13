package mocks

import (
	"net/http"

	"github.com/lansfy/gonkex/colorize"
	"github.com/lansfy/gonkex/compare"
)

func loadHeaderConstraint(def map[interface{}]interface{}) (verifier, error) {
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

	matcher := compare.StringAsMatcher(compare.MatchRegexpWrap(regexpStr))
	if m := compare.StringAsMatcher(valueStr); m != nil {
		valueStr = ""
		matcher = m
	}

	return newHeaderConstraint(header, valueStr, matcher), nil
}

func newHeaderConstraint(header, value string, matcher compare.Matcher) verifier {
	return &headerConstraint{
		header:  header,
		value:   value,
		matcher: matcher,
	}
}

type headerConstraint struct {
	header  string
	value   string
	matcher compare.Matcher
}

func (c *headerConstraint) GetName() string {
	return "headerIs"
}

func (c *headerConstraint) Verify(r *http.Request) []error {
	value := r.Header.Get(c.header)
	if value == "" {
		return []error{colorize.NewEntityError("request does not have header %s", c.header)}
	}
	if c.value != "" && c.value != value {
		return []error{colorize.NewNotEqualError("%s header value does not match:", c.header, c.value, value)}
	}
	if c.matcher != nil {
		err := c.matcher.MatchValues("%s header:", c.header, value)
		if err != nil {
			return []error{err}
		}
	}
	return nil
}
