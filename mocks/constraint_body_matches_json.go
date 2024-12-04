package mocks

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/lansfy/gonkex/compare"
)

func loadBodyMatchesJSONConstraint(def map[interface{}]interface{}) (verifier, error) {
	c, ok := def["body"]
	if !ok {
		return nil, errors.New("`bodyMatchesJSON` requires `body` key")
	}
	body, ok := c.(string)
	if !ok {
		return nil, errors.New("`body` must be string")
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

func (c *bodyMatchesJSONConstraint) Verify(r *http.Request) []error {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return []error{err}
	}
	// write body for future reusing
	r.Body = ioutil.NopCloser(bytes.NewReader(body))
	if len(body) == 0 {
		return []error{fmt.Errorf("request is empty")}
	}
	var actual interface{}
	err = json.Unmarshal(body, &actual)
	if err != nil {
		return []error{err}
	}
	return compare.Compare(c.expectedBody, actual, c.compareParams)
}

func (c *bodyMatchesJSONConstraint) Fields() []string {
	return []string{"body", "comparisonParams"}
}
