package mocks

import (
	"net/http"
	"sync"
)

func NewBasedOnRequestReply(variants []*Definition) ReplyStrategy {
	return &basedOnRequestReply{
		variants: variants,
	}
}

var _ contextAwareStrategy = (*basedOnRequestReply)(nil)

type basedOnRequestReply struct {
	mutex    sync.Mutex
	variants []*Definition
}

func (s *basedOnRequestReply) HandleRequest(w http.ResponseWriter, r *http.Request) []error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	var errors []error
	for _, def := range s.variants {
		errs := verifyRequestConstraints(def.requestConstraints, r)
		if errs == nil {
			return def.ExecuteWithoutVerifying(w, r)
		}
		errors = append(errors, errs...)
	}
	return append(errors, unhandledRequestError(r)...)
}

func (s *basedOnRequestReply) ResetRunningContext() {
	for _, def := range s.variants {
		def.ResetRunningContext()
	}
}

func (s *basedOnRequestReply) EndRunningContext() []error {
	var errs []error
	for _, def := range s.variants {
		errs = append(errs, def.EndRunningContext()...)
	}
	return errs
}
