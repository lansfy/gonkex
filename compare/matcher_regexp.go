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

type regexpMatcher struct {
	data string
}

func (r *regexpMatcher) MatchValues(description, entity string, actual interface{}) error {
	rx, err := regexp.Compile(r.data)
	if err != nil {
		// simplify error text
		errorText := strings.TrimPrefix(err.Error(), "error parsing regexp: ")
		return colorize.NewNotEqualError(description+" cannot compile regexp:", entity, nil, errorText)
	}

	value := fmt.Sprintf("%v", actual)
	if !rx.MatchString(value) {
		return colorize.NewNotEqualError(description+" value does not match regexp:",
			entity, MatchRegexpWrap(r.data), value)
	}
	return nil
}
