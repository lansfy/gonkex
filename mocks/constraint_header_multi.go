package mocks

import (
	"net/http"
	"sort"
)

func loadMultiHeadersConstraint(def map[string]interface{},
	headers map[string]string) (verifier, error) {
	keys := []string{}
	for name := range headers {
		keys = append(keys, name)
	}
	sort.Strings(keys)

	c := &multiHeadersConstraint{}
	for _, name := range keys {
		c.headers = append(c.headers, newHeaderConstraint(name, headers[name]))
	}
	return c, nil
}

type multiHeadersConstraint struct {
	headers []verifier
}

func (c *multiHeadersConstraint) GetName() string {
	return "headerIs"
}

func (c *multiHeadersConstraint) Verify(r *http.Request) []error {
	var errs []error
	for _, v := range c.headers {
		errs = append(errs, v.Verify(r)...)
	}
	return errs
}
