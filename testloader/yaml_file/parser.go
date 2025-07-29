package yaml_file

import (
	"bytes"
	"fmt"
	"os"
	"regexp"
	"strings"
	"text/template"

	"github.com/lansfy/gonkex/models"

	"gopkg.in/yaml.v3"
)

const (
	gonkexVariableLeftPart  = "{{ $"
	gonkexProtectSubstitute = "!protect!"
)

var gonkexProtectTemplate = regexp.MustCompile(`{{\s*\$`)

func parseTestDefinitionContent(f FileParseFun, absPath string, data []byte) ([]models.TestInterface, error) {
	testDefinitions, err := f(absPath, data)
	if err != nil {
		return nil, err
	}

	tests := []*testImpl{}
	for _, item := range testDefinitions {
		testCases, err := makeTestFromDefinition(absPath, item)
		if err != nil {
			return nil, fmt.Errorf("preprocess file %s: %w", absPath, err)
		}

		tests = append(tests, testCases...)
	}

	if len(tests) != 0 {
		tests[0].FirstTest = true
		tests[len(tests)-1].LastTest = true
	}

	result := make([]models.TestInterface, len(tests))
	for i := range tests {
		result[i] = tests[i]
	}

	return result, nil
}

// DefaultFileParse reads and unmarshals the YAML content of a test definition file (default implementation).
//
// It takes the file path (used only for error reporting) and the raw content in bytes,
// then attempts to strictly unmarshal the content into a slice of TestDefinition structs.
//
// Returns the parsed test definitions or an error if unmarshalling fails.
func DefaultFileParse(filePath string, content []byte) ([]*TestDefinition, error) {
	decoder := yaml.NewDecoder(bytes.NewReader(content))
	decoder.KnownFields(true)

	// reading the test source file
	testDefinitions := []*TestDefinition{}
	if err := decoder.Decode(&testDefinitions); err != nil {
		return nil, fmt.Errorf("unmarshal file %s: %w", filePath, err)
	}

	return testDefinitions, nil
}

func parseTestDefinitionFile(f FileParseFun, absPath string) ([]models.TestInterface, error) {
	data, err := os.ReadFile(absPath)
	if err != nil {
		return nil, fmt.Errorf("read file %s: %w", absPath, err)
	}

	moreTests, err := parseTestDefinitionContent(f, absPath, data)
	if err != nil {
		return nil, err
	}

	return moreTests, nil
}

func substituteArgs(tmpl string, args map[string]interface{}) (string, error) {
	tmpl = gonkexProtectTemplate.ReplaceAllString(tmpl, gonkexProtectSubstitute)

	compiledTmpl, err := template.New("").Parse(tmpl)
	if err != nil {
		return "", err
	}

	buf := &bytes.Buffer{}

	if err := compiledTmpl.Execute(buf, args); err != nil {
		return "", err
	}

	tmpl = strings.ReplaceAll(buf.String(), gonkexProtectSubstitute, gonkexVariableLeftPart)

	return tmpl, nil
}

func substituteArgsToMap(tmpl map[string]string, args map[string]interface{}) (map[string]string, error) {
	res := map[string]string{}
	for key, value := range tmpl {
		var err error
		res[key], err = substituteArgs(value, args)
		if err != nil {
			return nil, err
		}
	}

	return res, nil
}

