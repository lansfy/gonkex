package mocks

import (
	"net/http"
)

type ReplyStrategy interface {
	HandleRequest(w http.ResponseWriter, r *http.Request) []error
}

type contextAwareStrategy interface {
	ResetRunningContext()
	EndRunningContext(intermediate bool) []error
}
