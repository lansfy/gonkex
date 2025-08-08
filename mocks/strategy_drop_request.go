package mocks

import (
	"net/http"
)

func (l *loaderImpl) loadDropRequestStrategy() (ReplyStrategy, error) {
	return NewDropRequestReply(), nil
}

func NewDropRequestReply() ReplyStrategy {
	return &dropRequestReply{}
}

type dropRequestReply struct{}

func (s *dropRequestReply) HandleRequest(w http.ResponseWriter, r *http.Request) []error {
	wrap, ok := w.(*wrapResponseWriter)
	if ok {
		wrap.drop = true
	}
	return nil
}
