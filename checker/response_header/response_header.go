package response_header

import (
	"net/textproto"
	"sort"
	"strings"

	"github.com/lansfy/gonkex/checker"
	"github.com/lansfy/gonkex/colorize"
	"github.com/lansfy/gonkex/models"
)

func NewChecker() checker.CheckerInterface {
	return &responseHeaderChecker{}
}

type responseHeaderChecker struct{}

func (c *responseHeaderChecker) Check(t models.TestInterface, result *models.Result) ([]error, error) {
	expectedHeaders, ok := t.GetResponseHeaders(result.ResponseStatusCode)
	if !ok || len(expectedHeaders) == 0 {
		return nil, nil
	}

	result.ShowHeaders = true // output should show request headers

	var errs []error
	for k, v := range expectedHeaders {
		err := checkOneHeader(k, v, result.ResponseHeaders)
		if err != nil {
			errs = append(errs, err)
		}
	}

	return errs, nil
}

func checkOneHeader(key, value string, responseHeaders map[string][]string) error {
	key = textproto.CanonicalMIMEHeaderKey(key)
	actualValues, ok := responseHeaders[key]
	if !ok {
		return colorize.NewError("response does not include expected header %s", colorize.Cyan(key))
	}
	for _, actualValue := range actualValues {
		if value == actualValue {
			return nil
		}
	}
	sort.Strings(actualValues)
	return colorize.NewNotEqualError("response header %s value does not match:", key, value, strings.Join(actualValues, " / "))
}
