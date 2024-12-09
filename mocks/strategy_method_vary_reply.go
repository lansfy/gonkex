package mocks

import (
	"errors"
	"net/http"
	"strings"
)

func loadMethodVaryStrategy(path string, def map[interface{}]interface{}) (ReplyStrategy, error) {
	var methods map[string]*Definition
	if u, ok := def["methods"]; ok {
		methodsMap, ok := u.(map[interface{}]interface{})
		if !ok {
			return nil, errors.New("map under `methods` key required")
		}
		methods = make(map[string]*Definition, len(methodsMap))
		for method, v := range methodsMap {
			def, err := loadDefinition(path+"."+method.(string), v)
			if err != nil {
				return nil, err
			}
			methods[method.(string)] = def
		}
	}
	return NewMethodVaryReply(methods), nil
}

func NewMethodVaryReply(variants map[string]*Definition) ReplyStrategy {
	return &methodVaryReply{
		variants: variants,
	}
}

var _ contextAwareStrategy = (*methodVaryReply)(nil)

type methodVaryReply struct {
	variants map[string]*Definition
}

func (s *methodVaryReply) HandleRequest(w http.ResponseWriter, r *http.Request) []error {
	for method, def := range s.variants {
		if strings.EqualFold(r.Method, method) {
			return def.Execute(w, r)
		}
	}
	return unhandledRequestError(r)
}

func (s *methodVaryReply) ResetRunningContext() {
	for _, def := range s.variants {
		def.ResetRunningContext()
	}
}

func (s *methodVaryReply) EndRunningContext() []error {
	var errs []error
	for _, def := range s.variants {
		errs = append(errs, def.EndRunningContext()...)
	}
	return errs
}
