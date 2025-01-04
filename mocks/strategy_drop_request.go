package mocks

import (
	"fmt"
	"net/http"
)

func (l *loaderImpl) loadDropRequestStrategy(path string, _ map[interface{}]interface{}) (ReplyStrategy, error) {
	return NewDropRequestReply(), nil
}

func NewDropRequestReply() ReplyStrategy {
	return &dropRequestReply{}
}

type dropRequestReply struct{}

func (s *dropRequestReply) HandleRequest(w http.ResponseWriter, r *http.Request) []error {
	hj, ok := w.(http.Hijacker)
	if !ok {
		return []error{fmt.Errorf("Gonkex internal error during drop request: webserver doesn't support hijacking")}
	}
	conn, _, err := hj.Hijack()
	if err != nil {
		return []error{fmt.Errorf("Gonkex internal error during connection hijacking: %w", err)}
	}
	conn.Close()
	return nil
}
