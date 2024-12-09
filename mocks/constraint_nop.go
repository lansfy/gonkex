package mocks

import (
	"net/http"
)

type nopConstraint struct{}

func (c *nopConstraint) Verify(r *http.Request) []error {
	return nil
}
