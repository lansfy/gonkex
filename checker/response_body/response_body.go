package response_body

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/lansfy/gonkex/checker"
	"github.com/lansfy/gonkex/compare"
	"github.com/lansfy/gonkex/models"
	"github.com/lansfy/gonkex/xmlparsing"
)

type ResponseBodyChecker struct{}

func NewChecker() checker.CheckerInterface {
	return &ResponseBodyChecker{}
}

func (c *ResponseBodyChecker) Check(t models.TestInterface, result *models.Result) ([]error, error) {
	var errs []error
	var foundResponse bool
	// test response with the expected response body
	if expectedBody, ok := t.GetResponse(result.ResponseStatusCode); ok {
		foundResponse = true
		switch {
		// is the response JSON document?
		case strings.Contains(result.ResponseContentType, "json") && expectedBody != "":
			checkErrs, err := compareJsonBody(t, expectedBody, result)
			if err != nil {
				return nil, err
			}
			errs = append(errs, checkErrs...)
		// is the response XML document?
		case strings.Contains(result.ResponseContentType, "xml") && expectedBody != "":
			checkErrs, err := compareXmlBody(t, expectedBody, result)
			if err != nil {
				return nil, err
			}
			errs = append(errs, checkErrs...)
		default:
			// compare bodies as leaf nodes
			errs = append(errs, compare.Compare(expectedBody, result.ResponseBody, compare.Params{})...)
		}
	}
	if !foundResponse {
		err := fmt.Errorf("server responded with status %d", result.ResponseStatusCode)
		errs = append(errs, err)
	}

	return errs, nil
}

func compareJsonBody(t models.TestInterface, expectedBody string, result *models.Result) ([]error, error) {
	// decode expected body
	var expected interface{}
	if err := json.Unmarshal([]byte(expectedBody), &expected); err != nil {
		return nil, fmt.Errorf(
			"invalid JSON in response for test %s (status %d): %s",
			t.GetName(),
			result.ResponseStatusCode,
			err.Error(),
		)
	}

	// decode actual body
	var actual interface{}
	if err := json.Unmarshal([]byte(result.ResponseBody), &actual); err != nil {
		return []error{errors.New("could not parse response")}, nil
	}

	cmpOptions := t.GetComparisonParams()

	params := compare.Params{
		IgnoreValues:         cmpOptions.IgnoreValuesChecking(),
		IgnoreArraysOrdering: cmpOptions.IgnoreArraysOrdering(),
		DisallowExtraFields:  cmpOptions.DisallowExtraFields(),
	}

	return compare.Compare(expected, actual, params), nil
}

func compareXmlBody(t models.TestInterface, expectedBody string, result *models.Result) ([]error, error) {
	// decode expected body
	expected, err := xmlparsing.Parse(expectedBody)
	if err != nil {
		return nil, fmt.Errorf(
			"invalid XML in response for test %s (status %d): %s",
			t.GetName(),
			result.ResponseStatusCode,
			err.Error(),
		)
	}

	// decode actual body
	actual, err := xmlparsing.Parse(result.ResponseBody)
	if err != nil {
		return []error{errors.New("could not parse response")}, nil
	}

	cmpOptions := t.GetComparisonParams()

	params := compare.Params{
		IgnoreValues:         cmpOptions.IgnoreValuesChecking(),
		IgnoreArraysOrdering: cmpOptions.IgnoreArraysOrdering(),
		DisallowExtraFields:  cmpOptions.DisallowExtraFields(),
	}

	return compare.Compare(expected, actual, params), nil
}
