package compare

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/lansfy/gonkex/colorize"
)

func MatchRegexpWrap(s string) string {
	return fmt.Sprintf("$matchRegexp(%s)", s)
}

func createRegexpMatcher(args string) Matcher {
	return &regexpMatcher{args}
}

type regexpMatcher struct {
	data string
}

func (r *regexpMatcher) MatchValues(actual interface{}) error {
	rx, err := regexp.Compile(r.data)
	if err != nil {
		// simplify error text
		errorText := strings.TrimPrefix(err.Error(), "error parsing regexp: ")
		return makeMatcherParseError("$matchRegexp",
			colorize.NewNotEqualError("cannot compile regexp:", nil, errorText))
	}

	actualType := getLeafType(actual)
	if actualType != leafString && actualType != leafNumber {
		return makeTypeMismatchError([]leafType{leafString, leafNumber}, actualType)
	}

	value := fmt.Sprintf("%v", actual)
	if !rx.MatchString(value) {
		return colorize.NewNotEqualError("value does not match regexp:",
			MatchRegexpWrap(r.data), value)
	}
	return nil
}
