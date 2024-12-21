package mocks

import (
	"fmt"
	"net/http"
	"strings"
)

func loadMethodConstraint(def map[interface{}]interface{}) (verifier, error) {
	method, err := getRequiredStringKey(def, "method", false)
	if err != nil {
		return nil, err
	}
	return &methodConstraint{
		name:   "methodIs",
		method: method,
	}, nil
}

type methodConstraint struct {
	name   string
	method string
}

func (c *methodConstraint) GetName() string {
	return c.name
}

func (c *methodConstraint) Verify(r *http.Request) []error {
	if !strings.EqualFold(r.Method, c.method) {
		return []error{fmt.Errorf("method does not match: expected %s, actual %s", r.Method, c.method)}
	}
	return nil
}
