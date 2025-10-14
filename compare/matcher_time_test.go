package compare

import (
	"fmt"
	"testing"
	"time"
)

func Test_timeMatcher_MatchValues(t *testing.T) {
	oldNowTimeFunc := nowTimeFunc
	nowTimeFunc = func() time.Time {
		return time.Date(2023, 12, 25, 10, 20, 30, 0, time.Local)
	}
	defer func() {
		nowTimeFunc = oldNowTimeFunc
	}()

	tests := []matcherTest{
		{
			description: "matchTime MUST support strftime pattern",
			matcher:     "$matchTime(%Y-%m-%d %H:%M:%S)",
			actual:      "2023-12-25 10:20:30",
		},
		{
			description: "matchTime MUST support golang pattern",
			matcher:     "$matchTime(2006-01-02 15:04:05)",
			actual:      "2023-12-25 10:20:30",
		},
		{
			description: "strftime pattern MUST support reduced number of milliseconds",
			matcher:     "$matchTime(%Y-%m-%dT%H:%M:%S.%fZ)",
			actual:      "2025-05-05T01:01:01.12345Z",
		},
		{
			description: "strftime pattern MUST support absent milliseconds part",
			matcher:     "$matchTime(%Y-%m-%dT%H:%M:%S.%fZ)",
			actual:      "2025-05-05T01:01:01Z",
		},
		{
			description: "strftime pattern MUST support absent milliseconds part",
			matcher:     "$matchTime(%Y-%m-%dT%H:%M:%S.%fZ%z)",
			actual:      "2025-05-19T22:41:14.309131Z",
		},
		{
			description: "matchTime MUST support 'now' function",
			matcher:     "$matchTime(%d-%m-%Y %H:%M:%S, value=now)",
			actual:      "25-12-2023 10:20:30",
		},
		{
			description: "matchTime MUST support 'now()' function",
			matcher:     "$matchTime(%d-%m-%Y %H:%M:%S, value=now())",
			actual:      "25-12-2023 10:20:30",
		},
		{
			description: "time MUST check with accuracy precision (up to 5m after)",
			matcher:     "$matchTime(%d-%m-%Y %H:%M:%S, value=now)",
			actual:      "25-12-2023 10:25:30",
		},
		{
			description: "time MUST check with accuracy precision (up to 5m before)",
			matcher:     "$matchTime(%d-%m-%Y %H:%M:%S, value=now)",
			actual:      "25-12-2023 10:15:30",
		},
		{
			description: "expected time MUST support negative offset",
			matcher:     "$matchTime(%d-%m-%Y %H:%M:%S, value=now-1h)",
			actual:      "25-12-2023 09:25:30",
		},
		{
			description: "expected time MUST support positive offset",
			matcher:     "$matchTime(%d-%m-%Y %H:%M:%S, value=now+1h)",
			actual:      "25-12-2023 11:25:30",
		},
		{
			description: "time MUST support custom accuracy (before expected time)",
			matcher:     "$matchTime(%d-%m-%Y %H:%M:%S, value=now(), accuracy=10m)",
			actual:      "25-12-2023 10:10:30",
		},
		{
			description: "time MUST support custom accuracy (before after time)",
			matcher:     "$matchTime(%d-%m-%Y %H:%M:%S, value=now(), accuracy=10m)",
			actual:      "25-12-2023 10:30:30",
		},
		{
			description: "custom accuracy MUST support explicit direction ('+' for time equal or after value)",
			matcher:     "$matchTime(%d-%m-%Y %H:%M:%S, value=now(), accuracy=+10m)",
			actual:      "25-12-2023 10:30:30",
		},
		{
			description: "custom accuracy MUST support explicit direction ('-' for time equal or before value)",
			matcher:     "$matchTime(%d-%m-%Y %H:%M:%S, value=now(), accuracy=-10m)",
			actual:      "25-12-2023 10:10:30",
		},
		{
			description: "expected time MUST support direct specification",
			matcher:     "$matchTime(%d-%m-%Y %H:%M:%S, value=25-12-2023 20:30:00)",
			actual:      "25-12-2023 20:30:40",
		},
		{
			description: "matchTime MUST support timezone specification with direct value",
			matcher:     "$matchTime(%d-%m-%Y %H:%M:%S, value=25-12-2023 20:30:00, timezone=utc)",
			actual:      "25-12-2023 20:30:00",
		},
		{
			description: "matchTime MUST support timezone specification with now() value",
			matcher:     "$matchTime(%Y-%m-%d %H:%M:%S, value=now(), timezone=utc)",
			actual:      time.Date(2023, 12, 25, 10, 20, 30, 0, time.Local).In(time.UTC).Format("2006-01-02 15:04:05"),
		},
	}

	processTests(t, tests, Params{})
}

