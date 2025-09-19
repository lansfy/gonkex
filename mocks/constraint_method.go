package mocks

import (
	"net/http"
	"strings"

	"github.com/lansfy/gonkex/colorize"
)

func loadMethodConstraint(def map[string]interface{}) (verifier, error) {
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
		return []error{colorize.NewEntityNotEqualError("%s does not match:", "method", c.method, r.Method)}
	}
	return nil
}
