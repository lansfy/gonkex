package models

type ComparisonParams interface {
	IgnoreValuesChecking() bool
	IgnoreArraysOrdering() bool
	DisallowExtraFields() bool
}

type DatabaseCheck interface {
	DbQueryString() string
	DbResponseJson() []string
	GetComparisonParams() ComparisonParams
}

// Common Test interface
type TestInterface interface {
	ToQuery() string
	GetRequest() string
	ToJSON() ([]byte, error)
	GetMethod() string
	Path() string
	GetResponses() map[int]string
	GetResponse(code int) (string, bool)
	GetResponseHeaders(code int) (map[string]string, bool)
	GetName() string
	GetDescription() string
	GetStatus() string
	Fixtures() []string
	ServiceMocks() map[string]interface{}
	Pause() int
	BeforeScriptPath() string
	BeforeScriptTimeout() int
	AfterRequestScriptPath() string
	AfterRequestScriptTimeout() int
	Cookies() map[string]string
	Headers() map[string]string
	ContentType() string
	GetForm() *Form

	GetDatabaseChecks() []DatabaseCheck
	GetComparisonParams() ComparisonParams

	GetVariables() map[string]string
	GetCombinedVariables() map[string]string
	GetVariablesToSet() map[int]map[string]string

	GetFileName() string

	SetStatus(status string)

	// ApplyVariables run specified function for every string in object
	ApplyVariables(func(string) string)
	// Clone returns copy of current object
	Clone() TestInterface
}

// TODO: add support for form fields
type Form struct {
	Files map[string]string `json:"files" yaml:"files"`
}

type Summary struct {
	Success bool
	Failed  int
	Skipped int
	Broken  int
	Total   int
}
