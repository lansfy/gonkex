package response_body

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/lansfy/gonkex/checker"
	"github.com/lansfy/gonkex/colorize"
	"github.com/lansfy/gonkex/compare"
	"github.com/lansfy/gonkex/models"
	"github.com/lansfy/gonkex/types"
)

func NewChecker() checker.CheckerInterface {
	return &responseBodyChecker{}
}

type responseBodyChecker struct{}

func (c *responseBodyChecker) Check(t models.TestInterface, result *models.Result) ([]error, error) {
	expectedBody, ok := t.GetResponse(result.ResponseStatusCode)
	if !ok {
		return []error{createWrongStatusError(result.ResponseStatusCode, t.GetResponses())}, nil
	}

	// expected body has only regexp, so compare bodies as strings
	if compare.StringAsMatcher(expectedBody) != nil {
		return addMainError(compare.Compare(expectedBody, result.ResponseBody, compare.Params{})), nil
	}

	if expectedBody != "" {
		for _, b := range types.GetRegisteredBodyTypes() {
			if b.IsSupportedContentType(result.ResponseContentType) {
				return compareBody(t, expectedBody, result, b.GetName(), b.Decode)
			}
		}
	}

	// compare bodies as strings
	return addMainError(compare.Compare(expectedBody, result.ResponseBody, compare.Params{})), nil
}

func createWrongStatusError(statusCode int, known map[int]string) error {
	knownCodes := []string{}
	for code := range known {
		knownCodes = append(knownCodes, strconv.Itoa(code))
	}
	return colorize.NewNotEqualError("server responded with unexpected %s:", "status", strings.Join(knownCodes, " / "), statusCode)
}

func addMainError(source []error) []error {
	var errs []error
	for _, err := range source {
		errs = append(errs, colorize.NewEntityError("service %s comparison", "response body").SetSubError(err))
	}
	return errs
}

func compareBody(t models.TestInterface, expectedBody string, result *models.Result, typeName string,
	decode func(body string) (interface{}, error)) ([]error, error) {
	// decode expected body
	expected, err := decode(expectedBody)
	if err != nil {
		return nil, fmt.Errorf("invalid %s in response in the test declaration (for status %d): %w", typeName, result.ResponseStatusCode, err)
	}

	// decode actual body
	actual, err := decode(result.ResponseBody)
	if err != nil {
		return []error{fmt.Errorf("could not parse service response as %s", typeName)}, nil
	}

	return addMainError(compare.Compare(expected, actual, getCompareParams(t))), nil
}

func getCompareParams(t models.TestInterface) compare.Params {
	params := t.GetComparisonParams()
	return compare.Params{
		IgnoreValues:         params.IgnoreValuesChecking(),
		IgnoreArraysOrdering: params.IgnoreArraysOrdering(),
		DisallowExtraFields:  params.DisallowExtraFields(),
	}
}
