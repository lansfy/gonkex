package models

import (
	"time"
)

// Status represents the test execution status
// Can be used to mark tests as skipped, broken, or focused for execution
type Status string

const (
	StatusNone    Status = ""        // Default status, test will be executed normally
	StatusFocus   Status = "focus"   // Only tests with this status will be executed, others will be skipped
	StatusBroken  Status = "broken"  // Test is marked as broken and will not be executed
	StatusSkipped Status = "skipped" // Test will be skipped during execution
)

// ComparisonParams defines how responses should be compared
// Controls behavior of response comparison
type ComparisonParams interface {
	IgnoreValuesChecking() bool // If true, only structure is checked, values are ignored
	IgnoreArraysOrdering() bool // If true, arrays are considered equal regardless of element order
	DisallowExtraFields() bool  // If true, comparison fails if extra fields exist in compared structure
}

// DatabaseCheck represents a database query to be executed after an HTTP request
type DatabaseCheck interface {
	DbQueryString() string                 // storage query to execute against the database
	DbResponseJson() []string              // Expected records as serialized JSON strings
	GetComparisonParams() ComparisonParams // Comparison parameters for database response
}

// RetryPolicy defines how tests should be retried if they fail
type RetryPolicy interface {
	Attempts() int        // Number of retry attempts for failed tests
	Delay() time.Duration // Delay between retry attempts
	SuccessCount() int    // Required number of consecutive successful test passes to mark as successful
}

// Form represents multipart/form-data for file uploads and form submissions
type Form interface {
	GetFiles() map[string]string  // Map of field name to file path for file uploads
	GetFields() map[string]string // Map of field name to field value for form fields
}

// TestInterface defines the interface for Gonkex test cases
// Contains all methods necessary to execute a test
type TestInterface interface {
	GetName() string        // Test name, used for reporting
	GetDescription() string // Test description

	GetMethod() string          // HTTP method (GET, POST, etc.)
	Path() string               // URL path for the request
	ToQuery() string            // Query string parameters
	ContentType() string        // Content type header for the request
	Headers() map[string]string // HTTP headers for the request
	Cookies() map[string]string // Cookies for the request
	GetRequest() string         // Request body as string
	GetForm() Form              // Form data for multipart/form-data requests

	GetMeta(key string) interface{} // Additional metadata for the test

	GetStatus() Status                                     // Test execution status (focus, broken, skipped)
	GetResponses() map[int]string                          // Expected responses for different HTTP status codes
	GetResponse(code int) (string, bool)                   // Get expected response for a specific HTTP status code
	GetResponseHeaders(code int) (map[string]string, bool) // Get expected response headers for a specific status code

	Fixtures() []string                 // List of fixtures to load before test execution
	GetDatabaseChecks() []DatabaseCheck // Database checks to perform after the request

	GetComparisonParams() ComparisonParams // Comparison parameters for response checking
	GetRetryPolicy() RetryPolicy           // Retry policy for failed tests

	ServiceMocks() map[string]interface{} // Mocks for external services

	Pause() time.Duration             // Pause duration before test execution
	AfterRequestPause() time.Duration // Pause duration after request execution

	BeforeScriptPath() string           // Path to script to execute before the request
	BeforeScriptTimeout() time.Duration // Timeout for the before script

	AfterRequestScriptPath() string           // Path to script to execute after the request
	AfterRequestScriptTimeout() time.Duration // Timeout for the after request script

	GetVariables() map[string]string                      // Test-specific variables
	GetCombinedVariables() map[string]string              // Combined variables from all sources
	GetVariablesToSet(code int) (map[string]string, bool) // Variables to extract from the response

	GetFileName() string   // Source file name for the test
	FirstTestInFile() bool // Whether this is the first test in the file
	LastTestInFile() bool  // Whether this is the last test in the file
	OneOfCase() bool       // Whether this test is part of a case-based test

	SetStatus(status Status) // Set the execution status of the test

	// ApplyVariables applies a function to every string in the test
	// Used for variable substitution in test definitions
	ApplyVariables(func(string) string)
	// Clone returns a copy of the current test
	// Used when applying cases to create multiple test instances
	Clone() TestInterface
}
