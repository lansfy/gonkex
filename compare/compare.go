package compare

import (
	"errors"
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
type leafTypeSet map[leafType]bool

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
		err := matcher.MatchValues(actual)
		if err != nil {
			return []error{colorize.NewPathError(path, err)}
		}
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
		if params.IgnoreValues || expected == actual {
			return nil
		}
		return []error{makeValueCompareError(path, "values do not match", expected, actual)}
	}
}

func compareArrays(path string, expected, actual interface{}, params *Params) []error {
	expectedArray := convertToArray(expected)
	actualArray := convertToArray(actual)

	expectedArray, err := processMatchArrayByPattern(path, expectedArray, len(actualArray))
	if err != nil {
		return []error{err}
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

func checkTypeCompatibility(supported leafTypeSet, value interface{}) error {
	valueType := getLeafType(value)
	if _, ok := supported[valueType]; ok {
		return nil
	}

	available := []string{}
	for name := range supported {
		available = append(available, string(name))
	}
	sort.Strings(available)

	return makeTypeMismatchError(strings.Join(available, " / "), string(valueType))
}

func makeTypeMismatchError(expectedType, actualType string) error {
	return colorize.NewNotEqualError("type mismatch:", expectedType, actualType)
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

func fillArrayWithPattern(pattern interface{}, arr []interface{}) {
	for idx := range arr {
		arr[idx] = pattern
	}
}

func processMatchArrayByPattern(path string, expectedArray []interface{}, actualLen int) ([]interface{}, error) {
	expectedLen := len(expectedArray)
	if expectedLen == 0 {
		return expectedArray, nil
	}

	val, ok := expectedArray[0].(string)
	if !ok || !strings.HasPrefix(val, "$matchArray(") || !strings.HasSuffix(val, ")") {
		return expectedArray, nil
	}

	params := val[12 : len(val)-1]

	res := make([]interface{}, actualLen)

	switch params {
	case "pattern":
		if expectedLen != 2 {
			return nil, colorize.NewPathError(path, errors.New("array with $matchArray(pattern) must have one pattern element"))
		}
		fillArrayWithPattern(expectedArray[1], res)
	case "subset+pattern":
		if expectedLen < 3 {
			return nil, colorize.NewPathError(path, errors.New("array with $matchArray(subset+pattern) must have pattern and additional elements"))
		}
		fillArrayWithPattern(expectedArray[len(expectedArray)-1], res)
		copy(res, expectedArray[1:len(expectedArray)-1])
	case "pattern+subset":
		if expectedLen < 3 {
			return nil, colorize.NewPathError(path, errors.New("array with $matchArray(pattern+subset) must have pattern and additional elements"))
		}
		fillArrayWithPattern(expectedArray[1], res)
		subset := expectedArray[2:]
		copy(res[len(res)-len(subset):], subset)
	default:
		return nil, makeError(path, "unknown $matchArray mode", "pattern / pattern+subset / subset+pattern", params)
	}
	return res, nil
}
