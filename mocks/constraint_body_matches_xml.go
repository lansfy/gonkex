package mocks

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/lansfy/gonkex/compare"
	"github.com/lansfy/gonkex/xmlparsing"
)

func loadBodyMatchesXMLConstraint(def map[interface{}]interface{}) (verifier, error) {
	c, ok := def["body"]
	if !ok {
		return nil, errors.New("`bodyMatchesXML` requires `body` key")
	}
	body, ok := c.(string)
	if !ok {
		return nil, errors.New("`body` must be string")
	}

	params, err := readCompareParams(def)
	if err != nil {
		return nil, err
	}

	return newBodyMatchesXMLConstraint(body, params)
}

func newBodyMatchesXMLConstraint(expected string, params compare.Params) (verifier, error) {
	expectedBody, err := xmlparsing.Parse(expected)
	if err != nil {
		return nil, err
	}

	res := &bodyMatchesXMLConstraint{
		expectedBody:  expectedBody,
		compareParams: params,
	}
	return res, nil
}

type bodyMatchesXMLConstraint struct {
	expectedBody  interface{}
	compareParams compare.Params
}

func (c *bodyMatchesXMLConstraint) Verify(r *http.Request) []error {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return []error{err}
	}
	// write body for future reusing
	r.Body = ioutil.NopCloser(bytes.NewReader(body))
	if len(body) == 0 {
		return []error{fmt.Errorf("request is empty")}
	}

	actual, err := xmlparsing.Parse(string(body))
	if err != nil {
		return []error{err}
	}

	return compare.Compare(c.expectedBody, actual, c.compareParams)
}

func (c *bodyMatchesXMLConstraint) Fields() []string {
	return []string{"body", "comparisonParams"}
}
