package mocks

import (
	"net/http"
	"sync"
)

func NewSequentialReply(strategies []*Definition) ReplyStrategy {
	return &sequentialReply{
		sequence: strategies,
	}
}

var _ contextAwareStrategy = (*sequentialReply)(nil)

type sequentialReply struct {
	mutex    sync.Mutex
	count    int
	sequence []*Definition
}

func (s *sequentialReply) ResetRunningContext() {
	s.mutex.Lock()
	s.count = 0
	s.mutex.Unlock()
	for _, def := range s.sequence {
		def.ResetRunningContext()
	}
}

func (s *sequentialReply) EndRunningContext() []error {
	var errs []error
	for _, def := range s.sequence {
		errs = append(errs, def.EndRunningContext()...)
	}
	return errs
}

func (s *sequentialReply) HandleRequest(w http.ResponseWriter, r *http.Request) []error {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	// out of bounds, url requested more times than sequence length
	if s.count >= len(s.sequence) {
		return unhandledRequestError(r)
	}
	def := s.sequence[s.count]
	s.count++
	return def.Execute(w, r)
}
