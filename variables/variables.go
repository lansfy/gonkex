// Package variables provides functionality for managing and substituting variables in Gonkex tests
package variables

import (
	"os"
	"regexp"
	"strings"
)

// variableRx is a regular expression that matches variable patterns like "{{ $varname }}"
var variableRx = regexp.MustCompile(`{{\s*\$(\w+)\s*}}`)

// Variables represents a storage for variable names and their values
// Used for variable substitution in test descriptions, paths, queries, headers, requests, responses, etc.
type Variables struct {
	variables map[string]string
}

// New creates a new instance of Variables with an empty variables map
func New() *Variables {
	return &Variables{
		variables: make(map[string]string),
	}
}

// Set adds a new variable (or replaces an existing one) to the variables map
func (vs *Variables) Set(name, value string) {
	vs.variables[name] = value
}

// Merge adds new variables and replaces values of existing ones from a provided map
// Useful when applying variables from multiple sources (e.g., test cases, environment)
func (vs *Variables) Merge(variables map[string]string) {
	for n, v := range variables {
		vs.variables[n] = v
	}
}

// Len returns the number of variables currently stored in the variables map
func (vs *Variables) Len() int {
	return len(vs.variables)
}

// Substitute replaces all variable references in a string with their actual values
// Variables are specified in the format "{{ $varname }}" in the input string
// The method first looks for the variable in the internal variables map,
// and if not found, checks environment variables
// Returns the input string with all recognized variables replaced with their values
func (vs *Variables) Substitute(s string) string {
	return variableRx.ReplaceAllStringFunc(s, func(found string) string {
		name := getVarName(found)
		if name == "" {
			return found
		}
		if val, ok := vs.get(name); ok {
			return val
		}
		return found
	})
}

func (vs *Variables) get(name string) (string, bool) {
	val, ok := vs.variables[name]
	if ok {
		return val, ok
	}

	return os.LookupEnv(name)
}

func getVarName(part string) string {
	part = strings.TrimSpace(part[2 : len(part)-2])
	part = part[1:]
	return part
}
