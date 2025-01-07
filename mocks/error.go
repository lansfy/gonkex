package mocks

import (
	"fmt"
	"net/http"
	"net/http/httputil"
)

func unhandledRequestError(r *http.Request) []error {
	requestContent, err := httputil.DumpRequest(r, true)
	if err != nil {
		return []error{fmt.Errorf("gonkex internal error during request dump: %s", err)}
	}
	return []error{fmt.Errorf("unhandled request to mock:\n%s", requestContent)}
}
