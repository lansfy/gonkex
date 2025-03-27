package models

// DatabaseResult represents the result of a database check
// Contains both the query that was executed and the response records
type DatabaseResult struct {
	Query    string   // The storage query that was executed
	Response []string // The records returned from the database as JSON items serialized to strings
}

// Result contains the complete results of a test execution
// Includes both HTTP request/response details and database check results
type Result struct {
	Path        string // The request path (TODO: remove)
	Query       string // The query string (TODO: remove)
	RequestBody string // The HTTP request body that was sent

	ResponseStatusCode  int                 // The HTTP status code received
	ResponseStatus      string              // The HTTP status text received
	ResponseContentType string              // The content type of the response
	ResponseHeaders     map[string][]string // All HTTP response headers
	ResponseBody        string              // The body of the HTTP response

	Errors         []error          // Any errors encountered during test execution
	Test           TestInterface    // Reference to the test case that was executed
	DatabaseResult []DatabaseResult // Results of database checks after the request
}

// Passed returns true if the test execution passed without errors
// A test passes when there are no errors in the result
func (r *Result) Passed() bool {
	return len(r.Errors) == 0
}
