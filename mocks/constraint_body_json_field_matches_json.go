package mocks

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/lansfy/gonkex/compare"
	"github.com/tidwall/gjson"
)

func loadBodyJSONFieldMatchesJSONConstraint(def map[interface{}]interface{}) (verifier, error) {
	path, err := getRequiredStringKey(def, "path", false)
	if err != nil {
		return nil, err
	}
	value, err := getRequiredStringKey(def, "value", true)
	if err != nil {
		return nil, err
	}
	params, err := readCompareParams(def)
	if err != nil {
		return nil, err
	}

	return newBodyJSONFieldMatchesJSONConstraint(path, value, params)
}

func newBodyJSONFieldMatchesJSONConstraint(path, expected string, params compare.Params) (verifier, error) {
	var v interface{}
	err := json.Unmarshal([]byte(expected), &v)
	if err != nil {
		return nil, err
	}
	res := &bodyJSONFieldMatchesJSONConstraint{
		path:          path,
		expected:      v,
		compareParams: params,
	}
	return res, nil
}

type bodyJSONFieldMatchesJSONConstraint struct {
	path          string
	expected      interface{}
	compareParams compare.Params
}

func (c *bodyJSONFieldMatchesJSONConstraint) Verify(r *http.Request) []error {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return []error{err}
	}

	// write body for future reusing
	r.Body = ioutil.NopCloser(bytes.NewReader(body))

	value := gjson.Get(string(body), c.path)
	if !value.Exists() {
		return []error{fmt.Errorf("json field %s does not exist", c.path)}
	}
	if value.String() == "" {
		return []error{fmt.Errorf("json field %s is empty", c.path)}
	}

	var actual interface{}
	err = json.Unmarshal([]byte(value.String()), &actual)
	if err != nil {
		return []error{err}
	}
	return compare.Compare(c.expected, actual, c.compareParams)
}
