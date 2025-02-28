package mocks

import (
	"net/http"
	"os"
)

func (l *loaderImpl) loadFileStrategy(def map[interface{}]interface{}) (ReplyStrategy, error) {
	filename, err := getRequiredStringKey(def, "filename", false)
	if err != nil {
		return nil, err
	}
	statusCode, err := getOptionalIntKey(def, "statusCode", http.StatusOK)
	if err != nil {
		return nil, err
	}
	headers, err := loadHeaders(def)
	if err != nil {
		return nil, err
	}
	content, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	return NewConstantReplyWithCode(content, statusCode, headers), nil
}

func (l *loaderImpl) loadConstantStrategy(def map[interface{}]interface{}) (ReplyStrategy, error) {
	body, err := getRequiredStringKey(def, "body", true)
	if err != nil {
		return nil, err
	}
	statusCode, err := getOptionalIntKey(def, "statusCode", http.StatusOK)
	if err != nil {
		return nil, err
	}
	headers, err := loadHeaders(def)
	if err != nil {
		return nil, err
	}
	return NewConstantReplyWithCode([]byte(body), statusCode, headers), nil
}

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
	_, _ = w.Write(s.replyBody)
	return nil
}
