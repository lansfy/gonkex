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
				"response does not include expected header %s",
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
				"response header %s value does not match:",
				k,
				v,
				actualValues[0],
			))
		} else {
			errs = append(errs, colorize.NewError(
				"response header %s value does not match expected %s",
				colorize.Cyan(k),
				colorize.Green(v),
			))
		}
	}

	return errs, nil
}
