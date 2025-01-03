package yaml_file

import (
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

func (t *Test) ToJSON() ([]byte, error) {
	return []byte(t.Request), nil
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

func (t *Test) GetStatus() string {
	return t.Status
}

func (t *Test) Fixtures() []string {
	return t.FixtureFiles
}

func (t *Test) ServiceMocks() map[string]interface{} {
	return t.MocksDefinition
}

func (t *Test) Pause() int {
	return t.PauseValue
}

func (t *Test) BeforeScriptPath() string {
	return t.BeforeScript
}

func (t *Test) BeforeScriptTimeout() int {
	return t.BeforeScriptParams.Timeout
}

func (t *Test) AfterRequestScriptPath() string {
	return t.AfterRequestScript
}

func (t *Test) AfterRequestScriptTimeout() int {
	return t.AfterRequestScriptParams.Timeout
}

func (t *Test) Cookies() map[string]string {
	return t.CookiesVal
}

func (t *Test) Headers() map[string]string {
	return t.HeadersVal
}

// TODO: it might make sense to do support of case-insensitive checking
func (t *Test) ContentType() string {
	return t.HeadersVal["Content-Type"]
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
	return &res
}

func (t *Test) SetStatus(status string) {
	t.Status = status
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
