package mocks

import (
	"net/http"
	"strings"
)

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
