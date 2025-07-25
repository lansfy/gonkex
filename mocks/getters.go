package mocks

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/lansfy/gonkex/colorize"
	"github.com/lansfy/gonkex/compare"
)

func wrongTypeError(key, typeName string) error {
	return fmt.Errorf("key '%s' has non-%s value", key, typeName)
}

func getRequiredStringKey(def map[string]interface{}, name string, allowedEmpty bool) (string, error) {
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

func getOptionalStringKey(def map[string]interface{}, name string, allowedEmpty bool) (string, error) {
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

func getOptionalIntKey(def map[string]interface{}, name string, defaultValue int) (int, error) {
	c, ok := def[name]
	if !ok {
		return defaultValue, nil
	}

	var err error
	var parsedValue int
	switch v := c.(type) {
	case int:
		parsedValue = v
	case string:
		parsedValue, err = strconv.Atoi(v)
	default:
		err = errors.New("fake")
	}
	if err != nil {
		return 0, fmt.Errorf("value for key '%s' cannot be converted to integer", name)
	}
	if parsedValue < 0 {
		return 0, fmt.Errorf("value for the key '%s' cannot be negative", name)
	}
	return parsedValue, nil
}

func loadStringMap(def interface{}, key string) (map[string]interface{}, error) {
	if sMap, ok := def.(map[string]interface{}); ok {
		return sMap, nil
	}

	if iMap, ok := def.(map[interface{}]interface{}); ok {
		result := map[string]interface{}{}
		for ikey, ivalue := range iMap {
			skey, ok := ikey.(string)
			if !ok {
				return nil, fmt.Errorf("key '%v' has non-string type", ikey)
			}
			result[skey] = ivalue
		}
		return result, nil
	}

	if key == "" {
		return nil, errors.New("must be a map")
	}

	return nil, colorize.NewEntityError("map under %s key is required", key)
}

func loadHeaders(def map[string]interface{}) (map[string]string, error) {
	h, ok := def["headers"]
	if !ok {
		return nil, nil
	}

	hMap, err := loadStringMap(h, "headers")
	if err != nil {
		return nil, err
	}

	headers := map[string]string{}
	for key, v := range hMap {
		value, ok := v.(string)
		if !ok {
			return nil, errors.New("'headers' requires string values")
		}
		headers[key] = value
	}
	return headers, nil
}

func readCompareParams(def map[string]interface{}) (compare.Params, error) {
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

	values, err := loadStringMap(p, "")
	if err != nil {
		return params, wrap(err)
	}

	mapping := map[string]*bool{
		"ignoreValues":         &params.IgnoreValues,
		"ignoreArraysOrdering": &params.IgnoreArraysOrdering,
		"disallowExtraFields":  &params.DisallowExtraFields,
	}
	allowedKeys := []string{"ignoreValues", "ignoreArraysOrdering", "disallowExtraFields"}

	for skey, val := range values {
		bval, ok := val.(bool)
		if !ok {
			return params, wrap(wrongTypeError(skey, "bool"))
		}

		pbval, ok := mapping[skey]
		if !ok {
			return params, wrap(fmt.Errorf("unexpected key '%s' (allowed only %v)", skey, allowedKeys))
		}
		*pbval = bval
	}
	return params, nil
}

func getRequestBodyCopy(r *http.Request) ([]byte, error) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}

	err = r.Body.Close()
	if err != nil {
		return nil, err
	}

	// write body for future reusing
	setRequestBody(r, body)
	return body, nil
}

func setRequestBody(r *http.Request, body []byte) {
	r.Body = io.NopCloser(bytes.NewReader(body))
}
