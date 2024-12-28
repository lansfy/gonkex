package response_header

import (
	"fmt"
	"net/textproto"

	"github.com/lansfy/gonkex/checker"
	"github.com/lansfy/gonkex/models"

	"github.com/fatih/color"
)

type ResponseHeaderChecker struct{}

func NewChecker() checker.CheckerInterface {
	return &ResponseHeaderChecker{}
}

func (c *ResponseHeaderChecker) Check(t models.TestInterface, result *models.Result) ([]error, error) {
	// test response headers with the expected headers
	expectedHeaders, ok := t.GetResponseHeaders(result.ResponseStatusCode)
	if !ok || len(expectedHeaders) == 0 {
		return nil, nil
	}

	var errs []error
	for k, v := range expectedHeaders {
		k = textproto.CanonicalMIMEHeaderKey(k)
		actualValues, ok := result.ResponseHeaders[k]
		if !ok {
			errs = append(errs, fmt.Errorf(
				"response does not include expected header %s",
				color.CyanString(k),
			))
			continue
		}
		found := false
		for _, actualValue := range actualValues {
			if v == actualValue {
				found = true
				break
			}
		}
		if found {
			continue
		}
		if len(actualValues) == 1 {
			errs = append(errs, fmt.Errorf(
				"response header %s value does not match:\n     expected: %s\n       actual: %s",
				color.CyanString(k),
				color.GreenString("%s", v),
				color.RedString("%v", actualValues[0]),
			))
		} else {
			errs = append(errs, fmt.Errorf(
				"response header %s value does not match expected %s",
				color.CyanString(k),
				color.GreenString("%s", v),
			))
		}
	}

	return errs, nil
}
