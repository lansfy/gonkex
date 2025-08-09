package mocks

import (
	"net/http"
	"os"
	"time"
)

//nolint:nakedret // for convenience
func loadCommonParameters(def map[string]interface{}, isFile bool) (
	content []byte, statusCode int, pause time.Duration, headers map[string]string, err error) {
	statusCode, err = getOptionalIntKey(def, "statusCode", http.StatusOK)
	if err != nil {
		return
	}
	pause, err = getOptionalDurationKey(def, "pause")
	if err != nil {
		return
	}

	headers, err = loadHeaders(def)
	if err != nil {
		return
	}

	if isFile {
		var filename string
		filename, err = getRequiredStringKey(def, "filename", false)
		if err != nil {
			return
		}

		content, err = os.ReadFile(filename)
		return
	}

	var body string
	body, err = getRequiredStringKey(def, "body", true)
	if err != nil {
		return
	}
	content = []byte(body)
	return
}

func (l *loaderImpl) loadFileStrategy(def map[string]interface{}) (ReplyStrategy, error) {
	content, statusCode, pause, headers, err := loadCommonParameters(def, true)
	if err != nil {
		return nil, err
	}
	return NewConstantReplyWithCode(content, statusCode, pause, headers), nil
}

func (l *loaderImpl) loadConstantStrategy(def map[string]interface{}) (ReplyStrategy, error) {
	content, statusCode, pause, headers, err := loadCommonParameters(def, false)
	if err != nil {
		return nil, err
	}
	return NewConstantReplyWithCode(content, statusCode, pause, headers), nil
}

func NewConstantReplyWithCode(content []byte, statusCode int, pause time.Duration,
	headers map[string]string) ReplyStrategy {
	return &constantReply{
		replyBody:  content,
		statusCode: statusCode,
		pause:      pause,
		headers:    headers,
	}
}

type constantReply struct {
	replyBody  []byte
	statusCode int
	pause      time.Duration
	headers    map[string]string
}

func (s *constantReply) HandleRequest(w http.ResponseWriter, r *http.Request) []error {
	if s.pause > 0 {
		time.Sleep(s.pause)
	}
	for k, v := range s.headers {
		w.Header().Add(k, v)
	}
	w.WriteHeader(s.statusCode)
	_, _ = w.Write(s.replyBody)
	return nil
}
