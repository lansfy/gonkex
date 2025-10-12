package compare

import (
	"fmt"
	"reflect"
	"sort"
	"strings"

	"github.com/lansfy/gonkex/colorize"
)

type Params struct {
	IgnoreValues         bool `json:"ignoreValues" yaml:"ignoreValues"`
	IgnoreArraysOrdering bool `json:"ignoreArraysOrdering" yaml:"ignoreArraysOrdering"`
	DisallowExtraFields  bool `json:"disallowExtraFields" yaml:"disallowExtraFields"`
	failFast             bool // End compare operation after first error
}

// Compare compares expected and actual values
func Compare(expected, actual interface{}, params Params) []error {
	return compareBranch("$", expected, actual, &params)
}

type leafType string

func (t leafType) IsScalar() bool {
	return t != leafArray && t != leafMap
}

const (
	leafArray  leafType = "array"
	leafMap    leafType = "map"
	leafNil    leafType = "nil"
	leafBool   leafType = "bool"
	leafString leafType = "string"
	leafNumber leafType = "number"
)

func compareBranch(path string, expected, actual interface{}, params *Params) []error {
	expectedType := getLeafType(expected)
	actualType := getLeafType(actual)

	if matcher := CreateMatcher(expected); matcher != nil {
		if params.IgnoreValues && actualType.IsScalar() {
			return nil
		}
		err := matcher.MatchValues(actual)
		if err != nil {
			return []error{colorize.NewPathError(path, err)}
		}
		return nil
	}

	if params.IgnoreValues && actualType.IsScalar() && expectedType.IsScalar() {
		return nil
	}

	if expectedType != actualType {
		return []error{makeError(path, "types do not match", expectedType, actualType)}
	}

	switch actualType {
	case leafArray:
		return compareArrays(path, expected, actual, params)
	case leafMap:
		return compareMaps(path, expected, actual, params)
	default:
		if expected != actual {
			return []error{makeValueCompareError(path, "values do not match", expected, actual)}
		}
		return nil
	}
}

func compareArrays(path string, expected, actual interface{}, params *Params) []error {
	expectedArray := convertToArray(expected)
	actualArray := convertToArray(actual)

	expectedArray, err := processMatchArrayByPattern(expectedArray, len(actualArray))
	if err != nil {
		return []error{colorize.NewPathError(path, err)}
	}

	if len(expectedArray) != len(actualArray) {
		return []error{makeError(path, "array lengths do not match", len(expectedArray), len(actualArray))}
	}

	if params.IgnoreArraysOrdering {
		expectedArray, actualArray = getUnmatchedArrays(expectedArray, actualArray, params)
	}

	// iterate over children
	var errs []error
	for i, item := range expectedArray {
		subPath := fmt.Sprintf("%s[%d]", path, i)
		errs = append(errs, compareBranch(subPath, item, actualArray[i], params)...)
		if params.failFast && len(errs) != 0 {
			return errs
		}
	}
	return errs
}

func compareMaps(path string, expected, actual interface{}, params *Params) []error {
	expectedRef := reflect.ValueOf(expected)
	actualRef := reflect.ValueOf(actual)

	if params.DisallowExtraFields && expectedRef.Len() != actualRef.Len() {
		return []error{makeError(path, "map lengths do not match", expectedRef.Len(), actualRef.Len())}
	}

	var errs []error
	for _, key := range expectedRef.MapKeys() {
		// check keys presence
		if ok := actualRef.MapIndex(key); !ok.IsValid() {
			errs = append(errs, makeError(path, "key is missing", key.String(), "<missing>"))
			if params.failFast {
				return errs
			}
			continue
		}

		// check values
		subPath := fmt.Sprintf("%s.%s", path, key.String())
		res := compareBranch(
			subPath,
			expectedRef.MapIndex(key).Interface(),
			actualRef.MapIndex(key).Interface(),
			params,
		)
		errs = append(errs, res...)
		if params.failFast && len(errs) != 0 {
			return errs
		}
	}
	return errs
}

func getLeafType(value interface{}) leafType {
	if value == nil {
		return leafNil
	}

	rt := reflect.TypeOf(value)
	switch rt.Kind() {
	case reflect.Bool:
		return leafBool
	case reflect.String:
		return leafString
	case reflect.Slice, reflect.Array:
		return leafArray
	case reflect.Map:
		return leafMap
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Float32, reflect.Float64:
		return leafNumber
	default:
		return leafType(rt.String())
	}
}

func makeMatcherParseError(name string, err error) error {
	return colorize.NewEntityError("parse %s", name).WithSubError(err)
}

func makeValueNotInArrayError(text string, expectedValues []string, actualValue string) error {
	sort.Strings(expectedValues)
	return colorize.NewNotEqualError(
		text, strings.Join(expectedValues, " / "), actualValue)
}

func makeTypeMismatchError(expectedTypes []leafType, actualType leafType) error {
	arr := []string{}
	for _, i := range expectedTypes {
		arr = append(arr, string(i))
	}
	return makeValueNotInArrayError("type mismatch:", arr, string(actualType))
}

func makeValueCompareError(path, msg string, expected, actual interface{}) error {
	expectedStr, ok1 := expected.(string)
	actualStr, ok2 := actual.(string)
	if !ok1 || !ok2 || !strings.Contains(actualStr+expectedStr, "\n") {
		return makeError(path, msg, expected, actual)
	}

	// special case for multi-line strings
	diff := colorize.MakeColorDiff(
		"\n     diff (--- expected vs +++ actual):\n",
		strings.Split(expectedStr, "\n"),
		strings.Split(actualStr, "\n"),
	)
	return colorize.NewPathError(
		path,
		colorize.NewError(msg+":").WithPostfix(diff),
	)
}

func makeError(path, msg string, expected, actual interface{}) error {
	return colorize.NewPathError(path, colorize.NewNotEqualError(msg+":", expected, actual))
}

func convertToArray(array interface{}) []interface{} {
	ref := reflect.ValueOf(array)

	interfaceSlice := make([]interface{}, 0, ref.Len())
	for i := 0; i < ref.Len(); i++ {
		interfaceSlice = append(interfaceSlice, ref.Index(i).Interface())
	}

	return interfaceSlice
}

// For every elem in "expected" try to find elem in "actual". Returns arrays without matching.
func getUnmatchedArrays(expected, actual []interface{}, params *Params) (expectedUnmatched, actualUnmatched []interface{}) {
	expectedError := make([]interface{}, 0)

	failfastParams := *params
	failfastParams.failFast = true

	for _, expectedElem := range expected {
		found := false
		for i, actualElem := range actual {
			if len(compareBranch("", expectedElem, actualElem, &failfastParams)) == 0 {
				// expectedElem match actualElem
				found = true
				// remove actualElem from  actual
				if len(actual) != 1 {
					actual[i] = actual[len(actual)-1]
				}
				actual = actual[:len(actual)-1]

				break
			}
		}
		if !found {
			expectedError = append(expectedError, expectedElem)
			if params.failFast {
				return expectedError, actual[0:1]
			}
		}
	}

	return expectedError, actual
}
