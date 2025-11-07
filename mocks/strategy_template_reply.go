package mocks

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"text/template"
	"time"

	"github.com/lansfy/gonkex/colorize"
)

func (l *loaderImpl) loadTemplateReplyStrategy(def map[string]interface{}) (ReplyStrategy, error) {
	content, statusCode, pause, headers, err := loadCommonParameters(def, false)
	if err != nil {
		return nil, err
	}
	return NewTemplateReply(string(content), statusCode, pause, headers, l.templateReplyFuncs)
}

func NewTemplateReply(content string, statusCode int, pause time.Duration,
	headers map[string]string, funcs template.FuncMap) (ReplyStrategy, error) {
	tmpl, err := template.New("$.body").Funcs(funcs).Parse(content)
	if err != nil {
		return nil, fmt.Errorf("template syntax error: %w", err)
	}

	headersRes := map[string]string{}
	headersTmpl := map[string]*template.Template{}
	for name, value := range headers {
		if strings.Contains(value, "{{") { // if value has template
			tmpl, err := template.New(fmt.Sprintf("$.headers[%s]", name)).Funcs(funcs).Parse(value)
			if err != nil {
				return nil, fmt.Errorf("template syntax error: %w", err)
			}
			headersTmpl[name] = tmpl
		} else {
			headersRes[name] = value
		}
	}

	strategy := &templateReply{
		bodyTmpl:    tmpl,
		statusCode:  statusCode,
		pause:       pause,
		headers:     headersRes,
		headersTmpl: headersTmpl,
	}

	return strategy, nil
}

type templateReply struct {
	bodyTmpl    *template.Template
	statusCode  int
	pause       time.Duration
	headers     map[string]string
	headersTmpl map[string]*template.Template
}

type templateRequest struct {
	r *http.Request

	jsonOnce sync.Once
	jsonData map[string]interface{}
}

func (tr *templateRequest) Header(key string) string {
	return getHeader(tr.r, key)
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
		return nil, fmt.Errorf("parse request as json: %w", err)
	}

	return tr.jsonData, nil
}

func executeTemplate(tmpl *template.Template, r *http.Request, requestBody []byte) (string, *colorize.Error) {
	ctx := map[string]*templateRequest{
		"request": {
			r: r,
		},
	}

	reply := bytes.NewBuffer(nil)
	if err := tmpl.Execute(reply, ctx); err != nil {
		setRequestBody(r, requestBody)
		dump := makeRequestWasParts(r)
		return "", colorize.NewEntityError("strategy %s", "template").WithSubError(err).WithPostfix(dump)
	}

	return reply.String(), nil
}

func (s *templateReply) HandleRequest(w http.ResponseWriter, r *http.Request) []error {
	requestBody, err := getRequestBodyCopy(r)
	if err != nil {
		return []error{err}
	}

	if s.pause > 0 {
		time.Sleep(s.pause)
	}

	responseBody, cErr := executeTemplate(s.bodyTmpl, r, requestBody)
	if cErr != nil {
		return []error{cErr}
	}

	for k, v := range s.headers {
		w.Header().Add(k, v)
	}

	for k, tmpl := range s.headersTmpl {
		v, cErr := executeTemplate(tmpl, r, requestBody)
		if cErr != nil {
			return []error{cErr}
		}
		w.Header().Add(k, v)
	}

	w.WriteHeader(s.statusCode)
	_, _ = w.Write([]byte(responseBody))
	return nil
}
