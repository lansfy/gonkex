package mocks

import (
	"errors"
	"fmt"
	"text/template"

	"github.com/lansfy/gonkex/colorize"
	"github.com/lansfy/gonkex/types"
)

type Loader interface {
	LoadDefinition(rawDef interface{}) (*Definition, error)
}

type YamlLoaderOpts struct {
	TemplateReplyFuncs template.FuncMap
}

func NewYamlLoader(opts *YamlLoaderOpts) Loader {
	var funcs template.FuncMap
	if opts != nil {
		funcs = opts.TemplateReplyFuncs
	}
	return &loaderImpl{
		templateReplyFuncs: funcs,
		order:              newOrderChecker(),
	}
}

type loaderImpl struct {
	templateReplyFuncs template.FuncMap
	order              *orderChecker
}

func (l *loaderImpl) LoadDefinition(rawDef interface{}) (*Definition, error) {
	def, err := l.loadDefinition("$", rawDef)
	if err != nil {
		return nil, err
	}
	return def, nil
}

func (l *loaderImpl) loadDefinition(path string, rawDef interface{}) (*Definition, error) {
	wrapPath := func(path string, err error) error {
		return colorize.NewEntityError("path %s", path).SetSubError(err)
	}

	def, err := loadStringMap(rawDef, "")
	if err != nil {
		return nil, wrapPath(path, err)
	}

	// load reply strategy
	strategyName, err := getRequiredStringKey(def, "strategy", false)
	if err != nil {
		return nil, wrapPath(path, err)
	}

	wrap := func(err error) error {
		if colorize.HasPathComponent(err) {
			return err
		}
		err = colorize.NewEntityError("strategy %s", strategyName).SetSubError(err)
		if path == "$" {
			return err
		}
		return colorize.NewEntityError("path %s", path).SetSubError(err)
	}

	// load request constraints
	var requestConstraints []verifier
	if constraints, ok := def["requestConstraints"]; ok {
		constraints, ok := constraints.([]interface{})
		if !ok {
			return nil, wrapPath(path+".requestConstraints", colorize.NewEntityError("%s must be array", "requestConstraints"))
		}
		requestConstraints = []verifier{}
		for i, c := range constraints {
			constraint, err := loadConstraint(c)
			if err != nil {
				return nil, wrapPath(fmt.Sprintf("%s.requestConstraints[%d]", path, i), err)
			}
			requestConstraints = append(requestConstraints, constraint)
		}
	}

	ak := []string{
		"requestConstraints",
		"strategy",
		"calls",
		"order",
	}

	replyStrategy, err := l.loadStrategy(path, strategyName, def, &ak)
	if err != nil {
		return nil, wrap(err)
	}
	callsConstraint, err := getOptionalIntKey(def, "calls", CallsNoConstraint)
	if err != nil {
		return nil, wrap(err)
	}
	orderValue, err := getOptionalIntKey(def, "order", OrderNoValue)
	if err != nil {
		return nil, wrap(err)
	}
	if err := validateMapKeys(def, ak); err != nil {
		return nil, wrap(err)
	}

	res := NewDefinition(path, requestConstraints, replyStrategy, callsConstraint, orderValue)
	res.order = l.order
	return res, nil
}

func (l *loaderImpl) loadStrategy(path, strategyName string, definition map[string]interface{},
	ak *[]string) (ReplyStrategy, error) {
	switch strategyName {
	case "nop":
		return NewNopReply(), nil
	case "constant":
		*ak = append(*ak, "body", "statusCode", "headers")
		return l.loadConstantStrategy(definition)
	case "sequence":
		*ak = append(*ak, "sequence")
		return l.loadSequenceReplyStrategy(path, definition)
	case "template":
		*ak = append(*ak, "body", "statusCode", "headers")
		return l.loadTemplateReplyStrategy(definition)
	case "basedOnRequest":
		*ak = append(*ak, "basePath", "uris")
		return l.loadBasedOnRequestReplyStrategy(path, definition)
	case "file":
		*ak = append(*ak, "filename", "statusCode", "headers")
		return l.loadFileStrategy(definition)
	case "uriVary":
		*ak = append(*ak, "basePath", "uris")
		return l.loadUriVaryReplyStrategy(path, definition)
	case "methodVary":
		*ak = append(*ak, "methods")
		return l.loadMethodVaryStrategy(path, definition)
	case "dropRequest":
		return l.loadDropRequestStrategy()
	default:
		return nil, errors.New("unknown strategy")
	}
}

func loadConstraint(definition interface{}) (verifier, error) {
	wrap := func(err error) error {
		return colorize.NewError("load constraint").SetSubError(err)
	}

	def, err := loadStringMap(definition, "")
	if err != nil {
		return nil, wrap(err)
	}
	kind, err := getRequiredStringKey(def, "kind", false)
	if err != nil {
		return nil, wrap(err)
	}

	ak := []string{"kind"}

	wrap = func(err error) error {
		return colorize.NewEntityError("load constraint %s", kind).SetSubError(err)
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

func loadConstraintOfKind(kind string, def map[string]interface{}, ak *[]string) (verifier, error) {
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
	case "methodIsPUT":
		return &methodConstraint{name: kind, method: "PUT"}, nil
	case "methodIsDELETE":
		return &methodConstraint{name: kind, method: "DELETE"}, nil
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
	case "bodyJSONFieldMatchesJSON":
		*ak = append(*ak, "path", "value", "comparisonParams")
		return loadBodyJSONFieldMatchesJSONConstraint(def)
	}

	for _, b := range types.GetRegisteredBodyTypes() {
		if kind == "bodyMatches"+b.GetName() {
			*ak = append(*ak, "body", "comparisonParams")
			return loadBodyMatchesConstraint(def, b)
		}
	}

	return nil, errors.New("unknown constraint")
}

func validateMapKeys(m map[string]interface{}, allowedKeys []string) error {
	for skey := range m {
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
