package mocks

import (
	"errors"
	"net/http"
	"strconv"
	"sync"
)

func loadBasedOnRequestReplyStrategy(path string, def map[interface{}]interface{}) (ReplyStrategy, error) {
	var uris []*Definition
	if u, ok := def["uris"]; ok {
		urisList, ok := u.([]interface{})
		if !ok {
			return nil, errors.New("list under `uris` key required")
		}
		uris = make([]*Definition, 0, len(urisList))
		for i, v := range urisList {
			v, ok := v.(map[interface{}]interface{})
			if !ok {
				return nil, errors.New("`uris` list item must be a map")
			}
			def, err := loadDefinition(path+"."+strconv.Itoa(i), v)
			if err != nil {
				return nil, err
			}
			uris = append(uris, def)
		}
	}
	return NewBasedOnRequestReply(uris), nil
}

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
		if len(errs) == 0 {
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
