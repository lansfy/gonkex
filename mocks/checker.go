package mocks

import (
	"net/http"
)

type CheckerInterface interface {
	CheckRequest(mockName string, req *http.Request, resp *http.Response) []error
}
