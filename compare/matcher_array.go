package compare

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/lansfy/gonkex/colorize"
)

func fillArrayWithPattern(pattern interface{}, arr []interface{}) {
	for idx := range arr {
		arr[idx] = pattern
	}
}

type arrayParamsData struct {
	mode    string
	size    int
	minsize int
	maxsize int
}

var arrayDefaultParams = map[string]string{
	"size":    "",
	"minsize": "",
	"maxsize": "",
}

func newValueError(pattern, value string) *colorize.Error {
	return colorize.NewError(pattern, colorize.Red(value))
}

func parseNonNegativeInt(params map[string]string, paramName string) (int, error) {
	strValue := params[paramName]
	if strValue == "" {
		return -1, nil
	}

	value, err := strconv.Atoi(strValue)
	if err != nil {
		return 0, colorize.NewEntityError("parameter %s", paramName).WithSubError(
			newValueError("parameter value (%s) must be integer", strValue))
	}
	if value < 0 {
		return 0, colorize.NewEntityError("parameter %s", paramName).WithSubError(
			newValueError("parameter value (%s) must be non-negative", strValue))
	}
	return value, nil
}

func extractArrayArgs(data string) (*arrayParamsData, error) {
	mode, params, err := extractArgs(data, arrayDefaultParams)
	if err != nil {
		return nil, err
	}

	result := &arrayParamsData{
		mode: mode,
	}

	result.size, err = parseNonNegativeInt(params, "size")
	if err != nil {
		return nil, err
	}

	result.minsize, err = parseNonNegativeInt(params, "minsize")
	if err != nil {
		return nil, err
	}

	result.maxsize, err = parseNonNegativeInt(params, "maxsize")
	if err != nil {
		return nil, err
	}

	if result.size != -1 && (result.minsize != -1 || result.maxsize != -1) {
		return nil, colorize.NewEntityError("parameter %s", "size").WithSubError(
			errors.New("cannot be used with minsize/maxsize"),
		)
	}

	return result, nil
}

func validateArraySizeConstraints(actualLen int, args *arrayParamsData) error {
	if args.size != -1 && actualLen != args.size {
		return colorize.NewNotEqualError("array length does not match size constraint",
			args.size, actualLen)
	}

	if args.minsize != -1 && actualLen < args.minsize {
		return colorize.NewNotEqualError("array length is less than minsize constraint",
			fmt.Sprintf(">= %d", args.minsize), actualLen)
	}

	if args.maxsize != -1 && actualLen > args.maxsize {
		return colorize.NewNotEqualError("array length is greater than maxsize constraint",
			fmt.Sprintf("<= %d", args.maxsize), actualLen)
	}

	return nil
}

func processMatchArrayByPattern(expectedArray []interface{}, actualLen int) ([]interface{}, error) {
	expectedLen := len(expectedArray)
	if expectedLen == 0 {
		return expectedArray, nil
	}

	name, argsStr := findMatcher(expectedArray[0])
	if name != "$matchArray" {
		return expectedArray, nil
	}

	args, err := extractArrayArgs(argsStr)
	if err != nil {
		return nil, makeMatcherParseError("$matchArray", err)
	}

	err = validateArraySizeConstraints(actualLen, args)
	if err != nil {
		return nil, err
	}

	res := make([]interface{}, actualLen)

	switch args.mode {
	case "pattern":
		if expectedLen != 2 {
			return nil, errors.New("array with $matchArray(pattern) must have one pattern element")
		}
		fillArrayWithPattern(expectedArray[1], res)
	case "subset+pattern":
		if expectedLen < 3 {
			return nil, errors.New("array with $matchArray(subset+pattern) must have pattern and additional elements")
		}
		fillArrayWithPattern(expectedArray[len(expectedArray)-1], res)
		copy(res, expectedArray[1:len(expectedArray)-1])
	case "pattern+subset":
		if expectedLen < 3 {
			return nil, errors.New("array with $matchArray(pattern+subset) must have pattern and additional elements")
		}
		fillArrayWithPattern(expectedArray[1], res)
		subset := expectedArray[2:]
		copy(res[len(res)-len(subset):], subset)
	default:
		return nil, makeMatcherParseError("$matchArray", makeValueNotInArrayError(
			"unknown mode:", []string{"pattern", "pattern+subset", "subset+pattern"}, args.mode))
	}
	return res, nil
}
