package mocks

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
)

func (l *loaderImpl) loadMethodVaryStrategy(path string, def map[interface{}]interface{}) (ReplyStrategy, error) {
	u, ok := def["methods"]
	if !ok {
		return nil, errors.New("'methods' key required")
	}

	methodsMap, ok := u.(map[interface{}]interface{})
	if !ok {
		return nil, errors.New("map under 'methods' key required")
	}

	methods := map[string]*Definition{}
	for method, v := range methodsMap {
		methodStr, ok := method.(string)
		if !ok {
			return nil, fmt.Errorf("method '%v' has non-string name", method)
		}
		def, err := l.loadDefinition(path+"."+methodStr, v)
		if err != nil {
			return nil, err
		}
		methods[methodStr] = def
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
