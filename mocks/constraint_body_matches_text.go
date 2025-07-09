package mocks

import (
	"net/http"

	"github.com/lansfy/gonkex/compare"
)

func loadBodyMatchesTextConstraint(def map[interface{}]interface{}) (verifier, error) {
	bodyStr, err := getOptionalStringKey(def, "body", true)
	if err != nil {
		return nil, err
	}
	regexpStr, err := getOptionalStringKey(def, "regexp", false)
	if err != nil {
		return nil, err
	}

	if regexpStr != "" {
		bodyStr = compare.MatchRegexpWrap(regexpStr)
	}

	return newBodyMatchesTextConstraint(bodyStr), nil
}

func newBodyMatchesTextConstraint(body string) verifier {
	return &bodyMatchesTextConstraint{
		body: body,
	}
}

type bodyMatchesTextConstraint struct {
	body string
}

func (c *bodyMatchesTextConstraint) GetName() string {
	return "bodyMatchesText"
}

func (c *bodyMatchesTextConstraint) Verify(r *http.Request) []error {
	body, err := getRequestBodyCopy(r)
	if err != nil {
		return []error{err}
	}

	return compareValues("request %s", "body", c.body, string(body))
}
