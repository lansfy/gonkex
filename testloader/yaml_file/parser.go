package yaml_file

import (
	"bufio"
	"bytes"
	"errors"
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

func parseTestDefinitionContent(opts *LoaderOpts, absPath string, data []byte) ([]models.TestInterface, error) {
	testDefinitions, err := opts.CustomFileParse(absPath, data)
	if err != nil {
		return nil, err
	}

	wrap := func(err error) error {
		return fmt.Errorf("process '%s': %w", absPath, err)
	}

	tests := []*testImpl{}
	for _, item := range testDefinitions {
		testCases, err := makeTestFromDefinition(opts, absPath, item)
		if err != nil {
			return nil, wrap(err)
		}

		tests = append(tests, testCases...)
	}

	if len(tests) != 0 {
		tests[0].FirstTest = true
		tests[len(tests)-1].LastTest = true
	}

	// process persistent mocks
	err = updateShareState(tests)
	if err != nil {
		return nil, wrap(err)
	}

	result := make([]models.TestInterface, len(tests))
	for i := range tests {
		result[i] = tests[i]
	}

	return result, nil
}

func updateShareState(tests []*testImpl) error {
	var prevItem, curItem *testImpl
	var prevShareState, curShareState bool
	for _, t := range tests {
		prevItem = curItem
		prevShareState = curShareState

		curItem = t
		curShareState = curItem.MocksParams.ShareState
		if !curShareState {
			if prevShareState {
				prevItem.doNotResetMocksAfterTest = false
			}
			continue
		}

		// curShareState == true
		curItem.doNotResetMocksAfterTest = true
		if !prevShareState {
			if len(curItem.Mocks) == 0 {
				return fmt.Errorf("test '%s': shareState require non empty $.mocks declaration", curItem.Name)
			}
			continue
		}

		// curShareState == true && prevShareState == true
		if len(curItem.Mocks) != 0 {
			prevItem.doNotResetMocksAfterTest = false
			continue
		}

		curItem.doNotResetMocksBeforeTest = true
	}

	if curShareState {
		curItem.doNotResetMocksAfterTest = false
	}

	return nil
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

	appendLineNumber(content, testDefinitions)
	return testDefinitions, nil
}

func appendLineNumber(content []byte, defs []*TestDefinition) {
	var counter int
	var linesN []int
	scanner := bufio.NewScanner(bytes.NewReader(content))
	for scanner.Scan() {
		counter++
		if strings.HasPrefix(scanner.Text(), "- name:") {
			linesN = append(linesN, counter)
		}
	}

	if len(defs) != len(linesN) {
		return
	}

	for i := range linesN {
		defs[i].LineNumber = linesN[i]
	}
}

func parseTestDefinitionFile(opts *LoaderOpts, absPath string) ([]models.TestInterface, error) {
	data, err := os.ReadFile(absPath)
	if err != nil {
		return nil, fmt.Errorf("read file %s: %w", absPath, err)
	}

	moreTests, err := parseTestDefinitionContent(opts, absPath, data)
	if err != nil {
		return nil, err
	}

	return moreTests, nil
}

func substituteArgs(opts *LoaderOpts, path, tmpl string, args map[string]interface{}) (string, error) {
	tmpl = gonkexProtectTemplate.ReplaceAllString(tmpl, gonkexProtectSubstitute)

	compiledTmpl, err := template.New("$." + path).Funcs(opts.TemplateFuncs).Parse(tmpl)
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

func substituteArgsToMap(opts *LoaderOpts, path string, tmpl map[string]string,
	args map[string]interface{}) (map[string]string, error) {
	res := map[string]string{}
	for key, value := range tmpl {
		var err error
		res[key], err = substituteArgs(opts, path, value, args)
		if err != nil {
			return nil, err
		}
	}

	return res, nil
}

func hasObsoleteDbCheck(def *TestDefinition) bool {
	return def.DbQuery != "" || def.DbResponse != nil
}

func validateDbChecks(test *testImpl) error {
	for _, item := range test.GetDatabaseChecks() {
		if item.DbQueryString() == "" && len(item.DbResponseJson()) != 0 {
			return errors.New("'dbResponse' found without corresponding 'dbQuery'")
		}
	}
	return nil
}

// Make tests from the given test definition.
func makeTestFromDefinition(opts *LoaderOpts, filePath string, def *TestDefinition) ([]*testImpl, error) {
	wrap := func(err error) error {
		return fmt.Errorf("test '%s': %w", def.Name, err)
	}

	if hasObsoleteDbCheck(def) && len(def.DbChecks) != 0 {
		return nil, wrap(errors.New("mixing old dbQuery/dbResponse with dbChecks in a single test is forbidden"))
	}

	// test definition has no cases, so using request/response as is
	if len(def.Cases) == 0 {
		test := makeOneTest(filePath, def)
		err := validateDbChecks(test)
		if err != nil {
			return nil, wrap(err)
		}
		return []*testImpl{test}, nil
	}

	combinedVariables := map[string]string{}
	if def.Variables != nil {
		combinedVariables = def.Variables
	}

	tests := []*testImpl{}

	// produce as many tests as cases defined
	for caseIdx, testCase := range def.Cases {
		test := &testImpl{
			TestDefinition: *def,
			Filename:       filePath,
			IsOneOfCase:    true,
		}

		test.Name = fmt.Sprintf("%s #%d", test.Name, caseIdx+1)
		if testCase.Name != "" {
			test.Name += " (" + testCase.Name + ")"
		}

		if testCase.Description != "" {
			test.Description = testCase.Description
		}

		wrap = func(err error) error {
			return fmt.Errorf("test '%s': %w", test.Name, err)
		}

		var err error
		// substitute RequestArgs to different parts of request
		test.TestDefinition.Path, err = substituteArgs(opts, "path", def.Path, testCase.RequestArgs)
		if err != nil {
			return nil, wrap(err)
		}

		test.Request, err = substituteArgs(opts, "request", def.Request, testCase.RequestArgs)
		if err != nil {
			return nil, wrap(err)
		}

		test.Query, err = substituteArgs(opts, "query", def.Query, testCase.RequestArgs)
		if err != nil {
			return nil, wrap(err)
		}

		test.TestDefinition.Headers, err = substituteArgsToMap(opts, "headers", def.Headers, testCase.RequestArgs)
		if err != nil {
			return nil, wrap(err)
		}

		test.TestDefinition.Cookies, err = substituteArgsToMap(opts, "cookies", def.Cookies, testCase.RequestArgs)
		if err != nil {
			return nil, wrap(err)
		}

		// substitute ResponseArgs to different parts of response
		responses := map[int]string{}
		for status, tpl := range def.Response {
			args, ok := testCase.ResponseArgs[status]
			if !ok {
				// not found args, using response as is
				responses[status] = tpl
				continue
			}

			// found args for response status
			responses[status], err = substituteArgs(opts, fmt.Sprintf("response.%d", status), tpl, args)
			if err != nil {
				return nil, wrap(err)
			}
		}
		test.Response = responses

		test.ResponseHeaders = map[int]map[string]string{}
		for status, respHeaders := range def.ResponseHeaders {
			args, ok := testCase.ResponseArgs[status]
			if !ok {
				// not found args, using response as is
				test.ResponseHeaders[status] = respHeaders
				continue
			}

			// found args for response status
			test.ResponseHeaders[status], err = substituteArgsToMap(opts,
				fmt.Sprintf("responseHeaders.%d", status), respHeaders, args)
			if err != nil {
				return nil, wrap(err)
			}
		}

		test.TestDefinition.BeforeScript.Path, err = substituteArgs(opts, "beforeScript.path",
			def.BeforeScript.Path, testCase.BeforeScriptArgs)
		if err != nil {
			return nil, wrap(err)
		}

		test.TestDefinition.AfterRequestScript.Path, err = substituteArgs(opts, "afterRequestScript.path",
			def.AfterRequestScript.Path, testCase.AfterRequestScriptArgs)
		if err != nil {
			return nil, wrap(err)
		}

		for key, value := range testCase.Variables {
			combinedVariables[key] = value
		}
		test.CombinedVariables = cloneVariables(combinedVariables)

		if hasObsoleteDbCheck(def) {
			test.DbChecks, err = readObsoleteDatabaseCheck(opts, def, &testCase)
		} else {
			test.DbChecks, err = readDatabaseCheck(opts, def, &testCase)
		}
		if err != nil {
			return nil, wrap(err)
		}

		err = validateDbChecks(test)
		if err != nil {
			return nil, wrap(err)
		}

		tests = append(tests, test)
	}

	return tests, nil
}

func makeOneTest(filePath string, def *TestDefinition) *testImpl {
	test := &testImpl{
		TestDefinition:    *def,
		Filename:          filePath,
		CombinedVariables: def.Variables,
	}

	dbChecks := []models.DatabaseCheck{}
	if hasObsoleteDbCheck(def) {
		// old style db checks
		dbChecks = append(dbChecks, &dbCheck{
			query:    def.DbQuery,
			response: def.DbResponse,
		})
	}

	for _, check := range def.DbChecks {
		dbChecks = append(dbChecks, &dbCheck{
			query:    check.DbQuery,
			response: check.DbResponse,
			params:   check.ComparisonParams,
		})
	}
	test.DbChecks = dbChecks
	return test
}

func readObsoleteDatabaseCheck(opts *LoaderOpts, def *TestDefinition, testCase *CaseData) ([]models.DatabaseCheck, error) {
	tmpDbQuery, err := substituteArgs(opts, "dbQuery", def.DbQuery, testCase.DbQueryArgs)
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
		responseString, err := substituteArgs(opts, "dbResponse", tpl, testCase.DbResponseArgs)
		if err != nil {
			return nil, err
		}
		c.response = append(c.response, responseString)
	}
	return []models.DatabaseCheck{c}, nil
}

func readDatabaseCheck(opts *LoaderOpts, def *TestDefinition, testCase *CaseData) ([]models.DatabaseCheck, error) {
	dbChecks := []models.DatabaseCheck{}
	for idx, check := range def.DbChecks {
		query, err := substituteArgs(opts, fmt.Sprintf("dbChecks[%d].dbQuery", idx),
			check.DbQuery, testCase.DbQueryArgs)
		if err != nil {
			return nil, err
		}

		c := &dbCheck{
			query:  query,
			params: check.ComparisonParams,
		}
		for i, tpl := range check.DbResponse {
			responseString, err := substituteArgs(opts,
				fmt.Sprintf("dbChecks[%d].dbResponse[%d]", idx, i), tpl, testCase.DbResponseArgs)
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
