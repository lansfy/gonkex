package mocks

import (
	"net/http"
)

type Checker interface {
	Check(serviceName string, req *http.Request, resp *http.Response) []error
}
