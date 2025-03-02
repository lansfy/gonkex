package mocks

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/lansfy/gonkex/compare"
	"github.com/lansfy/gonkex/xmlparsing"

	"sigs.k8s.io/yaml"
)

func loadBodyMatchesJSONConstraint(def map[interface{}]interface{}) (verifier, error) {
	return loadBodyMatchesConstraint(def, "bodyMatchesJSON", decodeJSON)
}

func loadBodyMatchesXMLConstraint(def map[interface{}]interface{}) (verifier, error) {
	return loadBodyMatchesConstraint(def, "bodyMatchesXML", decodeXML)
}

func loadBodyMatchesYAMLConstraint(def map[interface{}]interface{}) (verifier, error) {
	return loadBodyMatchesConstraint(def, "bodyMatchesYAML", decodeYAML)
}

func decodeJSON(body string) (interface{}, error) {
	var expected interface{}
	err := json.Unmarshal([]byte(body), &expected)
	return expected, err
}

func decodeXML(body string) (interface{}, error) {
	return xmlparsing.Parse(body)
}

func decodeYAML(body string) (interface{}, error) {
	jsonBody, err := yaml.YAMLToJSON([]byte(body))
	if err != nil {
		return nil, err
	}
	return decodeJSON(string(jsonBody))
}

func loadBodyMatchesConstraint(def map[interface{}]interface{}, name string,
	decode func(body string) (interface{}, error)) (verifier, error) {
	body, err := getRequiredStringKey(def, "body", true)
	if err != nil {
		return nil, err
	}
	params, err := readCompareParams(def)
	if err != nil {
		return nil, err
	}

	return newBodyMatchesConstraint(body, params, name, decode)
}

func newBodyMatchesConstraint(expected string, params compare.Params, name string,
	decode func(body string) (interface{}, error)) (verifier, error) {
	expectedBody, err := decode(expected)
	if err != nil {
		return nil, err
	}
	return &bodyMatchesConstraint{
		name:          name,
		decode:        decode,
		expectedBody:  expectedBody,
		compareParams: params,
	}, nil
}

type bodyMatchesConstraint struct {
	name          string
	decode        func(body string) (interface{}, error)
	expectedBody  interface{}
	compareParams compare.Params
}

func (c *bodyMatchesConstraint) GetName() string {
	return c.name
}

func (c *bodyMatchesConstraint) Verify(r *http.Request) []error {
	body, err := getRequestBodyCopy(r)
	if err != nil {
		return []error{err}
	}

	if len(body) == 0 {
		return []error{errors.New("request is empty")}
	}
	actual, err := c.decode(string(body))
	if err != nil {
		return []error{err}
	}
	return compare.Compare(c.expectedBody, actual, c.compareParams)
}
