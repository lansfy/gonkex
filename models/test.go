package models

import (
	"time"
)

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

type RetryPolicy interface {
	Attempts() int
	Delay() time.Duration
	SuccessCount() int
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

	GetMeta() map[string]interface{}

	GetStatus() string
	GetResponses() map[int]string
	GetResponse(code int) (string, bool)
	GetResponseHeaders(code int) (map[string]string, bool)

	GetDatabaseChecks() []DatabaseCheck
	GetComparisonParams() ComparisonParams
	GetRetryPolicy() RetryPolicy

	Fixtures() []string
	ServiceMocks() map[string]interface{}

	Pause() time.Duration

	BeforeScriptPath() string
	BeforeScriptTimeout() time.Duration

	AfterRequestScriptPath() string
	AfterRequestScriptTimeout() time.Duration

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
