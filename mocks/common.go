package mocks

import (
	"fmt"
)

func getRequiredStringKey(def map[interface{}]interface{}, name string, allowedEmpty bool) (string, error) {
	f, ok := def[name]
	if !ok {
		return "", fmt.Errorf("`%s` key required", name)
	}
	value, ok := f.(string)
	if !ok {
		return "", fmt.Errorf("`%s` must be string", name)
	}

	if !allowedEmpty && value == "" {
		return "", fmt.Errorf("`%s` value can't be empty", name)
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
		return "", fmt.Errorf("`%s` must be string", name)
	}
	if !allowedEmpty && value == "" {
		return "", fmt.Errorf("`%s` value can't be empty", name)
	}
	return value, nil
}

func getOptionalIntKey(def map[interface{}]interface{}, name string, defaultValue int) (int, error) {
	if c, ok := def[name]; ok {
		value, ok := c.(int)
		if !ok {
			return 0, fmt.Errorf("`%s` must be integer", name)
		}
		return value, nil
	}
	return defaultValue, nil
}
