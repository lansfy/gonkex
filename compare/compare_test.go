package compare

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func makeErrorString(path, msg string, expected, actual interface{}) string {
	return fmt.Sprintf(
		"at path '%s' %s:\n     expected: %v\n       actual: %v",
		path,
		msg,
		expected,
		actual,
	)
}

func TestCompareNils(t *testing.T) {
	errors := Compare(nil, nil, Params{})
	if len(errors) != 0 {
		t.Error(
			"must return no errors",
			fmt.Sprintf("got result: %v", errors),
		)
		t.Fail()
	}
}

func TestCompareNilWithNonNil(t *testing.T) {
	errors := Compare("", nil, Params{})
	if errors[0].Error() != makeErrorString("$", "types do not match", "string", "nil") {
		t.Error(
			"must return one error",
			fmt.Sprintf("got result: %v", errors),
		)
		t.Fail()
	}
}

func TestCompareEqualStrings(t *testing.T) {
	errors := Compare("1", "1", Params{})
	if len(errors) != 0 {
		t.Error(
			"must return no errors",
			fmt.Sprintf("got result: %v", errors),
		)
		t.Fail()
	}
}

func TestCompareDifferStrings(t *testing.T) {
	errors := Compare("1", "2", Params{})
	if errors[0].Error() != makeErrorString("$", "values do not match", 1, 2) {
		t.Error(
			"must return one error",
			fmt.Sprintf("got result: %v", errors),
		)
		t.Fail()
	}
}

func TestCompareEqualIntegers(t *testing.T) {
	errors := Compare(1, 1, Params{})
	if len(errors) != 0 {
		t.Error(
			"must return no errors",
			fmt.Sprintf("got result: %v", errors),
		)
		t.Fail()
	}
}

func TestCompareDifferIntegers(t *testing.T) {
	errors := Compare(1, 2, Params{})
	if errors[0].Error() != makeErrorString("$", "values do not match", 1, 2) {
		t.Error(
			"must return one error",
			fmt.Sprintf("got result: %v", errors),
		)
		t.Fail()
	}
}

func TestCheckRegexMach(t *testing.T) {
	errors := Compare("$matchRegexp(x.+z)", "xyyyz", Params{})
	if len(errors) != 0 {
		t.Error(
			"must return no errors",
			fmt.Sprintf("got result: %v", errors),
		)
		t.Fail()
	}
}

func TestCheckRegexNotMach(t *testing.T) {
	errors := Compare("$matchRegexp(x.+z)", "ayyyb", Params{})
	if errors[0].Error() != makeErrorString("$",
		"value does not match regex", "$matchRegexp(x.+z)", "ayyyb") {
		t.Error(
			"must return one error",
			fmt.Sprintf("got result: %v", errors),
		)
		t.Fail()
	}
}

func TestCheckRegexCantCompile(t *testing.T) {
	errors := Compare("$matchRegexp((?x))", "2", Params{})
	if errors[0].Error() != makeErrorString("$", "can not compile regex", nil, "error") {
		t.Error(
			"must return one error",
			fmt.Sprintf("got result: %v", errors),
		)
		t.Fail()
	}
}

func TestCompareEqualArrays(t *testing.T) {
	array1 := []string{"1", "2"}
	array2 := []string{"1", "2"}
	errors := Compare(array1, array2, Params{})
	if len(errors) != 0 {
		t.Error(
			"must return no errors",
			fmt.Sprintf("got result: %v", errors),
		)
		t.Fail()
	}
}

func TestCompareEqualArraysWithDifferentElementsOrder(t *testing.T) {
	array1 := []string{"1", "2"}
	array2 := []string{"2", "1"}
	errors := Compare(array1, array2, Params{IgnoreArraysOrdering: true})
	if len(errors) != 0 {
		t.Error(
			"must return no errors",
			fmt.Sprintf("got result: %v", errors),
		)
		t.Fail()
	}
}

