package mocks

import (
	"fmt"
	"net/http"
	"regexp"
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
	return newBodyMatchesTextConstraint(bodyStr, regexpStr)
}

func newBodyMatchesTextConstraint(body, re string) (verifier, error) {
	var reCompiled *regexp.Regexp
	if re != "" {
		var err error
		reCompiled, err = regexp.Compile(re)
		if err != nil {
			return nil, err
		}
	}
	res := &bodyMatchesTextConstraint{
		body:   body,
		regexp: reCompiled,
	}
	return res, nil
}

type bodyMatchesTextConstraint struct {
	body   string
	regexp *regexp.Regexp
}

func (c *bodyMatchesTextConstraint) GetName() string {
	return "bodyMatchesText"
}

func (c *bodyMatchesTextConstraint) Verify(r *http.Request) []error {
	body, err := getBodyCopy(r)
	if err != nil {
		return []error{err}
	}

	textBody := string(body)

	if c.body != "" && c.body != textBody {
		return []error{fmt.Errorf("body value\n%s\ndoesn't match expected\n%s", textBody, c.body)}
	}
	if c.regexp != nil && !c.regexp.MatchString(textBody) {
		return []error{fmt.Errorf("body value\n%s\ndoesn't match regexp %s", textBody, c.regexp)}
	}
	return nil
}
