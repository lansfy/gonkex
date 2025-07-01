package mocks

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/lansfy/gonkex/colorize"
)

func (l *loaderImpl) loadUriVaryReplyStrategy(path string, def map[interface{}]interface{}) (ReplyStrategy, error) {
	basePath, err := getOptionalStringKey(def, "basePath", true)
	if err != nil {
		return nil, err
	}

	u, ok := def["uris"]
	if !ok {
		return nil, errors.New("'uris' key required")
	}

	urisMap, ok := u.(map[interface{}]interface{})
	if !ok {
		return nil, colorize.NewEntityError("map under %s key required", "uris")
	}

	uris := map[string]*Definition{}
	for uri, v := range urisMap {
		uriStr, ok := uri.(string)
		if !ok {
			return nil, fmt.Errorf("uri '%v' has non-string name", uri)
		}
		def, err := l.loadDefinition(path+"."+uriStr, v)
		if err != nil {
			return nil, err
		}
		uris[uriStr] = def
	}
	return NewUriVaryReplyStrategy(basePath, uris), nil
}

func NewUriVaryReplyStrategy(basePath string, variants map[string]*Definition) ReplyStrategy {
	return &uriVaryReply{
		basePath: strings.TrimRight(basePath, "/") + "/",
		variants: variants,
	}
}

var _ contextAwareStrategy = (*uriVaryReply)(nil)

type uriVaryReply struct {
	basePath string
	variants map[string]*Definition
}

func (s *uriVaryReply) HandleRequest(w http.ResponseWriter, r *http.Request) []error {
	for uri, def := range s.variants {
		uri = strings.TrimLeft(uri, "/")
		if s.basePath+uri == r.URL.Path {
			return def.Execute(w, r)
		}
	}
	return unhandledRequestError(r)
}

func (s *uriVaryReply) ResetRunningContext() {
	for _, def := range s.variants {
		def.ResetRunningContext()
	}
}

func (s *uriVaryReply) EndRunningContext() []error {
	var errs []error
	for _, def := range s.variants {
		errs = append(errs, def.EndRunningContext()...)
	}
	return errs
}
