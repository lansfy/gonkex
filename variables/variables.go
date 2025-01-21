package variables

import (
	"regexp"

	"github.com/lansfy/gonkex/models"
)

type Variables struct {
	variables map[string]*Variable
}

var variableRx = regexp.MustCompile(`{{\s*\$(\w+)\s*}}`)

func New() *Variables {
	return &Variables{
		variables: make(map[string]*Variable),
	}
}

// Load adds new variables and replaces values of existing
func (vs *Variables) Load(variables map[string]string) {
	for n, v := range variables {
		vs.variables[n] = NewVariable(n, v)
	}
}

// Load adds new variables and replaces values of existing
func (vs *Variables) Set(name, value string) {
	vs.variables[name] = NewVariable(name, value)
}

func (vs *Variables) Apply(t models.TestInterface) models.TestInterface {
	newTest := t.Clone()
	if vs != nil {
		newTest.ApplyVariables(vs.perform)
	}
	return newTest
}

// Merge adds given variables to set or overrides existed
func (vs *Variables) Merge(vars *Variables) {
	for k, v := range vars.variables {
		vs.variables[k] = v
	}
}

func (vs *Variables) Len() int {
	return len(vs.variables)
}

func usedVariables(str string) (res []string) {
	matches := variableRx.FindAllStringSubmatch(str, -1)
	for _, match := range matches {
		res = append(res, match[1])
	}

	return res
}

// perform replaces all variables in str to their values
// and returns result string
func (vs *Variables) perform(str string) string {
	varNames := usedVariables(str)

	for _, k := range varNames {
		if v := vs.get(k); v != nil {
			str = v.Perform(str)
		}
	}

	return str
}

func (vs *Variables) get(name string) *Variable {
	if v := vs.variables[name]; v != nil {
		return v
	}
	return NewFromEnvironment(name)
}

func (vs *Variables) Add(v *Variable) *Variables {
	vs.variables[v.name] = v
	return vs
}
