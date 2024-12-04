package mocks

import (
	"errors"
	"fmt"
	"net/http"
	"regexp"
)

func loadPathConstraint(def map[interface{}]interface{}) (verifier, error) {
	var pathStr, regexpStr string
	if path, ok := def["path"]; ok {
		pathStr, ok = path.(string)
		if !ok {
			return nil, errors.New("`path` must be string")
		}
	}
	if regexp, ok := def["regexp"]; ok {
		regexpStr, ok = regexp.(string)
		if !ok || regexp == "" {
			return nil, errors.New("`regexp` must be string")
		}
	}
	return newPathConstraint(pathStr, regexpStr)
}

func newPathConstraint(path, re string) (verifier, error) {
	var reCompiled *regexp.Regexp
	if re != "" {
		var err error
		reCompiled, err = regexp.Compile(re)
		if err != nil {
			return nil, err
		}
	}
	res := &pathConstraint{
		path:   path,
		regexp: reCompiled,
	}
	return res, nil
}

type pathConstraint struct {
	path   string
	regexp *regexp.Regexp
}

func (c *pathConstraint) Verify(r *http.Request) []error {
	path := r.URL.Path
	if c.path != "" && c.path != path {
		return []error{fmt.Errorf("url path %s doesn't match expected %s", path, c.path)}
	}
	if c.regexp != nil && !c.regexp.MatchString(path) {
		return []error{fmt.Errorf("url path %s doesn't match regexp %s", path, c.regexp)}
	}
	return nil
}

func (c *pathConstraint) Fields() []string {
	return []string{"path", "regexp"}
}
