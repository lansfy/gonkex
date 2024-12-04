package mocks

import (
	"bytes"
	"errors"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/lansfy/gonkex/compare"
	"github.com/tidwall/gjson"
)

func loadBodyJSONFieldMatchesJSONConstraint(def map[interface{}]interface{}) (verifier, error) {
	c, ok := def["path"]
	if !ok {
		return nil, errors.New("`bodyJSONFieldMatchesJSON` requires `path` key")
	}
	path, ok := c.(string)
	if !ok {
		return nil, errors.New("`path` must be string")
	}

	c, ok = def["value"]
	if !ok {
		return nil, errors.New("`bodyJSONFieldMatchesJSON` requires `value` key")
	}
	value, ok := c.(string)
	if !ok {
		return nil, errors.New("`value` must be string")
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

func (c *bodyJSONFieldMatchesJSONConstraint) Fields() []string {
	return []string{"path", "value", "comparisonParams"}
}
