package yaml_file

import (
	"github.com/lansfy/gonkex/compare"
)

type TestDefinition struct {
	Name                     string                    `json:"name" yaml:"name"`
	Description              string                    `json:"description" yaml:"description"`
	Status                   StatusEnum                `json:"status" yaml:"status"`
	Variables                map[string]string         `json:"variables" yaml:"variables"`
	VariablesToSet           VariablesToSet            `json:"variables_to_set" yaml:"variables_to_set"`
	Form                     *Form                     `json:"form" yaml:"form"`
	Method                   string                    `json:"method" yaml:"method"`
	RequestURL               string                    `json:"path" yaml:"path"`
	QueryParams              string                    `json:"query" yaml:"query"`
	RequestTmpl              string                    `json:"request" yaml:"request"`
	ResponseTmpls            map[int]string            `json:"response" yaml:"response"`
	ResponseHeaders          map[int]map[string]string `json:"responseHeaders" yaml:"responseHeaders"`
	BeforeScriptParams       scriptParams              `json:"beforeScript" yaml:"beforeScript"`
	AfterRequestScriptParams scriptParams              `json:"afterRequestScript" yaml:"afterRequestScript"`
	HeadersVal               map[string]string         `json:"headers" yaml:"headers"`
	CookiesVal               map[string]string         `json:"cookies" yaml:"cookies"`
	Cases                    []CaseData                `json:"cases" yaml:"cases"`
	ComparisonParams         compare.Params            `json:"comparisonParams" yaml:"comparisonParams"`
	FixtureFiles             []string                  `json:"fixtures" yaml:"fixtures"`
	MocksDefinition          map[string]interface{}    `json:"mocks" yaml:"mocks"`
	PauseValue               Duration                  `json:"pause" yaml:"pause"`
	AfterRequestPauseValue   Duration                  `json:"afterRequestPause" yaml:"afterRequestPause"`
	DbQueryTmpl              string                    `json:"dbQuery" yaml:"dbQuery"`
	DbResponseTmpl           []string                  `json:"dbResponse" yaml:"dbResponse"`
	DatabaseChecks           []DatabaseCheck           `json:"dbChecks" yaml:"dbChecks"`
	RetryPolicy              retryPolicy               `json:"retryPolicy" yaml:"retryPolicy"`
	Meta                     map[string]interface{}    `json:"meta" yaml:"meta"`
}

type CaseData struct {
	Name                   string                         `json:"name" yaml:"name"`
	Description            string                         `json:"description" yaml:"description"`
	RequestArgs            map[string]interface{}         `json:"requestArgs" yaml:"requestArgs"`
	ResponseArgs           map[int]map[string]interface{} `json:"responseArgs" yaml:"responseArgs"`
	BeforeScriptArgs       map[string]interface{}         `json:"beforeScriptArgs" yaml:"beforeScriptArgs"`
	AfterRequestScriptArgs map[string]interface{}         `json:"afterRequestScriptArgs" yaml:"afterRequestScriptArgs"`
	DbQueryArgs            map[string]interface{}         `json:"dbQueryArgs" yaml:"dbQueryArgs"`
	DbResponseArgs         map[string]interface{}         `json:"dbResponseArgs" yaml:"dbResponseArgs"`
	DbResponse             []string                       `json:"dbResponse" yaml:"dbResponse"`
	Variables              map[string]interface{}         `json:"variables" yaml:"variables"`
}

type Form struct {
	Files  map[string]string `json:"files" yaml:"files"`
	Fields map[string]string `json:"fields" yaml:"fields"`
}

type DatabaseCheck struct {
	DbQueryTmpl      string         `json:"dbQuery" yaml:"dbQuery"`
	DbResponseTmpl   []string       `json:"dbResponse" yaml:"dbResponse"`
	ComparisonParams compare.Params `json:"comparisonParams" yaml:"comparisonParams"`
}

type retryPolicy struct {
	Attempts     int      `json:"attempts" yaml:"attempts"`
	Delay        Duration `json:"delay" yaml:"delay"`
	SuccessCount int      `json:"successInRow" yaml:"successInRow"`
}

type scriptParams struct {
	PathTmpl string   `json:"path" yaml:"path"`
	Timeout  Duration `json:"timeout" yaml:"timeout"`
}

type VariablesToSet map[int]map[string]string

/*
There can be two types of data in yaml-file:
 1. JSON-paths:
    VariablesToSet:
    <code1>:
    <varName1>: <JSON_Path1>
    <varName2>: <JSON_Path2>
 2. Plain text:
    VariablesToSet:
    <code1>: <varName1>
    <code2>: <varName2>
    ...
    In this case we unmarshall values to format similar to JSON-paths format with empty paths:
    VariablesToSet:
    <code1>:
    <varName1>: ""
    <code2>:
    <varName2>: ""
*/
func (v *VariablesToSet) UnmarshalYAML(unmarshal func(interface{}) error) error {
	res := map[int]map[string]string{}

	// try to unmarshall as plain text
	var plain map[int]string
	if err := unmarshal(&plain); err == nil {
		for code, varName := range plain {
			res[code] = map[string]string{
				varName: "",
			}
		}

		*v = res
		return nil
	}

	// json-paths
	if err := unmarshal(&res); err != nil {
		return err
	}

	*v = res
	return nil
}
