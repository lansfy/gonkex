package mocks

import (
	"fmt"
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
	hj, ok := w.(http.Hijacker)
	if !ok {
		return []error{fmt.Errorf("gonkex internal error during drop request: webserver does not support hijacking")}
	}
	conn, _, err := hj.Hijack()
	if err != nil {
		return []error{fmt.Errorf("gonkex internal error during connection hijacking: %w", err)}
	}
	conn.Close()
	return nil
}
