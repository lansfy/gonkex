package yaml_file

import (
	"bytes"
	"fmt"
	"os"
	"strings"
	"testing"
	"text/template"
	"time"

	"github.com/lansfy/gonkex/models"
	"github.com/lansfy/gonkex/variables"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func TestParser(t *testing.T) {
	testfolder := "testdata/parser/"

	entries, err := os.ReadDir(testfolder)
	require.NoError(t, err)

	for _, v := range entries {
		if !strings.HasSuffix(v.Name(), ".yaml") || !strings.HasSuffix(v.Name(), "_expected.yaml") {
			continue
		}

		expected_result := v.Name()
		input_file := strings.ReplaceAll(expected_result, "_expected.yaml", ".yaml")

		t.Run(expected_result, func(t *testing.T) {
			// read expected result
			content, err := os.ReadFile(testfolder + expected_result)
			require.NoError(t, err)

			var expectedTests []TestInterfaceResult
			decoder := yaml.NewDecoder(bytes.NewBuffer(content))
			decoder.KnownFields(true)
			err = decoder.Decode(&expectedTests)
			require.NoError(t, err)

			expectedError := ""
			if len(expectedTests) == 1 {
				expectedError = expectedTests[0].Error
			}

			// read actual result
			loader := NewFileLoader(testfolder+input_file, &LoaderOpts{
				TemplateFuncs: template.FuncMap{
					"testFun": func(s string) string {
						return "testfun(" + s + ") result"
					},
				},
			})
			loader.SetFilter(testFilter)
			actualTests, err := loader.Load()

			if expectedError != "" {
				require.Error(t, err)
				require.Equal(t, expectedError, err.Error())
				return
			}

			require.NoError(t, err)
			require.Equal(t, len(expectedTests), len(actualTests), "number of generated test doesn't match expected test count for %s", input_file)

			// apply combined variables
			for idx := range actualTests {
				vars := variables.New()
				vars.Merge(actualTests[idx].GetCombinedVariables())
				actualTests[idx].ApplyVariables(vars.Substitute)
			}

			for idx := range expectedTests {
				t.Run(fmt.Sprintf("#%d", idx), func(t *testing.T) {
					compareTestInterface(t, &expectedTests[idx], actualTests[idx])
				})
			}
		})
	}
}

func Test_yamlInMemoryLoader(t *testing.T) {
	// use yamlFileLoader to fill yamlInMemoryLoader
	files := map[string]string{}
	loader := NewFileLoader("testdata/parser/read_from_folder.yaml", &LoaderOpts{
		CustomFileParse: func(filePath string, content []byte) ([]*TestDefinition, error) {
			files[filePath] = string(content)
			return nil, nil
		},
	})
	actualTests, err := loader.Load()
	require.NoError(t, err)
	require.Equal(t, 0, len(actualTests), "this reader shouldn't read any test")

	// read file content with yamlInMemoryLoader
	loader = NewInMemoryLoader(files, &LoaderOpts{})
	loader.SetFilter(testFilter)
	actualTests, err = loader.Load()
	require.NoError(t, err)

	// read expected result
	content, err := os.ReadFile("testdata/parser/read_from_folder_expected.yaml")
	require.NoError(t, err)

	var expectedTests []TestInterfaceResult
	decoder := yaml.NewDecoder(bytes.NewBuffer(content))
	decoder.KnownFields(true)
	err = decoder.Decode(&expectedTests)
	require.NoError(t, err)

	require.Equal(t, len(expectedTests), len(actualTests), "number of generated test doesn't match expected test count")

	for idx := range expectedTests {
		t.Run(fmt.Sprintf("#%d", idx), func(t *testing.T) {
			compareTestInterface(t, &expectedTests[idx], actualTests[idx])
		})
	}
}

func testFilter(fileName string) bool {
	return !strings.Contains(fileName, "file_for_skip_by_filter")
}

type comparisonParamsResult struct {
	IgnoreValuesChecking bool `yaml:"IgnoreValuesChecking"`
	IgnoreArraysOrdering bool `yaml:"IgnoreArraysOrdering"`
	DisallowExtraFields  bool `yaml:"DisallowExtraFields"`
}

type databaseCheckResult struct {
	DbQueryString       string                 `yaml:"DbQueryString"`
	DbResponseJson      []string               `yaml:"DbResponseJson"`
	GetComparisonParams comparisonParamsResult `yaml:"GetComparisonParams"`
}

type retryPolicyResult struct {
	Attempts     int           `yaml:"Attempts"`
	Delay        time.Duration `yaml:"Delay"`
	SuccessCount int           `yaml:"SuccessCount"`
}

type formResult struct {
	GetFiles  map[string]string `yaml:"GetFiles"`
	GetFields map[string]string `yaml:"GetFields"`
}

type scriptResult struct {
	CmdLine string        `yaml:"CmdLine"`
	Timeout time.Duration `yaml:"Timeout"`
}

type mocksResult struct {
	SkipMocksResetBeforeTest bool `yaml:"SkipMocksResetBeforeTest"`
	SkipMocksResetAfterTest  bool `yaml:"SkipMocksResetAfterTest"`
}

type TestInterfaceResult struct {
	Error               string                    `yaml:"Error"`
	GetName             string                    `yaml:"GetName"`
	GetDescription      string                    `yaml:"GetDescription"`
	GetMethod           string                    `yaml:"GetMethod"`
	Path                string                    `yaml:"Path"`
	ToQuery             string                    `yaml:"ToQuery"`
	ContentType         string                    `yaml:"ContentType"`
	Headers             map[string]string         `yaml:"Headers"`
	Cookies             map[string]string         `yaml:"Cookies"`
	GetRequest          string                    `yaml:"GetRequest"`
	GetForm             *formResult               `yaml:"GetForm"`
	GetMeta             map[string]interface{}    `yaml:"GetMeta"`
	GetStatus           models.Status             `yaml:"GetStatus"`
	GetResponses        map[int]string            `yaml:"GetResponses"`
	GetResponseHeaders  map[int]map[string]string `yaml:"GetResponseHeaders"`
	Fixtures            []string                  `yaml:"Fixtures"`
	GetDatabaseChecks   []databaseCheckResult     `yaml:"GetDatabaseChecks"`
	GetComparisonParams comparisonParamsResult    `yaml:"GetComparisonParams"`
	GetRetryPolicy      retryPolicyResult         `yaml:"GetRetryPolicy"`
	ServiceMocks        map[string]interface{}    `yaml:"ServiceMocks"`
	ServiceMocksParams  mocksResult               `yaml:"ServiceMocksParams"`
	Pause               time.Duration             `yaml:"Pause"`
	AfterRequestPause   time.Duration             `yaml:"AfterRequestPause"`
	BeforeScript        scriptResult              `yaml:"BeforeScript"`
	AfterRequestScript  scriptResult              `yaml:"AfterRequestScript"`

	GetVariables         map[string]string         `yaml:"GetVariables"`
	GetCombinedVariables map[string]string         `yaml:"GetCombinedVariables"`
	GetVariablesToSet    map[int]map[string]string `yaml:"GetVariablesToSet"`

	GetFileName     string `yaml:"GetFileName"`
	GetLineNumber   int    `yaml:"GetLineNumber"`
	FirstTestInFile bool   `yaml:"FirstTestInFile"`
	LastTestInFile  bool   `yaml:"LastTestInFile"`
	OneOfCase       bool   `yaml:"OneOfCase"`
}

func compareTestInterface(t *testing.T, expected *TestInterfaceResult, actual models.TestInterface) {
	assert.Equal(t, expected.GetName, actual.GetName(), "GetName returns wrong value")
	assert.Equal(t, expected.GetDescription, actual.GetDescription(), "GetDescription returns wrong value")
	assert.Equal(t, expected.GetMethod, actual.GetMethod(), "GetMethod returns wrong value")
	assert.Equal(t, expected.Path, actual.Path(), "Path returns wrong value")
	assert.Equal(t, expected.ToQuery, actual.ToQuery(), "ToQuery returns wrong value")
	assert.Equal(t, expected.ContentType, actual.ContentType(), "ContentType returns wrong value")
	if len(expected.Headers) != 0 {
		assert.Equal(t, expected.Headers, actual.Headers(), "Headers returns wrong value")
	} else {
		assert.Equal(t, len(expected.Headers), len(actual.Headers()), "Headers has different number of elements")
	}

	if len(expected.Cookies) != 0 {
		assert.Equal(t, expected.Cookies, actual.Cookies(), "Cookies returns wrong value")
	} else {
		assert.Empty(t, actual.Cookies(), "Cookies should be empty")
	}

	assert.Equal(t, expected.GetRequest, actual.GetRequest(), "GetRequest returns wrong value")

	compareForm(t, expected.GetForm, actual.GetForm())

	assert.Equal(t, expected.GetStatus, actual.GetStatus(), "GetStatus returns wrong value")
	if expected.GetResponses == nil {
		expected.GetResponses = map[int]string{}
	}

	sameResponse := assert.Equal(t, expected.GetResponses, actual.GetResponses(), "GetResponses returns wrong value")
	if sameResponse {
		for code, value := range expected.GetResponses {
			actualValue, valueExists := actual.GetResponse(code)
			assert.True(t, valueExists, "GetResponse for %d code doesn't not exists", code)
			assert.Equal(t, value, actualValue, "GetResponse for %d code returns wrong value", code)
		}
	}

	// get response for some not existing code
	if actualValue, actualExists := actual.GetResponse(12345); true {
		assert.False(t, actualExists, "GetResponse return something for 12345 code")
		assert.Equal(t, "", actualValue, "GetResponse return something for 12345 code")
	}

	for code, value := range expected.GetResponseHeaders {
		actualValue, valueExists := actual.GetResponseHeaders(code)
		assert.True(t, valueExists, "GetResponseHeaders for %d code doesn't not exists", code)
		assert.Equal(t, value, actualValue, "GetResponseHeaders for %d code returns wrong value", code)
	}

	// get response headers for some not existing code
	if actualValue, actualExists := actual.GetResponseHeaders(12345); true {
		assert.False(t, actualExists, "GetResponseHeaders return something for 12345 code")
		assert.Nil(t, actualValue, "GetResponseHeaders return something for 12345 code")
	}

	for key, value := range expected.GetMeta {
		assert.Equal(t, value, actual.GetMeta(key), "GetMeta for '%s' key returns wrong value", key)
	}

	// get meta for some not existing key
	if actualValue := actual.GetMeta("some random key"); true {
		assert.Nil(t, actualValue, "GetMeta return something for random key")
	}

	assert.Equal(t, expected.Fixtures, actual.Fixtures(), "Fixtures returns wrong value")

	compareDatabaseCheckResult(t, expected.GetDatabaseChecks, actual.GetDatabaseChecks())
	compareComparisonParams(t, expected.GetComparisonParams, actual.GetComparisonParams())
	compareRetryPolicy(t, expected.GetRetryPolicy, actual.GetRetryPolicy())

	assert.Equal(t, expected.ServiceMocks, actual.ServiceMocks(), "ServiceMocks returns wrong value")
	compareServiceMocksParams(t, expected.ServiceMocksParams, actual.ServiceMocksParams())

	assert.Equal(t, expected.Pause, actual.Pause(), "Pause returns wrong value")
	assert.Equal(t, expected.AfterRequestPause, actual.AfterRequestPause(), "AfterRequestPause returns wrong value")

	compareScript(t, expected.BeforeScript, actual.BeforeScript())
	compareScript(t, expected.AfterRequestScript, actual.AfterRequestScript())

	assert.Equal(t, expected.GetVariables, actual.GetVariables(), "GetVariables returns wrong value")

	if len(expected.GetCombinedVariables) != 0 {
		assert.Equal(t, expected.GetCombinedVariables, actual.GetCombinedVariables(), "GetCombinedVariables returns wrong value")
	} else {
		assert.Empty(t, actual.GetCombinedVariables(), "GetCombinedVariables should be empty")
	}

	for code := range expected.GetResponses {
		expectedValue, expectedExists := expected.GetVariablesToSet[code]
		actualValue, actualExists := actual.GetVariablesToSet(code)
		assert.Equal(t, expectedExists, actualExists, "GetVariablesToSet for %d code returns wrong expected state", code)
		assert.Equal(t, expectedValue, actualValue, "GetVariablesToSet for %d code returns wrong value", code)
	}

	// get variablesToSet for some not existing code
	if actualValue, actualExists := actual.GetVariablesToSet(12345); true {
		assert.False(t, actualExists, "GetVariablesToSet return something for 12345 code")
		assert.Nil(t, actualValue, "GetVariablesToSet return something for 12345 code")
	}

	normalizedPath := strings.ReplaceAll(actual.GetFileName(), "\\", "/")
	assert.Equal(t, expected.GetFileName, normalizedPath, "GetFileName returns wrong value")
	assert.Equal(t, expected.GetLineNumber, actual.GetLineNumber(), "GetLineNumber returns wrong value")
	assert.Equal(t, expected.FirstTestInFile, actual.FirstTestInFile(), "FirstTestInFile returns wrong value")
	assert.Equal(t, expected.LastTestInFile, actual.LastTestInFile(), "LastTestInFile returns wrong value")
	assert.Equal(t, expected.OneOfCase, actual.OneOfCase(), "OneOfCase returns wrong value")
}

func compareDatabaseCheckResult(t *testing.T, expected []databaseCheckResult, actual []models.DatabaseCheck) {
	if !assert.Equal(t, len(expected), len(actual), "GetDatabaseChecks returns wrong number of items") {
		return
	}
	for idx := range expected {
		assert.Equal(t, expected[idx].DbQueryString,
			actual[idx].DbQueryString(), "DbQueryString for DbCheck #%d", idx)
		assert.Equal(t, expected[idx].DbResponseJson,
			actual[idx].DbResponseJson(), "DbResponseJson for DbCheck #%d", idx)
		compareComparisonParams(t, expected[idx].GetComparisonParams,
			actual[idx].GetComparisonParams())
	}
}

func compareComparisonParams(t *testing.T, expected comparisonParamsResult, actual models.ComparisonParams) {
	assert.Equal(t, expected.IgnoreValuesChecking,
		actual.IgnoreValuesChecking(), "IgnoreValuesChecking returns wrong value")
	assert.Equal(t, expected.IgnoreArraysOrdering,
		actual.IgnoreArraysOrdering(), "IgnoreArraysOrdering returns wrong value")
	assert.Equal(t, expected.DisallowExtraFields,
		actual.DisallowExtraFields(), "DisallowExtraFields returns wrong value")
}

func compareScript(t *testing.T, expected scriptResult, actual models.Script) {
	assert.Equal(t, expected.CmdLine, actual.CmdLine(), "CmdLine returns wrong value")
	assert.Equal(t, expected.Timeout, actual.Timeout(), "Timeout returns wrong value")
}

func compareServiceMocksParams(t *testing.T, expected mocksResult, actual models.MocksParams) {
	assert.Equal(t, expected.SkipMocksResetBeforeTest, actual.SkipMocksResetBeforeTest(), "SkipMocksResetBeforeTest returns wrong value")
	assert.Equal(t, expected.SkipMocksResetAfterTest, actual.SkipMocksResetAfterTest(), "SkipMocksResetAfterTest returns wrong value")
}

func compareRetryPolicy(t *testing.T, expected retryPolicyResult, actual models.RetryPolicy) {
	assert.Equal(t, expected.Attempts, actual.Attempts(), "Attempts returns wrong value")
	assert.Equal(t, expected.Delay, actual.Delay(), "Delay returns wrong value")
	assert.Equal(t, expected.SuccessCount, actual.SuccessCount(), "SuccessCount returns wrong value")
}

func compareForm(t *testing.T, expected *formResult, actual models.Form) {
	if expected == nil {
		require.Nil(t, actual, "GetForm returns not-nil value")
		return
	}
	require.NotNil(t, actual, "GetForm returns nil value")
	assert.Equal(t, expected.GetFiles, actual.GetFiles(), "GetFiles returns wrong value")
	assert.Equal(t, expected.GetFields, actual.GetFields(), "GetFields returns wrong value")
}
