package yaml_file

import (
	"bytes"
	"fmt"
	"os"
	"regexp"
	"strings"
	"text/template"

	"github.com/lansfy/gonkex/models"

	"gopkg.in/yaml.v2"
)

const (
	gonkexVariableLeftPart  = "{{ $"
	gonkexProtectSubstitute = "!protect!"
)

var gonkexProtectTemplate = regexp.MustCompile(`{{\s*\$`)

func parseTestDefinitionContent(f FileReadFun, absPath string, data []byte) ([]models.TestInterface, error) {
	testDefinitions, err := f(absPath, data)
	if err != nil {
		return nil, err
	}

	tests := []*testImpl{}
	for i := range testDefinitions {
		testCases, err := makeTestFromDefinition(absPath, &testDefinitions[i])
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

// DefaultFileRead reads and unmarshals the YAML content of a test definition file (default implementation).
//
// It takes the file path (used only for error reporting) and the raw content in bytes,
// then attempts to strictly unmarshal the content into a slice of TestDefinition structs.
//
// Returns the parsed test definitions or an error if unmarshalling fails.
func DefaultFileRead(filePath string, content []byte) ([]TestDefinition, error) {
	testDefinitions := []TestDefinition{}

	// reading the test source file
	if err := yaml.UnmarshalStrict(content, &testDefinitions); err != nil {
		return nil, fmt.Errorf("unmarshal file %s: %w", filePath, err)
	}

	return testDefinitions, nil
}

func parseTestDefinitionFile(f FileReadFun, absPath string) ([]models.TestInterface, error) {
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
	// test definition has no cases, so using request/response as is
	if len(testDefinition.Cases) == 0 {
		test := makeOneTest(filePath, testDefinition)
		return []*testImpl{test}, nil
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
		test.RequestURL, err = substituteArgs(testDefinition.RequestURL, testCase.RequestArgs)
		if err != nil {
			return nil, err
		}

		test.Request, err = substituteArgs(testDefinition.RequestTmpl, testCase.RequestArgs)
		if err != nil {
			return nil, err
		}

		test.QueryParams, err = substituteArgs(testDefinition.QueryParams, testCase.RequestArgs)
		if err != nil {
			return nil, err
		}

		test.HeadersVal, err = substituteArgsToMap(testDefinition.HeadersVal, testCase.RequestArgs)
		if err != nil {
			return nil, err
		}

		test.CookiesVal, err = substituteArgsToMap(testDefinition.CookiesVal, testCase.RequestArgs)
		if err != nil {
			return nil, err
		}

		// substitute ResponseArgs to different parts of response
		test.Responses = map[int]string{}
		for status, tpl := range testDefinition.ResponseTmpls {
			args, ok := testCase.ResponseArgs[status]
			if ok {
				// found args for response status
				test.Responses[status], err = substituteArgs(tpl, args)
				if err != nil {
					return nil, err
				}
			} else {
				// not found args, using response as is
				test.Responses[status] = tpl
			}
		}

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

		test.BeforeScriptPath, err = substituteArgs(testDefinition.BeforeScriptParams.PathTmpl, testCase.BeforeScriptArgs)
		if err != nil {
			return nil, err
		}

		test.AfterRequestScriptPath, err = substituteArgs(testDefinition.AfterRequestScriptParams.PathTmpl, testCase.AfterRequestScriptArgs)
		if err != nil {
			return nil, err
		}

		for key, value := range testCase.Variables {
			var ok bool
			combinedVariables[key], ok = value.(string)
			if !ok {
				return nil, fmt.Errorf("variable %q has non-string value %v", key, value)
			}
		}
		test.CombinedVariables = cloneVariables(combinedVariables)

		test.DbChecks, err = readDatabaseCheck(testDefinition, &testCase)
		if err != nil {
			return nil, err
		}
		tests = append(tests, test)
	}

	return tests, nil
}

func makeOneTest(filePath string, testDefinition *TestDefinition) *testImpl {
	test := &testImpl{
		TestDefinition:         *testDefinition,
		Filename:               filePath,
		Request:                testDefinition.RequestTmpl,
		Responses:              testDefinition.ResponseTmpls,
		ResponseHeaders:        testDefinition.ResponseHeaders,
		BeforeScriptPath:       testDefinition.BeforeScriptParams.PathTmpl,
		AfterRequestScriptPath: testDefinition.AfterRequestScriptParams.PathTmpl,
		CombinedVariables:      testDefinition.Variables,
	}

	dbChecks := []models.DatabaseCheck{}
	if testDefinition.DbQueryTmpl != "" {
		// old style db checks
		dbChecks = append(dbChecks, &dbCheck{
			query:    testDefinition.DbQueryTmpl,
			response: testDefinition.DbResponseTmpl,
		})
	}
	for _, check := range testDefinition.DatabaseChecks {
		dbChecks = append(dbChecks, &dbCheck{
			query:    check.DbQueryTmpl,
			response: check.DbResponseTmpl,
			params:   check.ComparisonParams,
		})
	}
	test.DbChecks = dbChecks
	return test
}

func readDatabaseCheck(def *TestDefinition, testCase *CaseData) ([]models.DatabaseCheck, error) {
	tmpDbQuery, err := substituteArgs(def.DbQueryTmpl, testCase.DbQueryArgs)
	if err != nil {
		return nil, err
	}

	// compile DbResponse
	tmpDbResponse := []string{}
	if testCase.DbResponse != nil {
		// DbResponse from test case has top priority
		tmpDbResponse = testCase.DbResponse
	} else {
		if len(def.DbResponseTmpl) != 0 {
			// compile DbResponse string by string
			for _, tpl := range def.DbResponseTmpl {
				dbResponseString, err := substituteArgs(tpl, testCase.DbResponseArgs)
				if err != nil {
					return nil, err
				}
				tmpDbResponse = append(tmpDbResponse, dbResponseString)
			}
		} else {
			tmpDbResponse = def.DbResponseTmpl
		}
	}

	dbChecks := []models.DatabaseCheck{}
	if tmpDbQuery != "" {
		dbChecks = append(dbChecks, &dbCheck{
			query:    tmpDbQuery,
			response: tmpDbResponse,
		})
	}

	for _, check := range def.DatabaseChecks {
		query, err := substituteArgs(check.DbQueryTmpl, testCase.DbQueryArgs)
		if err != nil {
			return nil, err
		}

		c := &dbCheck{
			query:  query,
			params: check.ComparisonParams,
		}
		for _, tpl := range check.DbResponseTmpl {
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
