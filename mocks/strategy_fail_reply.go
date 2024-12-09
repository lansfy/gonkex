package mocks

import (
	"net/http"
)

func NewFailReply() ReplyStrategy {
	return &failReply{}
}

type failReply struct{}

func (s *failReply) HandleRequest(w http.ResponseWriter, r *http.Request) []error {
	return unhandledRequestError(r)
}