func Test_timeMatcher_MatchValues_Errors(t *testing.T) {
	oldNowTimeFunc := nowTimeFunc
	nowTimeFunc = func() time.Time {
		return time.Date(2023, 12, 25, 10, 20, 30, 0, time.Local)
	}
	defer func() {
		nowTimeFunc = oldNowTimeFunc
	}()

	makeMatchError := func(text, expected, actual string) string {
		return fmt.Sprintf("%s:\n     expected: %s\n       actual: %s",
			text, expected, actual)
	}

	tests := []matcherTest{
		{
			description: "invalid actual type",
			matcher:     "$matchTime(%Y-%m-%d)",
			actual:      nil,
			wantErr:     makeMatchError("type mismatch", "string", "nil"),
		},
		{
			description: "invalid strftime format specified",
			matcher:     "$matchTime(%Y-%m-%!)",
			actual:      "12-25-2023",
			wantErr:     "parse '$matchTime': pattern '%Y-%m-%!': strftime: unsupported directive: %! ",
		},
		{
			description: "time doesn't match to specified strftime format",
			matcher:     "$matchTime(%Y-%m-%d)",
			actual:      "12-25-2023",
			wantErr:     makeMatchError("time does not match the template", "$matchTime(%Y-%m-%d)", "12-25-2023"),
		},
		{
			description: "time doesn't match to specified golang format",
			matcher:     "$matchTime(2006-01-02)",
			actual:      "12-25-2023",
			wantErr:     makeMatchError("time does not match the template", "$matchTime(2006-01-02)", "12-25-2023"),
		},
		{
			description: "invalid duration format in accuracy parameter",
			matcher:     "$matchTime(%Y-%m-%d, accuracy=some-wrong-value)",
			actual:      "12-25-2023",
			wantErr:     "parse '$matchTime': parameter 'accuracy': wrong duration value 'some-wrong-value'",
		},
		{
			description: "invalid duration format in value parameter",
			matcher:     "$matchTime(%Y-%m-%d, value=now-1dddd)",
			actual:      "12-25-2023",
			wantErr:     "parse '$matchTime': parameter 'value': wrong duration value '-1dddd'",
		},
		{
			description: "invalid timezone value parameter",
			matcher:     "$matchTime(%Y-%m-%d, value=now, timezone=wrong)",
			actual:      "2023-12-25",
			wantErr:     makeMatchError("parse '$matchTime': wrong 'timezone' value", "local / utc", "wrong"),
		},
		{
			description: "invalid parameter name",
			matcher:     "$matchTime(%Y-%m-%d,fakeparam=aaaa)",
			actual:      "12-25-2023",
			wantErr:     "parse '$matchTime': parameter 'fakeparam=aaaa': unknown parameter name",
		},
		{
			description: "WHEN actual time before (expected-accuracy) MUST fail with error",
			matcher:     "$matchTime(%d-%m-%Y %H:%M:%S, value=now())",
			actual:      "25-12-2023 10:15:00",
			wantErr:     makeMatchError("values do not match", "25-12-2023 10:15:30 ... 25-12-2023 10:25:30", "25-12-2023 10:15:00"),
		},
		{
			description: "WHEN actual time after (expected+accuracy) MUST fail with error",
			matcher:     "$matchTime(%d-%m-%Y %H:%M:%S, value=now())",
			actual:      "25-12-2023 10:26:00",
			wantErr:     makeMatchError("values do not match", "25-12-2023 10:15:30 ... 25-12-2023 10:25:30", "25-12-2023 10:26:00"),
		},
		{
			description: "WHEN custom accuracy has explicit + actual time before expected MUST fail with error",
			matcher:     "$matchTime(%d-%m-%Y %H:%M:%S, value=now(), accuracy=+10m)",
			actual:      "25-12-2023 10:20:00",
			wantErr:     makeMatchError("values do not match", "25-12-2023 10:20:30 ... 25-12-2023 10:30:30", "25-12-2023 10:20:00"),
		},
		{
			description: "WHEN custom accuracy has explicit - actual time after expected MUST fail with error",
			matcher:     "$matchTime(%d-%m-%Y %H:%M:%S, value=now(), accuracy=-10m)",
			actual:      "25-12-2023 10:21:00",
			wantErr:     makeMatchError("values do not match", "25-12-2023 10:10:30 ... 25-12-2023 10:20:30", "25-12-2023 10:21:00"),
		},
		{
			description: "WHEN value parameter doesn't match pattern check MUST fail",
			matcher:     "$matchTime(%d-%m-%Y %H:%M:%S, value=2023-12-25 20:30:00)",
			actual:      "25-12-2023 20:30:40",
			wantErr:     "parse '$matchTime': parameter 'value': time value '2023-12-25 20:30:00' doesn't match pattern '%d-%m-%Y %H:%M:%S'",
		},
	}

	processTests(t, tests, Params{})
}

