package mocks

import (
	"fmt"
	"net/http"
	"regexp"

	"github.com/lansfy/gonkex/compare"
)

func loadPathConstraint(def map[interface{}]interface{}) (verifier, error) {
	pathStr, err := getOptionalStringKey(def, "path", true)
	if err != nil {
		return nil, err
	}
	regexpStr, err := getOptionalStringKey(def, "regexp", false)
	if err != nil {
		return nil, err
	}

	if s, ok := compare.StringAsRegexp(pathStr); ok {
		pathStr = ""
		regexpStr = s
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

func (c *pathConstraint) GetName() string {
	return "pathMatches"
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
