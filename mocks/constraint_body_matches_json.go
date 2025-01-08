package mocks

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/lansfy/gonkex/compare"
)

func loadBodyMatchesJSONConstraint(def map[interface{}]interface{}) (verifier, error) {
	body, err := getRequiredStringKey(def, "body", true)
	if err != nil {
		return nil, err
	}
	params, err := readCompareParams(def)
	if err != nil {
		return nil, err
	}

	return newBodyMatchesJSONConstraint(body, params)
}

func newBodyMatchesJSONConstraint(expected string, params compare.Params) (verifier, error) {
	var expectedBody interface{}
	err := json.Unmarshal([]byte(expected), &expectedBody)
	if err != nil {
		return nil, err
	}
	res := &bodyMatchesJSONConstraint{
		expectedBody:  expectedBody,
		compareParams: params,
	}
	return res, nil
}

type bodyMatchesJSONConstraint struct {
	expectedBody  interface{}
	compareParams compare.Params
}

func (c *bodyMatchesJSONConstraint) GetName() string {
	return "bodyMatchesJSON"
}

func (c *bodyMatchesJSONConstraint) Verify(r *http.Request) []error {
	body, err := getRequestBodyCopy(r)
	if err != nil {
		return []error{err}
	}

	if len(body) == 0 {
		return []error{errors.New("request is empty")}
	}
	var actual interface{}
	err = json.Unmarshal(body, &actual)
	if err != nil {
		return []error{err}
	}
	return compare.Compare(c.expectedBody, actual, c.compareParams)
}