func Test_timeMatcher_UnsupportedTypes(t *testing.T) {
	tests := []matcherTest{
		{
			description: "match regexp to array MUST fail with type error",
			matcher:     "$matchTime(%d-%m-%Y %H:%M:%S)",
			actual:      []string{},
			wantErr:     "type mismatch:\n     expected: string\n       actual: array",
		},
		{
			description: "match regexp to map MUST fail with type error",
			matcher:     "$matchTime(%d-%m-%Y %H:%M:%S)",
			actual:      map[string]string{},
			wantErr:     "type mismatch:\n     expected: string\n       actual: map",
		},
		{
			description: "match regexp to bool MUST fail with type error",
			matcher:     "$matchTime(%d-%m-%Y %H:%M:%S)",
			actual:      true,
			wantErr:     "type mismatch:\n     expected: string\n       actual: bool",
		},
		{
			description: "match regexp to nil MUST fail with type error",
			matcher:     "$matchTime(%d-%m-%Y %H:%M:%S)",
			actual:      nil,
			wantErr:     "type mismatch:\n     expected: string\n       actual: nil",
		},
		{
			description: "match regexp to invalid type MUST fail with type error",
			matcher:     "$matchTime(%d-%m-%Y %H:%M:%S)",
			actual:      t,
			wantErr:     "type mismatch:\n     expected: string\n       actual: *testing.T",
		},
	}

	processTests(t, tests, Params{})
}

func Test_timeMatcher_IgnoreValues(t *testing.T) {
	tests := []matcherTest{
		{
			description: "WHEN IgnoreValues specified $matchRegexp MUST be ignored with scalar type",
			matcher:     "$matchTime(%d-%m-%Y %H:%M:%S)",
			actual:      "other",
		},
		{
			description: "WHEN IgnoreValues specified and $matchRegexp compares with non-scalar type test MUST fail",
			matcher:     "$matchTime(%d-%m-%Y %H:%M:%S)",
			actual:      []string{"other"},
			wantErr:     "type mismatch:\n     expected: string\n       actual: array",
		},
	}

	processTests(t, tests, Params{
		IgnoreValues: true,
	})
}
