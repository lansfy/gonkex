package variables

import (
	"os"
	"regexp"
	"strings"
)

// variableRx is a regular expression that matches variable patterns like "{{ $varname }}"
var variableRx = regexp.MustCompile(`{{\s*\$(\w+)\s*}}`)

type VariablesImpl struct {
	variables map[string]string
}

// New creates a new default implementation of Variables interface with an empty variables map
func New() Variables {
	return &VariablesImpl{
		variables: make(map[string]string),
	}
}

func (vs *VariablesImpl) Set(name, value string) {
	vs.variables[name] = value
}

func (vs *VariablesImpl) Merge(variables map[string]string) {
	for n, v := range variables {
		vs.variables[n] = v
	}
}

func (vs *VariablesImpl) Len() int {
	return len(vs.variables)
}

func (vs *VariablesImpl) Substitute(s string) string {
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

func (vs *VariablesImpl) get(name string) (string, bool) {
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
