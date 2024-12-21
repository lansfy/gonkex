package mocks

import (
	"fmt"
	"net/http"
	"regexp"

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

	if s, ok := compare.StringAsRegexp(valueStr); ok {
		valueStr = ""
		regexpStr = s
	}

	return newHeaderConstraint(header, valueStr, regexpStr)
}

func newHeaderConstraint(header, value, re string) (verifier, error) {
	var reCompiled *regexp.Regexp
	if re != "" {
		var err error
		reCompiled, err = regexp.Compile(re)
		if err != nil {
			return nil, err
		}
	}
	res := &headerConstraint{
		header: header,
		value:  value,
		regexp: reCompiled,
	}
	return res, nil
}

type headerConstraint struct {
	header string
	value  string
	regexp *regexp.Regexp
}

func (c *headerConstraint) GetName() string {
	return "headerIs"
}

func (c *headerConstraint) Verify(r *http.Request) []error {
	value := r.Header.Get(c.header)
	if value == "" {
		return []error{fmt.Errorf("request doesn't have header %s", c.header)}
	}
	if c.value != "" && c.value != value {
		return []error{fmt.Errorf("%s header value %s doesn't match expected %s", c.header, value, c.value)}
	}
	if c.regexp != nil && !c.regexp.MatchString(value) {
		return []error{fmt.Errorf("%s header value %s doesn't match regexp %s", c.header, value, c.regexp)}
	}
	return nil
}
