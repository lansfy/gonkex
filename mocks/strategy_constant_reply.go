package mocks

import (
	"net/http"
)

func NewConstantReplyWithCode(content []byte, statusCode int, headers map[string]string) ReplyStrategy {
	return &constantReply{
		replyBody:  content,
		statusCode: statusCode,
		headers:    headers,
	}
}

type constantReply struct {
	replyBody  []byte
	statusCode int
	headers    map[string]string
}

func (s *constantReply) HandleRequest(w http.ResponseWriter, r *http.Request) []error {
	for k, v := range s.headers {
		w.Header().Add(k, v)
	}
	w.WriteHeader(s.statusCode)
	w.Write(s.replyBody)
	return nil
}