func TestCompareArraysDifferLengths(t *testing.T) {
	array1 := []string{"1", "2", "3"}
	array2 := []string{"1", "2"}
	errors := Compare(array1, array2, Params{})
	if errors[0].Error() != makeErrorString("$", "array lengths do not match", 3, 2) {
		t.Error(
			"must return one error",
			fmt.Sprintf("got result: %v", errors),
		)
		t.Fail()
	}
}

func TestCompareDifferArrays(t *testing.T) {
	array1 := []string{"1", "2"}
	array2 := []string{"1", "3"}
	errors := Compare(array1, array2, Params{})
	if errors[0].Error() != makeErrorString("$[1]", "values do not match", 2, 3) {
		t.Error(
			"must return one error",
			fmt.Sprintf("got result: %v", errors),
		)
		t.Fail()
	}
}

func TestCompareArraysFewErrors(t *testing.T) {
	array1 := []string{"1", "2", "3"}
	array2 := []string{"1", "3", "4"}
	errors := Compare(array1, array2, Params{})
	assert.Len(t, errors, 2)
}

func TestCompareNestedEqualArrays(t *testing.T) {
	array1 := [][]string{{"1", "2"}, {"3", "4"}}
	array2 := [][]string{{"1", "2"}, {"3", "4"}}
	errors := Compare(array1, array2, Params{})
	if len(errors) != 0 {
		t.Error(
			"must return no errors",
			fmt.Sprintf("got result: %v", errors),
		)
		t.Fail()
	}
}

func TestCompareNestedDifferArrays(t *testing.T) {
	array1 := [][]string{{"1", "2"}, {"3", "4"}}
	array2 := [][]string{{"1", "2"}, {"3", "5"}}
	errors := Compare(array1, array2, Params{})
	if errors[0].Error() != makeErrorString("$[1][1]", "values do not match", 4, 5) {
		t.Error(
			"must return one error",
			fmt.Sprintf("got result: %v", errors),
		)
		t.Fail()
	}
}

func TestCompareArraysWithRegex(t *testing.T) {
	arrayExpected := []string{"2", "$matchRegexp(x.+z)"}
	arrayActual := []string{"2", "xyyyz"}

	errors := Compare(arrayExpected, arrayActual, Params{})
	if len(errors) != 0 {
		t.Error(
			"must return no errors",
			fmt.Sprintf("got result: %v", errors),
		)
		t.Fail()
	}
}

func TestCompareArraysWithRegexMixedTypes(t *testing.T) {
	arrayExpected := []string{"2", "$matchRegexp([0-9]+)"}
	arrayActual := []interface{}{"2", 123}

	errors := Compare(arrayExpected, arrayActual, Params{})
	if len(errors) != 0 {
		t.Error(
			"must return no errors",
			fmt.Sprintf("got result: %v", errors),
		)
		t.Fail()
	}
}

func TestCompareArraysWithRegexNotMatch(t *testing.T) {
	arrayExpected := []string{"2", "$matchRegexp(x.+z)"}
	arrayActual := []string{"2", "ayyyb"}

	errors := Compare(arrayExpected, arrayActual, Params{})
	expectedErrors := makeErrorString("$[1]",
		"value does not match regex", "$matchRegexp(x.+z)", "ayyyb")
	if errors[0].Error() != expectedErrors {
		t.Error(
			"must return one error",
			fmt.Sprintf("got result: %v", errors),
		)
		t.Fail()
	}
}

func TestCompareEqualMaps(t *testing.T) {
	array1 := map[string]string{"a": "1", "b": "2"}
	array2 := map[string]string{"a": "1", "b": "2"}
	errors := Compare(array1, array2, Params{})
	if len(errors) != 0 {
		t.Error(
			"must return no errors",
			fmt.Sprintf("got result: %v", errors),
		)
		t.Fail()
	}
}

func TestCompareMapsWithRegex(t *testing.T) {
	mapExpected := map[string]string{"a": "1", "b": "$matchRegexp(x.+z)"}
	mapActual := map[string]string{"a": "1", "b": "xyyyz"}

	errors := Compare(mapExpected, mapActual, Params{})
	if len(errors) != 0 {
		t.Error(
			"must return no errors",
			fmt.Sprintf("got result: %v", errors),
		)
		t.Fail()
	}
}

