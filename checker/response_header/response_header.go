package response_header

import (
	"net/textproto"

	"github.com/lansfy/gonkex/checker"
	"github.com/lansfy/gonkex/colorize"
	"github.com/lansfy/gonkex/models"
)

func NewChecker() checker.CheckerInterface {
	return &responseHeaderChecker{}
}

type responseHeaderChecker struct{}

func (c *responseHeaderChecker) Check(t models.TestInterface, result *models.Result) ([]error, error) {
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
			errs = append(errs, colorize.NewError(
				colorize.None("response does not include expected header "),
				colorize.Cyan(k),
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
			errs = append(errs, colorize.NewNotEqualError(
				"response header ", k, " value does not match:",
				v,
				actualValues[0],
				nil,
			))
		} else {
			errs = append(errs, colorize.NewError(
				colorize.None("response header "),
				colorize.Cyan(k),
				colorize.None(" value does not match expected "),
				colorize.Green(v),
			))
		}
	}

	return errs, nil
}
