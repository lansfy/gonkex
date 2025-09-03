package mocks

import (
	"net/http"
)

// CheckerInterface defines an interface for validating mock HTTP requests and responses.
// Implementations of this interface perform custom checks on incoming mock requests and their responses,
// returning any errors encountered during validation.
// Checkers can be assigned to a mock using the RegisterChecker method.
// Note: Checkers may not restore the response body after performing validation.
type CheckerInterface interface {
	// CheckRequest validates a mock HTTP request and its corresponding response.
	//
	// Parameters:
	// - mockName: The name of the mock service handling the request.
	// - req: The HTTP request being checked.
	// - resp: The generated HTTP response for the request.
	//
	// Returns:
	// - A slice of errors encountered during validation.
	CheckRequest(mockName string, req *http.Request, resp *http.Response) []error
}
