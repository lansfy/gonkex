package models

type DatabaseResult struct {
	Query    string
	Response []string
}

// Result of test execution
type Result struct {
	Path                string // TODO: remove
	Query               string // TODO: remove
	RequestBody         string
	ResponseStatusCode  int
	ResponseStatus      string
	ResponseContentType string
	ResponseBody        string
	ResponseHeaders     map[string][]string
	Errors              []error
	Test                TestInterface
	DatabaseResult      []DatabaseResult
}

// Passed returns true if test passed (false otherwise)
func (r *Result) Passed() bool {
	return len(r.Errors) == 0
}
