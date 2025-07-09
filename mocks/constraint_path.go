package mocks

import (
	"net/http"

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

	if regexpStr != "" {
		pathStr = compare.MatchRegexpWrap(regexpStr)
	}
	return newPathConstraint(pathStr), nil
}

func newPathConstraint(path string) verifier {
	return &pathConstraint{
		path: path,
	}
}

type pathConstraint struct {
	path string
}

func (c *pathConstraint) GetName() string {
	return "pathMatches"
}

func (c *pathConstraint) Verify(r *http.Request) []error {
	return compareValues("url %s", "path", c.path, r.URL.Path)
}
