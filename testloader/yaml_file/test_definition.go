package yaml_file

import (
	"github.com/lansfy/gonkex/compare"
)

type TestDefinition struct {
	Name               string                    `json:"name" yaml:"name"`
	Description        string                    `json:"description" yaml:"description"`
	Status             StatusEnum                `json:"status" yaml:"status"`
	Variables          map[string]string         `json:"variables" yaml:"variables"`
	VariablesToSet     VariablesToSet            `json:"variables_to_set" yaml:"variables_to_set"`
	Form               *Form                     `json:"form" yaml:"form"`
	Method             string                    `json:"method" yaml:"method"`
	Path               string                    `json:"path" yaml:"path"`
	Query              string                    `json:"query" yaml:"query"`
	Request            string                    `json:"request" yaml:"request"`
	Response           map[int]string            `json:"response" yaml:"response"`
	ResponseHeaders    map[int]map[string]string `json:"responseHeaders" yaml:"responseHeaders"`
	BeforeScript       ScriptParams              `json:"beforeScript" yaml:"beforeScript"`
	AfterRequestScript ScriptParams              `json:"afterRequestScript" yaml:"afterRequestScript"`
	Headers            map[string]string         `json:"headers" yaml:"headers"`
	Cookies            map[string]string         `json:"cookies" yaml:"cookies"`
	Cases              []CaseData                `json:"cases" yaml:"cases"`
	ComparisonParams   compare.Params            `json:"comparisonParams" yaml:"comparisonParams"`
	Fixtures           []string                  `json:"fixtures" yaml:"fixtures"`
	Mocks              map[string]interface{}    `json:"mocks" yaml:"mocks"`
	MocksParams        MocksParams               `json:"mocksParams" yaml:"mocksParams"`
	Pause              Duration                  `json:"pause" yaml:"pause"`
	AfterRequestPause  Duration                  `json:"afterRequestPause" yaml:"afterRequestPause"`
	DbQuery            string                    `json:"dbQuery" yaml:"dbQuery"`
	DbResponse         []string                  `json:"dbResponse" yaml:"dbResponse"`
	DbChecks           []DatabaseCheck           `json:"dbChecks" yaml:"dbChecks"`
	RetryPolicy        RetryPolicy               `json:"retryPolicy" yaml:"retryPolicy"`
	Meta               map[string]interface{}    `json:"meta" yaml:"meta"`
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
	Variables              map[string]string              `json:"variables" yaml:"variables"`
}

type Form struct {
	Files  map[string]string `json:"files" yaml:"files"`
	Fields map[string]string `json:"fields" yaml:"fields"`
}

type DatabaseCheck struct {
	DbQuery          string         `json:"dbQuery" yaml:"dbQuery"`
	DbResponse       []string       `json:"dbResponse" yaml:"dbResponse"`
	ComparisonParams compare.Params `json:"comparisonParams" yaml:"comparisonParams"`
}

type RetryPolicy struct {
	Attempts     int      `json:"attempts" yaml:"attempts"`
	Delay        Duration `json:"delay" yaml:"delay"`
	SuccessInRow int      `json:"successInRow" yaml:"successInRow"`
}

type ScriptParams struct {
	Path    string   `json:"path" yaml:"path"`
	Timeout Duration `json:"timeout" yaml:"timeout"`
}

type MocksParams struct {
	ShareState bool `json:"shareState" yaml:"shareState"`
}
