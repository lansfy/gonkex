package models

import (
	"time"
)

type Status string

const (
	StatusNone    Status = ""
	StatusFocus   Status = "focus"
	StatusBroken  Status = "broken"
	StatusSkipped Status = "skipped"
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

type Form interface {
	GetFiles() map[string]string
	GetFields() map[string]string
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
	GetForm() Form

	GetMeta(key string) interface{}

	GetStatus() Status
	GetResponses() map[int]string
	GetResponse(code int) (string, bool)
	GetResponseHeaders(code int) (map[string]string, bool)

	GetDatabaseChecks() []DatabaseCheck
	GetComparisonParams() ComparisonParams
	GetRetryPolicy() RetryPolicy

	Fixtures() []string
	ServiceMocks() map[string]interface{}

	Pause() time.Duration
	AfterRequestPause() time.Duration

	BeforeScriptPath() string
	BeforeScriptTimeout() time.Duration

	AfterRequestScriptPath() string
	AfterRequestScriptTimeout() time.Duration

	GetVariables() map[string]string
	GetCombinedVariables() map[string]string
	GetVariablesToSet(code int) (map[string]string, bool)

	GetFileName() string
	FirstTestInFile() bool
	LastTestInFile() bool
	OneOfCase() bool

	SetStatus(status Status)

	// ApplyVariables run specified function for every string in object
	ApplyVariables(func(string) string)
	// Clone returns copy of current object
	Clone() TestInterface
}
