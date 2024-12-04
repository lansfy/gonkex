package mocks

import (
	"fmt"
	"net/http"
)

func NewDropRequestReply() ReplyStrategy {
	return &dropRequestReply{}
}

type dropRequestReply struct{}

func (s *dropRequestReply) HandleRequest(w http.ResponseWriter, r *http.Request) []error {
	hj, ok := w.(http.Hijacker)
	if !ok {
		return []error{fmt.Errorf("Gonkex internal error during drop request: webserver doesn't support hijacking\n")}
	}
	conn, _, err := hj.Hijack()
	if err != nil {
		return []error{fmt.Errorf("Gonkex internal error during connection hijacking: %s\n", err)}
	}
	conn.Close()
	return nil
}
