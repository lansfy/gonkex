package compare

import (
	"regexp"
)

type Matcher interface {
	MatchValues(description, entity string, actual interface{}) error
}

var matcherExprRx = regexp.MustCompile(`^\$match(Regexp|Time|Base64)\((.+)\)$`)

func StringAsMatcher(expr string) Matcher {
	matches := matcherExprRx.FindStringSubmatch(expr)
	if matches != nil {
		switch matches[1] {
		case "Regexp":
			return &regexpMatcher{matches[2]}
		case "Base64":
			return &base64Matcher{matches[2]}
		case "Time":
			return &timeMatcher{matches[2]}
		}
	}

	return nil
}
