package mocks

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"text/template"
)

func (l *loaderImpl) loadTemplateReplyStrategy(def map[interface{}]interface{}) (ReplyStrategy, error) {
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
	return NewTemplateReply(body, statusCode, headers, l.templateReplyFuncs)
}

func NewTemplateReply(content string, statusCode int, headers map[string]string,
	funcs template.FuncMap) (ReplyStrategy, error) {
	tmpl, err := template.New("").Funcs(funcs).Parse(content)
	if err != nil {
		return nil, fmt.Errorf("template syntax error: %w", err)
	}

	strategy := &templateReply{
		replyBodyTemplate: tmpl,
		statusCode:        statusCode,
		headers:           headers,
	}

	return strategy, nil
}

type templateReply struct {
	replyBodyTemplate *template.Template
	statusCode        int
	headers           map[string]string
}

type templateRequest struct {
	r *http.Request

	jsonOnce sync.Once
	jsonData map[string]interface{}
}

func (tr *templateRequest) Query(key string) string {
	return tr.r.URL.Query().Get(key)
}

func (tr *templateRequest) Json() (map[string]interface{}, error) {
	var err error
	tr.jsonOnce.Do(func() {
		err = json.NewDecoder(tr.r.Body).Decode(&tr.jsonData)
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse request as Json: %w", err)
	}

	return tr.jsonData, nil
}

func (s *templateReply) executeResponseTemplate(r *http.Request) (string, error) {
	ctx := map[string]*templateRequest{
		"request": {
			r: r,
		},
	}

	reply := bytes.NewBuffer(nil)
	if err := s.replyBodyTemplate.Execute(reply, ctx); err != nil {
		return "", fmt.Errorf("template mock error: %w", err)
	}

	return reply.String(), nil
}

func (s *templateReply) HandleRequest(w http.ResponseWriter, r *http.Request) []error {
	for k, v := range s.headers {
		w.Header().Add(k, v)
	}

	responseBody, err := s.executeResponseTemplate(r)
	if err != nil {
		return []error{err}
	}

	w.WriteHeader(s.statusCode)
	_, _ = w.Write([]byte(responseBody))
	return nil
}
