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

func parseTestDefinitionFile(absPath string) ([]Test, error) {
	data, err := os.ReadFile(absPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %s:\n%s", absPath, err)
	}

	var testDefinitions []TestDefinition

	// reading the test source file
	if err := yaml.UnmarshalStrict(data, &testDefinitions); err != nil {
		return nil, fmt.Errorf("failed to unmarshal %s:\n%s", absPath, err)
	}

	var tests []Test

	for i := range testDefinitions {
		testCases, err := makeTestFromDefinition(absPath, &testDefinitions[i])
		if err != nil {
			return nil, err
		}

		tests = append(tests, testCases...)
	}

	if len(tests) != 0 {
		tests[0].FirstTest = true
		tests[len(tests)-1].LastTest = true
	}

	return tests, nil
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
	res := make(map[string]string)
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
func makeTestFromDefinition(filePath string, testDefinition *TestDefinition) ([]Test, error) {
	var tests []Test

	// test definition has no cases, so using request/response as is
	if len(testDefinition.Cases) == 0 {
		test := Test{
			TestDefinition: *testDefinition,
			Filename:       filePath,
		}
		test.Description = testDefinition.Description
		test.Request = testDefinition.RequestTmpl
		test.Responses = testDefinition.ResponseTmpls
		test.ResponseHeaders = testDefinition.ResponseHeaders
		test.BeforeScript = testDefinition.BeforeScriptParams.PathTmpl
		test.AfterRequestScript = testDefinition.AfterRequestScriptParams.PathTmpl
		test.CombinedVariables = testDefinition.Variables

		dbChecks := []models.DatabaseCheck{}
		if testDefinition.DbQueryTmpl != "" {
			dbChecks = append(dbChecks, &dbCheck{query: testDefinition.DbQueryTmpl, response: testDefinition.DbResponseTmpl})
		}
		for _, check := range testDefinition.DatabaseChecks {
			dbChecks = append(dbChecks, &dbCheck{
				query:    check.DbQueryTmpl,
				response: check.DbResponseTmpl,
				params:   check.ComparisonParams,
			})
		}
		test.DbChecks = dbChecks

		return append(tests, test), nil
	}

	var err error

	combinedVariables := map[string]string{}

	if testDefinition.Variables != nil {
		combinedVariables = testDefinition.Variables
	}

	// produce as many tests as cases defined
	for caseIdx, testCase := range testDefinition.Cases {
		test := Test{
			TestDefinition: *testDefinition,
			Filename:       filePath,
		}
		test.Name = fmt.Sprintf("%s #%d", test.Name, caseIdx+1)

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
		test.Responses = make(map[int]string)
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

		test.ResponseHeaders = make(map[int]map[string]string)
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

		test.BeforeScript, err = substituteArgs(testDefinition.BeforeScriptParams.PathTmpl, testCase.BeforeScriptArgs)
		if err != nil {
			return nil, err
		}

		test.AfterRequestScript, err = substituteArgs(testDefinition.AfterRequestScriptParams.PathTmpl, testCase.AfterRequestScriptArgs)
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

		var tmpDbQuery string
		var tmpDbResponse []string

		tmpDbQuery, err = substituteArgs(testDefinition.DbQueryTmpl, testCase.DbQueryArgs)
		if err != nil {
			return nil, err
		}

		// compile DbResponse
		if testCase.DbResponse != nil {
			// DbResponse from test case has top priority
			tmpDbResponse = testCase.DbResponse
		} else {
			if len(testDefinition.DbResponseTmpl) != 0 {
				// compile DbResponse string by string
				for _, tpl := range testDefinition.DbResponseTmpl {
					dbResponseString, err := substituteArgs(tpl, testCase.DbResponseArgs)
					if err != nil {
						return nil, err
					}
					tmpDbResponse = append(tmpDbResponse, dbResponseString)
				}
			} else {
				tmpDbResponse = testDefinition.DbResponseTmpl
			}
		}

		dbChecks := []models.DatabaseCheck{}
		if tmpDbQuery != "" {
			dbChecks = append(dbChecks, &dbCheck{
				query:    tmpDbQuery,
				response: tmpDbResponse,
			})
		}

		for _, check := range testDefinition.DatabaseChecks {
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

		test.DbChecks = dbChecks

		tests = append(tests, test)
	}

	return tests, nil
}

func cloneVariables(s map[string]string) map[string]string {
	clone := map[string]string{}
	for k, v := range s {
		clone[k] = v
	}
	return clone
}
