package yaml_file

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewTestWithCases(t *testing.T) {
	data := TestDefinition{
		RequestTmpl: `{"foo": "bar", "hello": {{ .hello }} }`,
		ResponseTmpls: map[int]string{
			200: `{"foo": "bar", "hello": {{ .hello }} }`,
			400: `{"foo": "bar", "hello": {{ .hello }} }`,
		},
		ResponseHeaders: map[int]map[string]string{
			200: {
				"hello": "world",
				"say":   "hello",
			},
			400: {
				"hello": "world",
				"foo":   "bar",
			},
		},
		Cases: []CaseData{
			{
				RequestArgs: map[string]interface{}{
					"hello": `"world"`,
				},
				ResponseArgs: map[int]map[string]interface{}{
					200: {
						"hello": "world",
					},
					400: {
						"hello": "world",
					},
				},
			},
			{
				RequestArgs: map[string]interface{}{
					"hello": `"world2"`,
				},
				ResponseArgs: map[int]map[string]interface{}{
					200: {
						"hello": "world2",
					},
					400: {
						"hello": "world2",
					},
				},
			},
		},
	}

	tests, err := makeTestFromDefinition("cases/example.yaml", &data)
	require.NoError(t, err)
	require.Len(t, tests, 2, "expected 2 tests")

	reqData, err := tests[0].ToJSON()
	require.NoError(t, err)
	require.JSONEq(t, `{"foo": "bar", "hello": "world" }`, string(reqData), "unexpected request JSON")

	filename := tests[0].GetFileName()
	require.Equal(t, "cases/example.yaml", filename, "unexpected filename")

	reqData, err = tests[1].ToJSON()
	require.NoError(t, err)
	require.JSONEq(t, `{"foo": "bar", "hello": "world2" }`, string(reqData), "unexpected request JSON")

	filename = tests[1].GetFileName()
	require.Equal(t, "cases/example.yaml", filename, "unexpected filename")
}
