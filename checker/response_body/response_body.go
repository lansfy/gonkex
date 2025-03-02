package response_body

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/lansfy/gonkex/checker"
	"github.com/lansfy/gonkex/colorize"
	"github.com/lansfy/gonkex/compare"
	"github.com/lansfy/gonkex/models"
	"github.com/lansfy/gonkex/xmlparsing"

	"sigs.k8s.io/yaml"
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
	if _, ok := compare.StringAsRegexp(expectedBody); ok {
		return addMainError(compare.Compare(expectedBody, result.ResponseBody, compare.Params{})), nil
	}

	if expectedBody != "" {
		switch {
		case isJSONResponseBody(result):
			return compareBody(t, expectedBody, result, "JSON", decodeJSON)
		case isXMLResponseBody(result):
			return compareBody(t, expectedBody, result, "XML", decodeXML)
		case isYAMLResponseBody(result):
			return compareBody(t, expectedBody, result, "YAML", decodeYAML)
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

func decodeJSON(body string) (interface{}, error) {
	var expected interface{}
	err := json.Unmarshal([]byte(body), &expected)
	return expected, err
}

func decodeXML(body string) (interface{}, error) {
	return xmlparsing.Parse(body)
}

func decodeYAML(body string) (interface{}, error) {
	jsonBody, err := yaml.YAMLToJSON([]byte(body))
	if err != nil {
		return nil, err
	}
	return decodeJSON(string(jsonBody))
}

func getCompareParams(t models.TestInterface) compare.Params {
	params := t.GetComparisonParams()
	return compare.Params{
		IgnoreValues:         params.IgnoreValuesChecking(),
		IgnoreArraysOrdering: params.IgnoreArraysOrdering(),
		DisallowExtraFields:  params.DisallowExtraFields(),
	}
}