func TestCompareMapsWithRegexNotMatch(t *testing.T) {
	mapExpected := map[string]string{"a": "1", "b": "$matchRegexp(x.+z)"}
	mapActual := map[string]string{"a": "1", "b": "ayyyb"}

	errors := Compare(mapExpected, mapActual, Params{})
	expectedErrors := makeErrorString("$.b", "value does not match regex", "$matchRegexp(x.+z)", "ayyyb")

	if errors[0].Error() != expectedErrors {
		t.Error(
			"must return one error",
			fmt.Sprintf("got result: %v", errors),
		)
		t.Fail()
	}
}

func TestCompareEqualMapsWithExtraFields(t *testing.T) {
	array1 := map[string]string{"a": "1", "b": "2"}
	array2 := map[string]string{"a": "1", "b": "2", "c": "3"}
	errors := Compare(array1, array2, Params{})
	if len(errors) != 0 {
		t.Error(
			"must return no errors",
			fmt.Sprintf("got result: %v", errors),
		)
		t.Fail()
	}
}

func TestCompareEqualMapsWithExtraFieldsCheckingEnabled(t *testing.T) {
	array1 := map[string]string{"a": "1", "b": "2"}
	array2 := map[string]string{"a": "1", "b": "2", "c": "3"}
	errors := Compare(array1, array2, Params{DisallowExtraFields: true})
	if errors[0].Error() != makeErrorString("$", "map lengths do not match", 2, 3) {
		t.Error(
			"must return one error",
			fmt.Sprintf("got result: %v", errors),
		)
		t.Fail()
	}
}

func TestCompareEqualMapsWithDifferentKeysOrder(t *testing.T) {
	array1 := map[string]string{"a": "1", "b": "2"}
	array2 := map[string]string{"b": "2", "a": "1"}
	errors := Compare(array1, array2, Params{})
	if len(errors) != 0 {
		t.Error(
			"must return no errors",
			fmt.Sprintf("got result: %v", errors),
		)
		t.Fail()
	}
}

func TestCompareMapsWithDifferentKeys(t *testing.T) {
	array1 := map[string]string{"a": "1", "b": "2"}
	array2 := map[string]string{"a": "1", "c": "2"}
	errors := Compare(array1, array2, Params{})
	expectedErr := makeErrorString("$", "key is missing", "b", "<missing>")
	if errors[0].Error() != expectedErr {
		t.Error(
			"must return one error",
			fmt.Sprintf("got result: %v", errors),
		)
		t.Fail()
	}
}

func TestCompareMapsWithDifferentValues(t *testing.T) {
	array1 := map[string]string{"a": "1", "b": "2"}
	array2 := map[string]string{"a": "1", "b": "3"}
	errors := Compare(array1, array2, Params{})
	if errors[0].Error() != makeErrorString("$.b", "values do not match", 2, 3) {
		t.Error(
			"must return one error",
			fmt.Sprintf("got result: %v", errors),
		)
		t.Fail()
	}
}

func TestCompareMapsWithFewErrors(t *testing.T) {
	array1 := map[string]string{"a": "1", "b": "2", "c": "5"}
	array2 := map[string]string{"a": "1", "b": "3", "d": "4"}
	errors := Compare(array1, array2, Params{})
	assert.Len(t, errors, 2)
}

func TestCompareEqualNestedMaps(t *testing.T) {
	array1 := map[string]map[string]string{"a": {"i": "3", "j": "4"}, "b": {"k": "5", "l": "6"}}
	array2 := map[string]map[string]string{"a": {"i": "3", "j": "4"}, "b": {"k": "5", "l": "6"}}
	errors := Compare(array1, array2, Params{})
	if len(errors) != 0 {
		t.Error(
			"must return no errors",
			fmt.Sprintf("got result: %v", errors),
		)
		t.Fail()
	}
}

