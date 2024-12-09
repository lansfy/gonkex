package mocks

import (
	"net/http"
)

func NewNopReply() ReplyStrategy {
	return &nopReply{}
}

type nopReply struct{}

func (s *nopReply) HandleRequest(w http.ResponseWriter, r *http.Request) []error {
	w.WriteHeader(http.StatusNoContent)
	return nil
}
