package mocks

import (
	"fmt"
	"net/http"
	"net/http/httputil"
)

func dumpRequest(r *http.Request) string {
	requestDump, err := httputil.DumpRequest(r, true)
	if err != nil {
		return fmt.Sprintf("gonkex internal error during request dump: %s", err)
	}
	return string(requestDump)
}

func unhandledRequestError(r *http.Request) []error {
	return []error{fmt.Errorf("unhandled request to mock:\n%s", dumpRequest(r))}
}
