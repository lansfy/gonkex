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

var regexpMatcherSupported = leafTypeSet{
	leafString: true,
	leafNumber: true,
}

type regexpMatcher struct {
	data string
}

func (r *regexpMatcher) MatchValues(actual interface{}) error {
	rx, err := regexp.Compile(r.data)
	if err != nil {
		// simplify error text
		errorText := strings.TrimPrefix(err.Error(), "error parsing regexp: ")
		return colorize.NewNotEqualError("cannot compile regexp:", nil, errorText)
	}

	err = checkTypeCompatibility(regexpMatcherSupported, actual)
	if err != nil {
		return err
	}

	value := fmt.Sprintf("%v", actual)
	if !rx.MatchString(value) {
		return colorize.NewNotEqualError("value does not match regexp:",
			MatchRegexpWrap(r.data), value)
	}
	return nil
}
