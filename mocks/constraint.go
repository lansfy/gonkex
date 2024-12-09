package mocks

import (
	"net/http"
)

type verifier interface {
	Verify(r *http.Request) []error
}
