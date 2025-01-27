package response_body

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/lansfy/gonkex/checker"
	"github.com/lansfy/gonkex/colorize"
	"github.com/lansfy/gonkex/compare"
	"github.com/lansfy/gonkex/models"
	"github.com/lansfy/gonkex/xmlparsing"
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

	switch {
	// is the response JSON document?
	case strings.Contains(result.ResponseContentType, "json") && expectedBody != "":
		return compareJsonBody(t, expectedBody, result)
	// is the response XML document?
	case strings.Contains(result.ResponseContentType, "xml") && expectedBody != "":
		return compareXmlBody(t, expectedBody, result)
	default:
		// compare bodies as leaf nodes
		return compare.Compare(expectedBody, result.ResponseBody, compare.Params{}), nil
	}
}

func createWrongStatusError(statusCode int, known map[int]string) error {
	knownCodes := []string{}
	for code := range known {
		knownCodes = append(knownCodes, strconv.Itoa(code))
	}
	return colorize.NewNotEqualError("server responded with unexpected %s:", "status", strings.Join(knownCodes, " / "), statusCode)
}

func compareJsonBody(t models.TestInterface, expectedBody string, result *models.Result) ([]error, error) {
	// decode expected body
	var expected interface{}
	if err := json.Unmarshal([]byte(expectedBody), &expected); err != nil {
		return nil, fmt.Errorf("invalid JSON in response in the test declaration (for status %d): %w", result.ResponseStatusCode, err)
	}

	// decode actual body
	var actual interface{}
	if err := json.Unmarshal([]byte(result.ResponseBody), &actual); err != nil {
		return []error{errors.New("could not parse service response as JSON")}, nil
	}

	return compare.Compare(expected, actual, getCompareParams(t)), nil
}

func compareXmlBody(t models.TestInterface, expectedBody string, result *models.Result) ([]error, error) {
	// decode expected body
	expected, err := xmlparsing.Parse(expectedBody)
	if err != nil {
		return nil, fmt.Errorf("invalid XML in response in the test declaration (for status %d): %w", result.ResponseStatusCode, err)
	}

	// decode actual body
	actual, err := xmlparsing.Parse(result.ResponseBody)
	if err != nil {
		return []error{errors.New("could not parse service response as XML")}, nil
	}

	return compare.Compare(expected, actual, getCompareParams(t)), nil
}

func getCompareParams(t models.TestInterface) compare.Params {
	params := t.GetComparisonParams()
	return compare.Params{
		IgnoreValues:         params.IgnoreValuesChecking(),
		IgnoreArraysOrdering: params.IgnoreArraysOrdering(),
		DisallowExtraFields:  params.DisallowExtraFields(),
	}
}
