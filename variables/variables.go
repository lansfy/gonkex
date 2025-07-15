// Package variables provides functionality for managing and substituting variables in tests
package variables

// Variables represents a storage for variable names and their values.
// Used for variable substitution in test descriptions, paths, queries, headers, requests, responses, etc.
type Variables interface {
	// Set adds a new variable (or replaces an existing one) to the variables map
	Set(name, value string)
	// Merge adds new variables and replaces values of existing ones from a provided map
	// Useful when applying variables from multiple sources (e.g., test cases, environment)
	Merge(variables map[string]string)
	// Substitute replaces all variable references in a string with their actual values
	// VariablesImpl are specified in the format "{{ $varname }}" in the input string
	// The method first looks for the variable in the internal variables map,
	// and if not found, checks environment variables
	// Returns the input string with all recognized variables replaced with their values
	Substitute(s string) string
	// Len returns the number of variables currently stored in the variables map
	Len() int
}