func TestCompareNestedMapsWithDifferentKeys(t *testing.T) {
	array1 := map[string]map[string]string{"a": {"i": "3", "j": "4"}, "b": {"k": "5", "l": "6"}}
	array2 := map[string]map[string]string{"a": {"i": "3", "j": "4"}, "b": {"l": "6"}}
	errors := Compare(array1, array2, Params{})
	expectedErr := makeErrorString("$.b", "key is missing", "k", "<missing>")
	if errors[0].Error() != expectedErr {
		t.Error(
			"must return one error",
			fmt.Sprintf("got result: %v", errors),
		)
		t.Fail()
	}
}

func TestCompareNestedMapsWithDifferentValues(t *testing.T) {
	array1 := map[string]map[string]string{"a": {"i": "3", "j": "4"}, "b": {"k": "5", "l": "6"}}
	array2 := map[string]map[string]string{"a": {"i": "3", "j": "4"}, "b": {"k": "5", "l": "7"}}
	errors := Compare(array1, array2, Params{})
	if errors[0].Error() != makeErrorString("$.b.l", "values do not match", 6, 7) {
		t.Error(
			"must return one error",
			fmt.Sprintf("got result: %v", errors),
		)
		t.Fail()
	}
}

func TestCompareEqualJsonScalars(t *testing.T) {
	var json1, json2 interface{}
	_ = json.Unmarshal([]byte("1"), &json1)
	_ = json.Unmarshal([]byte("1"), &json2)
	errors := Compare(json1, json2, Params{})
	if len(errors) != 0 {
		t.Error(
			"must return no errors",
			fmt.Sprintf("got result: %v", errors),
		)
		t.Fail()
	}
}

func TestCompareDifferJsonScalars(t *testing.T) {
	var json1, json2 interface{}
	_ = json.Unmarshal([]byte("1"), &json1)
	_ = json.Unmarshal([]byte("2"), &json2)
	errors := Compare(json1, json2, Params{})
	if errors[0].Error() != makeErrorString("$", "values do not match", 1, 2) {
		t.Error(
			"must return one error",
			fmt.Sprintf("got result: %v", errors),
		)
		t.Fail()
	}
}

var expectedArrayJson = `
{
  "data":[
    {"name": "n111"},
    {"name": "n222"},
    {"name": "n333"}
  ]
}
`

var actualArrayJson = `
{
  "data": [
    {"message": "m555", "name": "n333"},
    {"message": "m777", "name": "n111"},
    {"message": "m999","name": "n222"}
  ]
}
`

func TestCompareEqualArraysWithIgnoreArraysOrdering(t *testing.T) {
	var json1, json2 interface{}
	_ = json.Unmarshal([]byte(expectedArrayJson), &json1)
	_ = json.Unmarshal([]byte(actualArrayJson), &json2)
	errors := Compare(json1, json2, Params{
		IgnoreArraysOrdering: true,
	})
	if len(errors) != 0 {
		t.Error(
			"must return no errors",
			fmt.Sprintf("got result: %v", errors),
		)
		t.Fail()
	}
}

func TestCompareEqualComplexJson(t *testing.T) {
	var json1, json2 interface{}
	_ = json.Unmarshal([]byte(complexJson1), &json1)
	_ = json.Unmarshal([]byte(complexJson1), &json2) // compare json with same json
	errors := Compare(json1, json2, Params{})
	if len(errors) != 0 {
		t.Error(
			"must return no errors",
			fmt.Sprintf("got result: %v", errors),
		)
		t.Fail()
	}
}

func TestCompareDifferComplexJson(t *testing.T) {
	var json1, json2 interface{}
	_ = json.Unmarshal([]byte(complexJson1), &json1)
	_ = json.Unmarshal([]byte(complexJson2), &json2)
	errors := Compare(json1, json2, Params{})
	expectedErr := makeErrorString(
		"$.paths./api/get-delivery-info.get.parameters[2].$ref",
		"values do not match",
		"#/parameters/profile_id",
		"#/parameters/profile_id2",
	)
	if len(errors) == 0 || errors[0].Error() != expectedErr {
		t.Error(
			"must return one error",
			fmt.Sprintf("got result: %v", errors),
		)
		t.Fail()
	}
}

//go:embed testdata/complex_data_1.json
var complexJson1 string

//go:embed testdata/complex_data_2.json
var complexJson2 string
