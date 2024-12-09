package mocks

import (
	"bytes"
	"fmt"
	"io/ioutil"
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

func (c *bodyMatchesTextConstraint) Verify(r *http.Request) []error {
	ioBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return []error{err}
	}

	// write body for future reusing
	r.Body = ioutil.NopCloser(bytes.NewReader(ioBody))

	body := string(ioBody)

	if c.body != "" && c.body != body {
		return []error{fmt.Errorf("body value\n%s\ndoesn't match expected\n%s", body, c.body)}
	}
	if c.regexp != nil && !c.regexp.MatchString(body) {
		return []error{fmt.Errorf("body value\n%s\ndoesn't match regexp %s", body, c.regexp)}
	}
	return nil
}
