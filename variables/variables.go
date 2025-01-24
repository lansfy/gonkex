package variables

import (
	"os"
	"regexp"
	"strings"
)

var variableRx = regexp.MustCompile(`{{\s*\$(\w+)\s*}}`)

type Variables struct {
	variables map[string]string
}

func New() *Variables {
	return &Variables{
		variables: make(map[string]string),
	}
}

// Set adds new variable (or replace existing)
func (vs *Variables) Set(name, value string) {
	vs.variables[name] = value
}

// Load adds new variables and replaces values of existing
func (vs *Variables) Load(variables map[string]string) {
	for n, v := range variables {
		vs.variables[n] = v
	}
}

// Merge adds given variables to set or overrides existed
func (vs *Variables) Merge(vars *Variables) {
	vs.Load(vars.variables)
}

// Len returns number of variables in storage
func (vs *Variables) Len() int {
	return len(vs.variables)
}

// Substitute replaces all variables in str to their values and returns result string
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
