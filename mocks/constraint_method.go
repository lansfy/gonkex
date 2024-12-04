package mocks

import (
	"fmt"
	"errors"
	"net/http"
	"strings"
)

func loadMethodConstraint(def map[interface{}]interface{}) (verifier, error) {
	c, ok := def["method"]
	if !ok {
		return nil, errors.New("`methodIs` requires `method` key")
	}
	method, ok := c.(string)
	if !ok || method == "" {
		return nil, errors.New("`method` must be string")
	}
	return &methodConstraint{method: method}, nil
}

type methodConstraint struct {
	method string
}

func (c *methodConstraint) Verify(r *http.Request) []error {
	if !strings.EqualFold(r.Method, c.method) {
		return []error{fmt.Errorf("method does not match: expected %s, actual %s", r.Method, c.method)}
	}
	return nil
}

func (c *methodConstraint) Fields() []string {
	return []string{"method"}
}
