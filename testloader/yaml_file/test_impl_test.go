package yaml_file

import (
	"testing"

	"github.com/lansfy/gonkex/models"
	"github.com/lansfy/gonkex/variables"

	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParse_TestWithEniromentVariables(t *testing.T) {
	err := godotenv.Load("testdata/test.env")
	require.NoError(t, err)

	tests, err := parseTestDefinitionFile(DefaultFileParse, "testdata/variables-enviroment.yaml")
	require.NoError(t, err)

	testOriginal := tests[0]

	vars := variables.New()
	testApplied := testOriginal.Clone()
	testApplied.ApplyVariables(vars.Substitute)

	assert.Equal(t, "/some/path/path_value", testApplied.Path())

	resp, ok := testApplied.GetResponse(200)
	assert.True(t, ok)
	assert.Equal(t, "resp_val", resp)
}

func TestParse_TestsWithVariables(t *testing.T) {
	tests, err := parseTestDefinitionFile(DefaultFileParse, "testdata/variables.yaml")
	require.NoError(t, err)

	testOriginal := tests[0]

	vars := variables.New()
	vars.Merge(testOriginal.GetVariables())
	assert.NoError(t, err)

	testApplied := testOriginal.Clone()
	testApplied.ApplyVariables(vars.Substitute)

	// check that original test is not changed
	checkOriginal(t, testOriginal, false)

	checkApplied(t, testApplied, false)
}

func TestParse_TestsWithCombinedVariables(t *testing.T) {
	tests, err := parseTestDefinitionFile(DefaultFileParse, "testdata/combined-variables.yaml")
	require.NoError(t, err)

	testOriginal := tests[0]

	vars := variables.New()
	vars.Merge(testOriginal.GetCombinedVariables())
	assert.NoError(t, err)

	testApplied := testOriginal.Clone()
	testApplied.ApplyVariables(vars.Substitute)

	// check that original test is not changed
	checkOriginal(t, testOriginal, true)

	checkApplied(t, testApplied, true)
}

func checkOriginal(t *testing.T, test models.TestInterface, combined bool) {
	t.Helper()

	req := test.GetRequest()
	assert.Equal(t, `{"reqParam": "{{ $reqParam }}"}`, req)

	assert.Equal(t, "{{ $method }}", test.GetMethod())
	assert.Equal(t, "/some/path/{{ $pathPart }}", test.Path())
	assert.Equal(t, "{{ $query }}", test.ToQuery())
	assert.Equal(t, map[string]string{"header1": "{{ $header }}"}, test.Headers())

	resp, ok := test.GetResponse(200)
	assert.True(t, ok)
	assert.Equal(t, "{{ $resp }}", resp)

	resp, ok = test.GetResponse(404)
	assert.True(t, ok)
	assert.Equal(t, "{{ $respRx }}", resp)

	if combined {
		resp, ok = test.GetResponse(501)
		assert.True(t, ok)
		assert.Equal(t, "{{ $newVar }} - {{ $redefinedVar }}", resp)
	}
}

func checkApplied(t *testing.T, test models.TestInterface, combined bool) {
	t.Helper()

	req := test.GetRequest()
	assert.Equal(t, `{"reqParam": "reqParam_value"}`, req)

	assert.Equal(t, "POST", test.GetMethod())
	assert.Equal(t, "/some/path/part_of_path", test.Path())
	assert.Equal(t, "?query_val", test.ToQuery())
	assert.Equal(t, map[string]string{"header1": "header_val"}, test.Headers())

	resp, ok := test.GetResponse(200)
	assert.True(t, ok)
	assert.Equal(t, "resp_val", resp)

	resp, ok = test.GetResponse(404)
	assert.True(t, ok)
	assert.Equal(t, "$matchRegexp(^[0-9.]+$)", resp)

	resp, ok = test.GetResponse(500)
	assert.True(t, ok)
	assert.Equal(t, "existingVar_Value - {{ $notExistingVar }}", resp)

	raw, ok := test.ServiceMocks()["server"]
	assert.True(t, ok)
	mockMap, ok := raw.(map[string]interface{})
	assert.True(t, ok)
	mockBody, ok := mockMap["body"]
	assert.True(t, ok)
	assert.Equal(t, "{\"reqParam\": \"reqParam_value\"}", mockBody)

	if combined {
		resp, ok = test.GetResponse(501)
		assert.True(t, ok)
		assert.Equal(t, "some_value - redefined_value", resp)
	}
}
