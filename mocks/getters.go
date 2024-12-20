package mocks

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/lansfy/gonkex/compare"
)

func wrongTypeError(key, typeName string) error {
	return fmt.Errorf("key '%s' has non-%s value", key, typeName)
}

func getRequiredStringKey(def map[interface{}]interface{}, name string, allowedEmpty bool) (string, error) {
	f, ok := def[name]
	if !ok {
		return "", fmt.Errorf("'%s' key required", name)
	}
	value, ok := f.(string)
	if !ok {
		return "", wrongTypeError(name, "string")
	}

	if !allowedEmpty && value == "" {
		return "", fmt.Errorf("'%s' value can't be empty", name)
	}
	return value, nil
}

func getOptionalStringKey(def map[interface{}]interface{}, name string, allowedEmpty bool) (string, error) {
	f, ok := def[name]
	if !ok {
		return "", nil
	}
	value, ok := f.(string)
	if !ok {
		return "", wrongTypeError(name, "string")
	}
	if !allowedEmpty && value == "" {
		return "", fmt.Errorf("'%s' value can't be empty", name)
	}
	return value, nil
}

func getOptionalIntKey(def map[interface{}]interface{}, name string, defaultValue int) (int, error) {
	if c, ok := def[name]; ok {
		value, ok := c.(int)
		if !ok {
			return 0, wrongTypeError(name, "integer")
		}
		return value, nil
	}
	return defaultValue, nil
}

func loadHeaders(def map[interface{}]interface{}) (map[string]string, error) {
	var headers map[string]string
	if h, ok := def["headers"]; ok {
		hMap, ok := h.(map[interface{}]interface{})
		if !ok {
			return nil, errors.New("'headers' must be a map")
		}
		headers = make(map[string]string, len(hMap))
		for k, v := range hMap {
			key, ok := k.(string)
			if !ok {
				return nil, errors.New("'headers' requires string keys")
			}
			value, ok := v.(string)
			if !ok {
				return nil, errors.New("'headers' requires string values")
			}
			headers[key] = value
		}
	}
	return headers, nil
}

func readCompareParams(def map[interface{}]interface{}) (compare.Params, error) {
	params := compare.Params{
		IgnoreArraysOrdering: true,
	}

	p, ok := def["comparisonParams"]
	if !ok {
		return params, nil
	}

	wrap := func(err error) error {
		return fmt.Errorf("section 'comparisonParams': %w", err)
	}

	values, ok := p.(map[interface{}]interface{})
	if !ok {
		return params, wrap(errors.New("section can't be parsed"))
	}

	mapping := map[string]*bool{
		"ignoreValues":         &params.IgnoreValues,
		"ignoreArraysOrdering": &params.IgnoreArraysOrdering,
		"disallowExtraFields":  &params.DisallowExtraFields,
	}
	allowedKeys := []string{"ignoreValues", "ignoreArraysOrdering", "disallowExtraFields"}

	for key, val := range values {
		skey, ok := key.(string)
		if !ok {
			return params, wrap(fmt.Errorf("key '%v' has non-string type", key))
		}

		bval, ok := val.(bool)
		if !ok {
			return params, wrap(wrongTypeError(skey, "bool"))
		}

		if pbval, ok := mapping[skey]; ok {
			*pbval = bval
		} else {
			return params, wrap(fmt.Errorf("unexpected key '%s' (allowed only %v)", skey, allowedKeys))
		}
	}
	return params, nil
}

func getBodyCopy(r *http.Request) ([]byte, error) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}

	// write body for future reusing
	r.Body = ioutil.NopCloser(bytes.NewReader(body))
	return body, nil
}
