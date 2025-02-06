package yaml_file

import (
	"strings"
	"time"

	"github.com/lansfy/gonkex/compare"
	"github.com/lansfy/gonkex/models"
)

type dbCheck struct {
	query    string
	response []string
	params   compare.Params
}

func (c *dbCheck) DbQueryString() string {
	return c.query
}

func (c *dbCheck) DbResponseJson() []string {
	return c.response
}

func (c *dbCheck) GetComparisonParams() models.ComparisonParams {
	return &cmpParams{c.params}
}

type cmpParams struct {
	params compare.Params
}

func (c *cmpParams) IgnoreValuesChecking() bool {
	return c.params.IgnoreValues
}

func (c *cmpParams) IgnoreArraysOrdering() bool {
	return c.params.IgnoreArraysOrdering
}

func (c *cmpParams) DisallowExtraFields() bool {
	return c.params.DisallowExtraFields
}

type retry struct {
	params retryPolicy
}

func (r *retry) Attempts() int {
	return r.params.Attempts
}

func (r *retry) Delay() time.Duration {
	return r.params.Delay.Duration
}

func (r *retry) SuccessCount() int {
	return r.params.SuccessCount
}

type Test struct {
	TestDefinition

	Filename string

	Request            string
	Responses          map[int]string
	ResponseHeaders    map[int]map[string]string
	BeforeScript       string
	AfterRequestScript string

	CombinedVariables map[string]string

	DbChecks []models.DatabaseCheck

	FirstTest   bool
	LastTest    bool
	IsOneOfCase bool
}

func (t *Test) ToQuery() string {
	return t.QueryParams
}

func (t *Test) GetMethod() string {
	return t.Method
}

func (t *Test) Path() string {
	return t.RequestURL
}

func (t *Test) GetRequest() string {
	return t.Request
}

func (t *Test) GetResponses() map[int]string {
	return t.Responses
}

func (t *Test) GetResponse(code int) (string, bool) {
	val, ok := t.Responses[code]
	return val, ok
}

func (t *Test) GetResponseHeaders(code int) (map[string]string, bool) {
	val, ok := t.ResponseHeaders[code]

	return val, ok
}

func (t *Test) GetName() string {
	return t.Name
}

func (t *Test) GetDescription() string {
	return t.Description
}

func (t *Test) GetStatus() models.Status {
	return t.Status.value
}

func (t *Test) Fixtures() []string {
	return t.FixtureFiles
}

func (t *Test) GetMeta() map[string]interface{} {
	return t.Meta
}

func (t *Test) ServiceMocks() map[string]interface{} {
	return t.MocksDefinition
}

func (t *Test) Pause() time.Duration {
	return t.PauseValue.Duration
}

func (t *Test) AfterRequestPause() time.Duration {
	return t.AfterRequestPauseValue.Duration
}

func (t *Test) BeforeScriptPath() string {
	return t.BeforeScript
}

func (t *Test) BeforeScriptTimeout() time.Duration {
	return t.BeforeScriptParams.Timeout.Duration
}

func (t *Test) AfterRequestScriptPath() string {
	return t.AfterRequestScript
}

func (t *Test) AfterRequestScriptTimeout() time.Duration {
	return t.AfterRequestScriptParams.Timeout.Duration
}

func (t *Test) Cookies() map[string]string {
	return t.CookiesVal
}

func (t *Test) Headers() map[string]string {
	return t.HeadersVal
}

func (t *Test) GetRetryPolicy() models.RetryPolicy {
	return &retry{t.RetryPolicy}
}

func (t *Test) ContentType() string {
	for key, val := range t.HeadersVal {
		if strings.ToLower(key) == "content-type" {
			return val
		}
	}
	return ""
}

func (t *Test) GetComparisonParams() models.ComparisonParams {
	return &cmpParams{t.ComparisonParams}
}

func (t *Test) GetDatabaseChecks() []models.DatabaseCheck {
	return t.DbChecks
}

func (t *Test) GetVariables() map[string]string {
	return t.Variables
}

func (t *Test) GetCombinedVariables() map[string]string {
	return t.CombinedVariables
}

func (t *Test) GetForm() *models.Form {
	return t.Form
}

func (t *Test) GetVariablesToSet() map[int]map[string]string {
	return t.VariablesToSet
}

func (t *Test) GetFileName() string {
	return t.Filename
}

func (t *Test) Clone() models.TestInterface {
	res := *t
	if t.MocksDefinition != nil {
		res.MocksDefinition = map[string]interface{}{}
		for s := range t.MocksDefinition {
			res.MocksDefinition[s] = deepClone(t.MocksDefinition[s])
		}
	}
	return &res
}

func (t *Test) SetStatus(status models.Status) {
	t.Status.value = status
}

func (t *Test) ApplyVariables(perform func(string) string) {
	t.QueryParams = performQuery(t.QueryParams, perform)
	t.Method = perform(t.Method)
	t.RequestURL = perform(t.RequestURL)
	t.Request = perform(t.Request)

	dbChecks := []models.DatabaseCheck{}
	for _, def := range t.GetDatabaseChecks() {
		cmpOptions := def.GetComparisonParams()
		newCheck := &dbCheck{
			query:    perform(def.DbQueryString()),
			response: performDbResponses(def.DbResponseJson(), perform),
			params: compare.Params{
				IgnoreValues:         cmpOptions.IgnoreValuesChecking(),
				IgnoreArraysOrdering: cmpOptions.IgnoreArraysOrdering(),
				DisallowExtraFields:  cmpOptions.DisallowExtraFields(),
			},
		}
		dbChecks = append(dbChecks, newCheck)
	}
	t.DbChecks = dbChecks

	t.Responses = performResponses(t.Responses, perform)
	t.HeadersVal = performHeaders(t.HeadersVal, perform)

	if t.Form != nil {
		t.Form = performForm(t.Form, perform)
	}

	for _, definition := range t.ServiceMocks() {
		performInterface(definition, perform)
	}
}

func (t *Test) FirstTestInFile() bool {
	return t.FirstTest
}

func (t *Test) LastTestInFile() bool {
	return t.LastTest
}

func (t *Test) OneOfCase() bool {
	return t.IsOneOfCase
}
