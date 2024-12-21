package mocks

import (
	"errors"
	"fmt"
)

func loadDefinition(path string, rawDef interface{}) (*Definition, error) {
	def, ok := rawDef.(map[interface{}]interface{})
	if !ok {
		return nil, fmt.Errorf("at path %s: Definition must be key-values", path)
	}

	// load request constraints
	var requestConstraints []verifier
	if constraints, ok := def["requestConstraints"]; ok {
		constraints, ok := constraints.([]interface{})
		if !ok || len(constraints) == 0 {
			return nil, fmt.Errorf("at path %s: `requestConstraints` requires array", path)
		}
		requestConstraints = make([]verifier, len(constraints))
		for i, c := range constraints {
			constraint, err := loadConstraint(c)
			if err != nil {
				return nil, fmt.Errorf("at path %s: unable to load constraint %d: %w", path, i+1, err)
			}
			requestConstraints[i] = constraint
		}
	}

	ak := []string{
		"requestConstraints",
		"strategy",
		"calls",
	}

	// load reply strategy
	strategyName, err := getRequiredStringKey(def, "strategy", false)
	if err != nil {
		return nil, fmt.Errorf("at path %s: %w", path, err)
	}

	wrap := func(err error) error {
		return fmt.Errorf("strategy '%s': %w", strategyName, err)
	}

	replyStrategy, err := loadStrategy(path, strategyName, def, &ak)
	if err != nil {
		return nil, wrap(err)
	}
	callsConstraint, err := getOptionalIntKey(def, "calls", CallsNoConstraint)
	if err != nil {
		return nil, wrap(err)
	}
	if err := validateMapKeys(def, ak); err != nil {
		return nil, wrap(err)
	}

	return NewDefinition(path, requestConstraints, replyStrategy, callsConstraint), nil
}

func loadStrategy(path, strategyName string, definition map[interface{}]interface{}, ak *[]string) (ReplyStrategy, error) {
	path = path + "." + strategyName
	switch strategyName {
	case "nop":
		return NewNopReply(), nil
	case "constant":
		*ak = append(*ak, "body", "statusCode", "headers")
		return loadConstantStrategy(path, definition)
	case "sequence":
		*ak = append(*ak, "sequence")
		return loadSequenceReplyStrategy(path, definition)
	case "template":
		*ak = append(*ak, "body", "statusCode", "headers")
		return loadTemplateReplyStrategy(path, definition)
	case "basedOnRequest":
		*ak = append(*ak, "basePath", "uris")
		return loadBasedOnRequestReplyStrategy(path, definition)
	case "file":
		*ak = append(*ak, "filename", "statusCode", "headers")
		return loadFileStrategy(path, definition)
	case "uriVary":
		*ak = append(*ak, "basePath", "uris")
		return loadUriVaryReplyStrategy(path, definition)
	case "methodVary":
		*ak = append(*ak, "methods")
		return loadMethodVaryStrategy(path, definition)
	case "dropRequest":
		return loadDropRequestStrategy(path, definition)
	default:
		return nil, errors.New("unknown strategy")
	}
}

func loadConstraint(definition interface{}) (verifier, error) {
	def, ok := definition.(map[interface{}]interface{})
	if !ok {
		return nil, errors.New("must be map")
	}
	kind, err := getRequiredStringKey(def, "kind", false)
	if err != nil {
		return nil, err
	}
	ak := []string{"kind"}

	wrap := func(err error) error {
		return fmt.Errorf("constraint '%s': %w", kind, err)
	}

	c, err := loadConstraintOfKind(kind, def, &ak)
	if err != nil {
		return nil, wrap(err)
	}
	if err := validateMapKeys(def, ak); err != nil {
		return nil, wrap(err)
	}
	return c, nil
}

func loadConstraintOfKind(kind string, def map[interface{}]interface{}, ak *[]string) (verifier, error) {
	switch kind {
	case "nop":
		return &nopConstraint{}, nil
	case "methodIs":
		*ak = append(*ak, "method")
		return loadMethodConstraint(def)
	case "methodIsGET":
		return &methodConstraint{name: kind, method: "GET"}, nil
	case "methodIsPOST":
		return &methodConstraint{name: kind, method: "POST"}, nil
	case "headerIs":
		*ak = append(*ak, "header", "value", "regexp")
		return loadHeaderConstraint(def)
	case "pathMatches":
		*ak = append(*ak, "path", "regexp")
		return loadPathConstraint(def)
	case "queryMatches":
		*ak = append(*ak, "expectedQuery")
		return loadQueryConstraint(def)
	case "queryMatchesRegexp":
		*ak = append(*ak, "expectedQuery")
		return loadQueryRegexpConstraint(def)
	case "bodyMatchesText":
		*ak = append(*ak, "body", "regexp")
		return loadBodyMatchesTextConstraint(def)
	case "bodyMatchesJSON":
		*ak = append(*ak, "body", "comparisonParams")
		return loadBodyMatchesJSONConstraint(def)
	case "bodyMatchesXML":
		*ak = append(*ak, "body", "comparisonParams")
		return loadBodyMatchesXMLConstraint(def)
	case "bodyJSONFieldMatchesJSON":
		*ak = append(*ak, "path", "value", "comparisonParams")
		return loadBodyJSONFieldMatchesJSONConstraint(def)
	default:
		return nil, errors.New("unknown constraint")
	}
}

func validateMapKeys(m map[interface{}]interface{}, allowedKeys []string) error {
	for key := range m {
		skey, ok := key.(string)
		if !ok {
			return fmt.Errorf("key '%v' has non-string type", key)
		}

		found := false
		for _, ak := range allowedKeys {
			if ak == skey {
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("unexpected key '%s' (allowed only %v)", skey, allowedKeys)
		}
	}
	return nil
}
