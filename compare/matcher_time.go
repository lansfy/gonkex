package compare

import (
	"fmt"
	"regexp"
	"time"

	"github.com/lansfy/gonkex/colorize"
)

type timeMatcher struct {
	data string
}

func (m *timeMatcher) MatchValues(description, entity string, actual interface{}) error {
	layout := convertPythonToGoFormat(m.data)
	value := fmt.Sprintf("%v", actual)
	_, err := time.Parse(layout, value)
	if err != nil {
		return colorize.NewNotEqualError(description+" time does not match the template:",
			entity, fmt.Sprintf("$matchTime(%s)", m.data), value)
	}
	return nil
}

var (
	timeFormatExprRx  = regexp.MustCompile("%[a-zA-Z]")
	pythonToGoFormats = map[string]string{
		"%Y": "2006",
		"%y": "06",
		"%m": "01",
		"%d": "02",
		"%H": "15",
		"%I": "03",
		"%M": "04",
		"%S": "05",
		"%f": "999999",
		"%p": "PM",
		"%z": "-0700",
		"%Z": "MST",
		"%j": "002",
		"%U": "__WEEK_NUMBER__",
		"%W": "__ISO_WEEK__",
		"%a": "Mon",
		"%A": "Monday",
		"%b": "Jan",
		"%B": "January",
		"%c": "Mon Jan 2 15:04:05 2006",
		"%x": "01/02/06",
		"%X": "15:04:05",
	}
)

func convertPythonToGoFormat(pyFormat string) string {
	return timeFormatExprRx.ReplaceAllStringFunc(pyFormat, func(match string) string {
		goFmt, ok := pythonToGoFormats[match]
		if ok {
			return goFmt
		}
		return match
	})
}
