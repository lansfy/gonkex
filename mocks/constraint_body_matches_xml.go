package mocks

import (
	"fmt"
	"net/http"

	"github.com/lansfy/gonkex/compare"
	"github.com/lansfy/gonkex/xmlparsing"
)

func loadBodyMatchesXMLConstraint(def map[interface{}]interface{}) (verifier, error) {
	body, err := getRequiredStringKey(def, "body", true)
	if err != nil {
		return nil, err
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

func (c *bodyMatchesXMLConstraint) GetName() string {
	return "bodyMatchesXML"
}

func (c *bodyMatchesXMLConstraint) Verify(r *http.Request) []error {
	body, err := getBodyCopy(r)
	if err != nil {
		return []error{err}
	}

	if len(body) == 0 {
		return []error{fmt.Errorf("request is empty")}
	}

	actual, err := xmlparsing.Parse(string(body))
	if err != nil {
		return []error{err}
	}

	return compare.Compare(c.expectedBody, actual, c.compareParams)
}
