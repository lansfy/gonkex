package response_body

import (
	"fmt"
	"sort"
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
	if len(t.GetResponses()) == 0 {
		// possible response codes not specified, ignore any checks
		return nil, nil
	}
	expectedBody, ok := t.GetResponse(result.ResponseStatusCode)
	if !ok {
		return []error{createWrongStatusError(result.ResponseStatusCode, t.GetResponses())}, nil
	}

	// expected body has only matcher, so compare bodies as strings
	if compare.CreateMatcher(expectedBody) != nil {
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
	sort.Strings(knownCodes)
	return colorize.NewEntityNotEqualError("server responded with unexpected %s:",
		"status code", strings.Join(knownCodes, " / "), statusCode)
}

func addMainError(source []error) []error {
	var errs []error
	for _, err := range source {
		errs = append(errs, colorize.NewEntityError("service %s comparison", "response body").WithSubError(err))
	}
	return errs
}

func compareBody(t models.TestInterface, expectedBody string, result *models.Result, typeName string,
	decode func(body string) (interface{}, error)) ([]error, error) {
	// decode expected body and service response
	expected, expectedErr := decode(expectedBody)
	actual, responseErr := decode(result.ResponseBody)

	if expectedErr != nil && responseErr != nil {
		// both entities can't be parsed as provided type, so compare bodies as strings
		return addMainError(compare.Compare(expectedBody, result.ResponseBody, compare.Params{})), nil
	}

	if expectedErr != nil {
		err := fmt.Errorf("failed to load value as %s, compare response body as plain text", typeName)
		errs := []error{
			colorize.NewEntityError("body definition at path %s", fmt.Sprintf("$.response.%d", result.ResponseStatusCode)).WithSubError(err),
		}
		errs = append(errs, addMainError(compare.Compare(expectedBody, result.ResponseBody, compare.Params{}))...)
		return errs, nil
	}

	// decode actual body
	if responseErr != nil {
		return []error{
			colorize.NewEntityError("parse service %s as "+typeName, "response body").WithSubError(responseErr),
		}, nil
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
