package compare

import (
	"errors"
)

type unknownMatcher struct {
	name string
}

func (r *unknownMatcher) MatchValues(actual interface{}) error {
	if r.name == "$matchArray" {
		return makeMatcherParseError(r.name, errors.New("must be first element in array"))
	}
	return makeMatcherParseError(r.name, errors.New("unknown matcher name"))
}
