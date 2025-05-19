package compare

import (
	"fmt"
	"reflect"
	"regexp"
	"strings"
	"time"

	"github.com/lansfy/gonkex/colorize"
	"github.com/ncruces/go-strftime"
	"github.com/xhit/go-str2duration/v2"
)

type timeMatcher struct {
	data string
}

func (m *timeMatcher) MatchValues(description, entity string, actual interface{}) error {
	actualStr, ok := actual.(string)
	if !ok {
		return colorize.NewNotEqualError(description+" type mismatch:", entity, "string", reflect.TypeOf(actual))
	}

	args, err := extractTimeArgs(m.data)
	if err != nil {
		return colorize.NewEntityError(description, entity).SetSubError(err)
	}

	parsed, err := time.ParseInLocation(args.layout, actualStr, time.Local)
	if err != nil {
		return colorize.NewNotEqualError(description+" time does not match the template:",
			entity, fmt.Sprintf("$matchTime(%s)", m.data), actualStr)
	}

	if args.fromTime.Equal(time.Time{}) {
		return nil
	}

	fromTime := args.fromTime.In(parsed.Location())
	toTime := args.toTime.In(parsed.Location())
	if parsed.Before(fromTime) || parsed.After(toTime) {
		expected := fmt.Sprintf("%s ... %s", fromTime.Format(args.layout), toTime.Format(args.layout))
		return colorize.NewNotEqualError(description+" values do not match:", entity, expected, actualStr)
	}

	return nil
}

var timeDefaultParams = map[string]string{
	"value":    "",
	"accuracy": "5m",
}

var valueFormatExpr = regexp.MustCompile(`^(.+?)([+-](\d+[wdhmnus]+)+)?$`)
var nowTimeFunc = time.Now

func parseValue(args *timeParamsData, value string) (time.Time, error) {
	if value == "" {
		return time.Time{}, nil
	}
	// any expression matched by this regexp
	matches := valueFormatExpr.FindStringSubmatch(value)

	baseStr := matches[1]
	shiftStr := matches[2]

	var shift time.Duration
	var err error
	if shiftStr != "" {
		shift, err = str2duration.ParseDuration(shiftStr)
		if err != nil {
			return time.Time{}, fmt.Errorf("wrong duration value '%s'", shiftStr)
		}
	}

	if baseStr == "now" || baseStr == "now()" {
		return nowTimeFunc().Add(shift), nil
	}

	base, err := time.ParseInLocation(args.layout, baseStr, time.Local)
	if err != nil {
		return time.Time{}, fmt.Errorf("time value '%s' doesn't match pattern '%s'", baseStr, args.origLayout)
	}

	return base.Add(shift), nil
}

type timeParamsData struct {
	origLayout       string
	layout           string
	fromTime, toTime time.Time
}

var millisecondsFixExpr = regexp.MustCompile(`\.0{3,9}`)

func patternNormalization(pattern string) string {
	pattern = millisecondsFixExpr.ReplaceAllStringFunc(pattern, func(s string) string {
		return strings.ReplaceAll(s, "0", "9")
	})

	pattern = strings.ReplaceAll(pattern, "Z-0700", "Z0700")
	return pattern
}

func extractTimeArgs(data string) (*timeParamsData, error) {
	result := &timeParamsData{}

	value, params, err := extractArgs(data, timeDefaultParams)
	if err != nil {
		return nil, err
	}

	result.origLayout = value
	if strings.ContainsAny(value, "0123456789") {
		// golang time pattern
		result.layout = value
	} else {
		// strftime time pattern
		result.layout, err = strftime.Layout(value)
		if err != nil {
			return nil, colorize.NewEntityError("pattern %s", value).SetSubError(err)
		}
		result.layout = patternNormalization(result.layout)
	}

	accuracyStr := params["accuracy"]
	accuracy, err := str2duration.ParseDuration(accuracyStr)
	if err != nil {
		return nil,
			colorize.NewEntityError("parameter %s", "accuracy").SetSubError(
				fmt.Errorf("wrong duration value '%s'", accuracyStr))
	}

	if accuracy < 0 {
		accuracy = -1 * accuracy
	}

	initial, err := parseValue(result, params["value"])
	if err != nil {
		return nil, colorize.NewEntityError("parameter %s", "value").SetSubError(err)
	}

	if !initial.Equal(time.Time{}) {
		result.fromTime = initial
		result.toTime = initial
		if accuracyStr[0] != '+' {
			result.fromTime = initial.Add(-1 * accuracy)
		}
		if accuracyStr[0] != '-' {
			result.toTime = initial.Add(accuracy)
		}
	}

	return result, nil
}
