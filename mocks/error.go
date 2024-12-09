package mocks

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"reflect"
)

type Error struct {
	error
	ServiceName string
}

func (e *Error) Error() string {
	return fmt.Sprintf("mock %s: %s", e.ServiceName, e.error.Error())
}

type RequestConstraintError struct {
	error
	Constraint  verifier
	RequestDump []byte
}

func (e *RequestConstraintError) Error() string {
	kind := reflect.TypeOf(e.Constraint).String()
	return fmt.Sprintf("request constraint %s failed: %s, request was:\n %s", kind, e.error.Error(), e.RequestDump)
}

func unhandledRequestError(r *http.Request) []error {
	requestContent, err := httputil.DumpRequest(r, true)
	if err != nil {
		return []error{fmt.Errorf("Gonkex internal error during request dump: %s\n", err)}
	}
	return []error{fmt.Errorf("unhandled request to mock:\n%s", requestContent)}
}
