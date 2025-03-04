package compare

import (
	"regexp"
)

type Matcher interface {
	MatchValues(description, entity string, actual interface{}) error
}

var matcherExprRx = regexp.MustCompile(`^\$match(Regexp)\((.+)\)$`)

func StringAsMatcher(expr string) Matcher {
	matches := matcherExprRx.FindStringSubmatch(expr)
	if matches != nil {
		switch matches[1] {
		case "Regexp":
			return &regexpMatcher{matches[2]}
		}
	}

	return nil
}
