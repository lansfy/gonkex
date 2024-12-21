package mocks

import (
	"net/http"
)

type verifier interface {
	GetName() string
	Verify(r *http.Request) []error
}
