package compare

import (
	"fmt"
	"reflect"
	"regexp"
	"strings"

	"github.com/lansfy/gonkex/colorize"

	"github.com/kylelemons/godebug/diff"
)

type Params struct {
	IgnoreValues         bool `json:"ignoreValues" yaml:"ignoreValues"`
	IgnoreArraysOrdering bool `json:"ignoreArraysOrdering" yaml:"ignoreArraysOrdering"`
	DisallowExtraFields  bool `json:"disallowExtraFields" yaml:"disallowExtraFields"`
	failFast             bool // End compare operation after first error
}

type leafsMatchType int

const (
	pure leafsMatchType = iota
	regex
)

const (
	arrayType = "array"
	mapType   = "map"
)

var regexExprRx = regexp.MustCompile(`^\$matchRegexp\((.+)\)$`)

// StringAsRegexp ensures that provided string has format "$matchRegexp(...)" and returns
// the value from brackets
func StringAsRegexp(expr string) (string, bool) {
	if matches := regexExprRx.FindStringSubmatch(expr); matches != nil {
		return matches[1], true
	}

	return "", false
}

// Compare compares values as plain text
// It can be compared several ways:
//   - Pure values: should be equal
//   - Regex: try to compile 'expected' as regex and match 'actual' with it
//     It activates on following syntax: $matchRegexp(%EXPECTED_VALUE%)
func Compare(expected, actual interface{}, params Params) []error {
	return compareBranch("$", expected, actual, &params)
}

func compareBranch(path string, expected, actual interface{}, params *Params) []error {
	expectedType := getType(expected)
	actualType := getType(actual)

	// compare types
	if leafMatchType(expected) != regex && expectedType != actualType {
		return []error{makeError(path, "types do not match", expectedType, actualType)}
	}

	// compare scalars
	if isScalarType(actualType) && !params.IgnoreValues {
		return compareLeafs(path, expected, actual)
	}

	// compare arrays
	var errors []error
	if actualType == arrayType {
		expectedArray := convertToArray(expected)
		actualArray := convertToArray(actual)

		expectedArray, err := processMatchArrayByPattern(path, expectedArray, len(actualArray))
		if err != nil {
			return append(errors, err)
		}

		if len(expectedArray) != len(actualArray) {
			errors = append(errors, makeError(path, "array lengths do not match", len(expectedArray), len(actualArray)))
			return errors
		}

		if params.IgnoreArraysOrdering {
			expectedArray, actualArray = getUnmatchedArrays(expectedArray, actualArray, params)
		}

		// iterate over children
		for i, item := range expectedArray {
			subPath := fmt.Sprintf("%s[%d]", path, i)
			errors = append(errors, compareBranch(subPath, item, actualArray[i], params)...)
			if params.failFast && len(errors) != 0 {
				return errors
			}
		}
	}

	// compare maps
	if actualType == mapType {
		expectedRef := reflect.ValueOf(expected)
		actualRef := reflect.ValueOf(actual)

		if params.DisallowExtraFields && expectedRef.Len() != actualRef.Len() {
			errors = append(errors, makeError(path, "map lengths do not match", expectedRef.Len(), actualRef.Len()))
			return errors
		}

		for _, key := range expectedRef.MapKeys() {
			// check keys presence
			if ok := actualRef.MapIndex(key); !ok.IsValid() {
				errors = append(errors, makeError(path, "key is missing", key.String(), "<missing>"))
				if params.failFast {
					return errors
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
			errors = append(errors, res...)
			if params.failFast && len(errors) != 0 {
				return errors
			}
		}
	}

	return errors
}

func getType(value interface{}) string {
	if value == nil {
		return "nil"
	}

	rt := reflect.TypeOf(value)
	switch {
	case rt.Kind() == reflect.Slice || rt.Kind() == reflect.Array:
		return "array"
	case rt.Kind() == reflect.Map:
		return "map"
	default:
		return rt.String()
	}
}

func isScalarType(t string) bool {
	return !(t == "array" || t == "map")
}

func compareLeafs(path string, expected, actual interface{}) []error {
	switch leafMatchType(expected) {
	case pure:
		return comparePure(path, expected, actual)
	case regex:
		return compareRegex(path, expected, actual)
	default:
		return []error{fmt.Errorf("unknown compare type %q", expected)}
	}
}

func comparePure(path string, expected, actual interface{}) []error {
	if expected != actual {
		return []error{makeValueCompareError(path, "values do not match", expected, actual)}
	}
	return nil
}

func compareRegex(path string, expected, actual interface{}) []error {
	if !isScalarType(getType(actual)) {
		return []error{makeError(path, "type mismatch", "string", reflect.TypeOf(expected))}
	}

	regexExpr, _ := StringAsRegexp(expected.(string))

	rx, err := regexp.Compile(regexExpr)
	if err != nil {
		// simplify error text
		errorText := strings.TrimPrefix(err.Error(), "error parsing regexp: ")
		return []error{makeError(path, "cannot compile regexp", nil, errorText)}
	}

	value := fmt.Sprintf("%v", actual)
	if !rx.MatchString(value) {
		return []error{makeError(path, "value does not match regexp", expected, actual)}
	}

	return nil
}

func leafMatchType(expected interface{}) leafsMatchType {
	val, ok := expected.(string)
	if !ok {
		return pure
	}

	if _, ok = StringAsRegexp(val); ok {
		return regex
	}

	return pure
}

func diffStrings(a, b string) []colorize.Part {
	chunks := diff.DiffChunks(strings.Split(a, "\n"), strings.Split(b, "\n"))
	return colorize.MakeColorDiff(chunks)
}

func makeValueCompareError(path, msg string, expected, actual interface{}) error {
	expectedStr, ok1 := expected.(string)
	actualStr, ok2 := actual.(string)
	if !ok1 || !ok2 || !strings.Contains(actualStr+expectedStr, "\n") {
		return makeError(path, msg, expected, actual)
	}

	// special case for multi-line strings
	parts := []colorize.Part{
		colorize.Cyan(path),
	}

	parts = append(parts, diffStrings(expectedStr, actualStr)...)
	return colorize.NewError("at path %s "+msg+":\n     diff (--- expected vs +++ actual):\n", parts...)
}

func makeError(path, msg string, expected, actual interface{}) error {
	return colorize.NewNotEqualError("at path %s "+msg+":", path, expected, actual)
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

func processMatchArrayByPattern(path string, expectedArray []interface{}, actualLen int) ([]interface{}, error) {
	expectedLen := len(expectedArray)
	if expectedLen == 0 {
		return expectedArray, nil
	}

	val, ok := expectedArray[0].(string)
	if !ok || val != "$matchArray(pattern)" {
		return expectedArray, nil
	}

	if expectedLen != 2 {
		return expectedArray, makeError(path, "$matchArray(pattern) require only one additional element in array", 1, expectedLen-1)
	}

	res := make([]interface{}, 0, actualLen)
	for i := 0; i < actualLen; i++ {
		res = append(res, expectedArray[1])
	}

	return res, nil
}
