package mocks

import (
	"errors"
	"net/http"
	"strings"
)

func (l *loaderImpl) loadUriVaryReplyStrategy(path string, def map[string]interface{}) (ReplyStrategy, error) {
	basePath, err := getOptionalStringKey(def, "basePath", true)
	if err != nil {
		return nil, err
	}

	u, ok := def["uris"]
	if !ok {
		return nil, errors.New("'uris' key required")
	}

	urisMap, err := loadStringMap(u, "uris")
	if err != nil {
		return nil, err
	}

	uris := map[string]*Definition{}
	for uriStr, v := range urisMap {
		def, err := l.loadDefinition(path+".uris."+uriStr, v)
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
