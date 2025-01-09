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
	GetName() string
	GetDescription() string

	GetMethod() string
	Path() string
	ToQuery() string
	ContentType() string
	Headers() map[string]string
	Cookies() map[string]string
	GetRequest() string
	GetForm() *Form

	GetStatus() string
	GetResponses() map[int]string
	GetResponse(code int) (string, bool)
	GetResponseHeaders(code int) (map[string]string, bool)

	GetDatabaseChecks() []DatabaseCheck
	GetComparisonParams() ComparisonParams

	Fixtures() []string
	ServiceMocks() map[string]interface{}

	Pause() int

	BeforeScriptPath() string
	BeforeScriptTimeout() int

	AfterRequestScriptPath() string
	AfterRequestScriptTimeout() int

	GetVariables() map[string]string
	GetCombinedVariables() map[string]string
	GetVariablesToSet() map[int]map[string]string

	GetFileName() string
	FirstTestInFile() bool
	LastTestInFile() bool

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
