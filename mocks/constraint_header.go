package mocks

import (
	"errors"
	"fmt"
	"net/http"
	"regexp"
)

func loadHeaderConstraint(def map[interface{}]interface{}) (verifier, error) {
	c, ok := def["header"]
	if !ok {
		return nil, errors.New("`headerIs` requires `header` key")
	}
	header, ok := c.(string)
	if !ok || header == "" {
		return nil, errors.New("`header` must be string")
	}
	var valueStr, regexpStr string
	if value, ok := def["value"]; ok {
		valueStr, ok = value.(string)
		if !ok {
			return nil, errors.New("`value` must be string")
		}
	}
	if regexp, ok := def["regexp"]; ok {
		regexpStr, ok = regexp.(string)
		if !ok || regexp == "" {
			return nil, errors.New("`regexp` must be string")
		}
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

func (c *headerConstraint) Fields() []string {
	return []string{"header", "value", "regexp"}
}
