package mocks

import (
	"errors"

	"github.com/lansfy/gonkex/compare"
)

func readCompareParams(def map[interface{}]interface{}) (compare.Params, error) {
	params := compare.Params{
		IgnoreArraysOrdering: true,
	}

	p, ok := def["comparisonParams"]
	if !ok {
		return params, nil
	}

	values, ok := p.(map[interface{}]interface{})
	if !ok {
		return params, errors.New("`comparisonParams` can't be parsed")
	}

	mapping := map[string]*bool{
		"ignoreValues":         &params.IgnoreValues,
		"ignoreArraysOrdering": &params.IgnoreArraysOrdering,
		"disallowExtraFields":  &params.DisallowExtraFields,
	}

	for key, val := range values {
		skey, ok := key.(string)
		if !ok {
			return params, errors.New("`comparisonParams` has non-string key")
		}

		bval, ok := val.(bool)
		if !ok {
			return params, errors.New("`comparisonParams` has non-bool values")
		}

		if pbval, ok := mapping[skey]; ok {
			*pbval = bval
		}
	}
	return params, nil
}
