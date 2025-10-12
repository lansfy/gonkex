package compare

import (
	"regexp"
)

var knownMatchers = map[string]func(args string) Matcher{
	"$matchBase64": createBase64Matcher,
	"$matchRegexp": createRegexpMatcher,
	"$matchTime":   createTimeMatcher,
}

type Matcher interface {
	MatchValues(actual interface{}) error
}

var matcherExprRx = regexp.MustCompile(`^(\$match[[:alnum:]]+)\((.*)\)$`)

func CreateMatcher(expr interface{}) Matcher {
	name, args := findMatcher(expr)
	if name == "" {
		return nil
	}
	if f, ok := knownMatchers[name]; ok {
		return f(args)
	}
	return &unknownMatcher{name}
}

func findMatcher(expr interface{}) (string, string) {
	sval, ok := expr.(string)
	if !ok {
		return "", ""
	}

	matches := matcherExprRx.FindStringSubmatch(sval)
	if matches == nil {
		return "", ""
	}
	return matches[1], matches[2]
}