// Make tests from the given test definition.
func makeTestFromDefinition(filePath string, testDefinition *TestDefinition) ([]*testImpl, error) {
	if testDefinition.DbQuery != "" && len(testDefinition.DbChecks) != 0 {
		return nil, fmt.Errorf("mix of old dbQuery/dbResponse and dbChecks in one test forbidden")
	}

	// test definition has no cases, so using request/response as is
	if len(testDefinition.Cases) == 0 {
		return makeOneTest(filePath, testDefinition)
	}

	combinedVariables := map[string]string{}
	if testDefinition.Variables != nil {
		combinedVariables = testDefinition.Variables
	}

	var err error
	tests := []*testImpl{}

	// produce as many tests as cases defined
	for caseIdx, testCase := range testDefinition.Cases {
		test := &testImpl{
			TestDefinition: *testDefinition,
			Filename:       filePath,
		}

		postfix := ""
		if testCase.Name != "" {
			postfix = " (" + testCase.Name + ")"
		}

		test.Name = fmt.Sprintf("%s #%d%s", test.Name, caseIdx+1, postfix)
		test.IsOneOfCase = true

		if testCase.Description != "" {
			test.Description = testCase.Description
		}

		// substitute RequestArgs to different parts of request
		test.TestDefinition.Path, err = substituteArgs(testDefinition.Path, testCase.RequestArgs)
		if err != nil {
			return nil, err
		}

		test.Request, err = substituteArgs(testDefinition.Request, testCase.RequestArgs)
		if err != nil {
			return nil, err
		}

		test.Query, err = substituteArgs(testDefinition.Query, testCase.RequestArgs)
		if err != nil {
			return nil, err
		}

		test.TestDefinition.Headers, err = substituteArgsToMap(testDefinition.Headers, testCase.RequestArgs)
		if err != nil {
			return nil, err
		}

		test.TestDefinition.Cookies, err = substituteArgsToMap(testDefinition.Cookies, testCase.RequestArgs)
		if err != nil {
			return nil, err
		}

		// substitute ResponseArgs to different parts of response
		responses := map[int]string{}
		for status, tpl := range testDefinition.Response {
			args, ok := testCase.ResponseArgs[status]
			if ok {
				// found args for response status
				responses[status], err = substituteArgs(tpl, args)
				if err != nil {
					return nil, err
				}
			} else {
				// not found args, using response as is
				responses[status] = tpl
			}
		}
		test.Response = responses

		test.ResponseHeaders = map[int]map[string]string{}
		for status, respHeaders := range testDefinition.ResponseHeaders {
			args, ok := testCase.ResponseArgs[status]
			if ok {
				// found args for response status
				test.ResponseHeaders[status], err = substituteArgsToMap(respHeaders, args)
				if err != nil {
					return nil, err
				}
			} else {
				// not found args, using response as is
				test.ResponseHeaders[status] = respHeaders
			}
		}

		test.TestDefinition.BeforeScript.Path, err = substituteArgs(testDefinition.BeforeScript.Path, testCase.BeforeScriptArgs)
		if err != nil {
			return nil, err
		}

		test.TestDefinition.AfterRequestScript.Path, err = substituteArgs(testDefinition.AfterRequestScript.Path, testCase.AfterRequestScriptArgs)
		if err != nil {
			return nil, err
		}

		for key, value := range testCase.Variables {
			combinedVariables[key] = value
		}
		test.CombinedVariables = cloneVariables(combinedVariables)

		if testDefinition.DbQuery != "" {
			test.DbChecks, err = readObsoleteDatabaseCheck(testDefinition, &testCase)
		} else {
			test.DbChecks, err = readDatabaseCheck(testDefinition, &testCase)
		}
		if err != nil {
			return nil, err
		}
		tests = append(tests, test)
	}

	return tests, nil
}

func makeOneTest(filePath string, testDefinition *TestDefinition) ([]*testImpl, error) {
	test := &testImpl{
		TestDefinition:    *testDefinition,
		Filename:          filePath,
		CombinedVariables: testDefinition.Variables,
	}

	dbChecks := []models.DatabaseCheck{}
	if testDefinition.DbQuery != "" {
		// old style db checks
		dbChecks = append(dbChecks, &dbCheck{
			query:    testDefinition.DbQuery,
			response: testDefinition.DbResponse,
		})
	}

	for _, check := range testDefinition.DbChecks {
		dbChecks = append(dbChecks, &dbCheck{
			query:    check.DbQuery,
			response: check.DbResponse,
			params:   check.ComparisonParams,
		})
	}
	test.DbChecks = dbChecks
	return []*testImpl{test}, nil
}

func readObsoleteDatabaseCheck(def *TestDefinition, testCase *CaseData) ([]models.DatabaseCheck, error) {
	tmpDbQuery, err := substituteArgs(def.DbQuery, testCase.DbQueryArgs)
	if err != nil {
		return nil, err
	}

	c := &dbCheck{
		query: tmpDbQuery,
	}

	if testCase.DbResponse != nil {
		// DbResponse from test case has top priority
		c.response = testCase.DbResponse
		return []models.DatabaseCheck{c}, nil
	}

	for _, tpl := range def.DbResponse {
		responseString, err := substituteArgs(tpl, testCase.DbResponseArgs)
		if err != nil {
			return nil, err
		}
		c.response = append(c.response, responseString)
	}
	return []models.DatabaseCheck{c}, nil
}

func readDatabaseCheck(def *TestDefinition, testCase *CaseData) ([]models.DatabaseCheck, error) {
	dbChecks := []models.DatabaseCheck{}
	for _, check := range def.DbChecks {
		query, err := substituteArgs(check.DbQuery, testCase.DbQueryArgs)
		if err != nil {
			return nil, err
		}

		c := &dbCheck{
			query:  query,
			params: check.ComparisonParams,
		}
		for _, tpl := range check.DbResponse {
			responseString, err := substituteArgs(tpl, testCase.DbResponseArgs)
			if err != nil {
				return nil, err
			}

			c.response = append(c.response, responseString)
		}
		dbChecks = append(dbChecks, c)
	}
	return dbChecks, nil
}

func cloneVariables(s map[string]string) map[string]string {
	clone := map[string]string{}
	for k, v := range s {
		clone[k] = v
	}
	return clone
}

func deepClone(src interface{}) interface{} {
	switch v := src.(type) {
	case map[string]interface{}:
		clone := map[string]interface{}{}
		for key, value := range v {
			clone[key] = deepClone(value)
		}
		return clone
	case map[interface{}]interface{}:
		clone := map[interface{}]interface{}{}
		for key, value := range v {
			clone[key] = deepClone(value)
		}
		return clone
	case []interface{}:
		clone := []interface{}{}
		for idx := range v {
			clone = append(clone, deepClone(v[idx]))
		}
		return clone
	default:
		return src
	}
}
