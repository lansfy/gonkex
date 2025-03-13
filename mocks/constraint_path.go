package mocks

import (
	"net/http"

	"github.com/lansfy/gonkex/colorize"
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

	matcher := compare.StringAsMatcher(compare.MatchRegexpWrap(regexpStr))
	if m := compare.StringAsMatcher(pathStr); m != nil {
		pathStr = ""
		matcher = m
	}

	return newPathConstraint(pathStr, matcher), nil
}

func newPathConstraint(path string, matcher compare.Matcher) verifier {
	return &pathConstraint{
		path:    path,
		matcher: matcher,
	}
}

type pathConstraint struct {
	path    string
	matcher compare.Matcher
}

func (c *pathConstraint) GetName() string {
	return "pathMatches"
}

func (c *pathConstraint) Verify(r *http.Request) []error {
	path := r.URL.Path
	if c.path != "" && c.path != path {
		return []error{colorize.NewNotEqualError("url %s does not match expected:", "path", c.path, path)}
	}

	if c.matcher != nil {
		err := c.matcher.MatchValues("url %s:", "path", path)
		if err != nil {
			return []error{err}
		}
	}
	return nil
}
