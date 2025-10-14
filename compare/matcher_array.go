package compare

import (
	"errors"
)

func fillArrayWithPattern(pattern interface{}, arr []interface{}) {
	for idx := range arr {
		arr[idx] = pattern
	}
}

func processMatchArrayByPattern(expectedArray []interface{}, actualLen int) ([]interface{}, error) {
	expectedLen := len(expectedArray)
	if expectedLen == 0 {
		return expectedArray, nil
	}

	name, args := findMatcher(expectedArray[0])
	if name != "$matchArray" {
		return expectedArray, nil
	}

	res := make([]interface{}, actualLen)

	switch args {
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
			"unknown mode:", []string{"pattern", "pattern+subset", "subset+pattern"}, args))
	}
	return res, nil
}
