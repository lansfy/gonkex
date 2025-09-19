package compare

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func Test_TimeMatcher_MatchValues(t *testing.T) {
	oldNowTimeFunc := nowTimeFunc
	nowTimeFunc = func() time.Time {
		return time.Date(2023, 12, 25, 10, 20, 30, 0, time.Local)
	}
	defer func() {
		nowTimeFunc = oldNowTimeFunc
	}()

	tests := []struct {
		description string
		matcher     Matcher
		actual      interface{}
	}{
		{
			description: "matchTime MUST support strftime pattern",
			matcher:     StringAsMatcher("$matchTime(%Y-%m-%d %H:%M:%S)"),
			actual:      "2023-12-25 10:20:30",
		},
		{
			description: "matchTime MUST support golang pattern",
			matcher:     StringAsMatcher("$matchTime(2006-01-02 15:04:05)"),
			actual:      "2023-12-25 10:20:30",
		},
		{
			description: "strftime pattern MUST support reduced number of milliseconds",
			matcher:     StringAsMatcher("$matchTime(%Y-%m-%dT%H:%M:%S.%fZ)"),
			actual:      "2025-05-05T01:01:01.12345Z",
		},
		{
			description: "strftime pattern MUST support absent milliseconds part",
			matcher:     StringAsMatcher("$matchTime(%Y-%m-%dT%H:%M:%S.%fZ)"),
			actual:      "2025-05-05T01:01:01Z",
		},
		{
			description: "strftime pattern MUST support absent milliseconds part",
			matcher:     StringAsMatcher("$matchTime(%Y-%m-%dT%H:%M:%S.%fZ%z)"),
			actual:      "2025-05-19T22:41:14.309131Z",
		},
		{
			description: "matchTime MUST support 'now' function",
			matcher:     StringAsMatcher("$matchTime(%d-%m-%Y %H:%M:%S, value=now)"),
			actual:      "25-12-2023 10:20:30",
		},
		{
			description: "matchTime MUST support 'now()' function",
			matcher:     StringAsMatcher("$matchTime(%d-%m-%Y %H:%M:%S, value=now())"),
			actual:      "25-12-2023 10:20:30",
		},
		{
			description: "time MUST check with accuracy precision (up to 5m after)",
			matcher:     StringAsMatcher("$matchTime(%d-%m-%Y %H:%M:%S, value=now)"),
			actual:      "25-12-2023 10:25:30",
		},
		{
			description: "time MUST check with accuracy precision (up to 5m before)",
			matcher:     StringAsMatcher("$matchTime(%d-%m-%Y %H:%M:%S, value=now)"),
			actual:      "25-12-2023 10:15:30",
		},
		{
			description: "expected time MUST support negative offset",
			matcher:     StringAsMatcher("$matchTime(%d-%m-%Y %H:%M:%S, value=now-1h)"),
			actual:      "25-12-2023 09:25:30",
		},
		{
			description: "expected time MUST support positive offset",
			matcher:     StringAsMatcher("$matchTime(%d-%m-%Y %H:%M:%S, value=now+1h)"),
			actual:      "25-12-2023 11:25:30",
		},
		{
			description: "time MUST support custom accuracy (before expected time)",
			matcher:     StringAsMatcher("$matchTime(%d-%m-%Y %H:%M:%S, value=now(), accuracy=10m)"),
			actual:      "25-12-2023 10:10:30",
		},
		{
			description: "time MUST support custom accuracy (before after time)",
			matcher:     StringAsMatcher("$matchTime(%d-%m-%Y %H:%M:%S, value=now(), accuracy=10m)"),
			actual:      "25-12-2023 10:30:30",
		},
		{
			description: "custom accuracy MUST support explicit direction ('+' for time equal or after value)",
			matcher:     StringAsMatcher("$matchTime(%d-%m-%Y %H:%M:%S, value=now(), accuracy=+10m)"),
			actual:      "25-12-2023 10:30:30",
		},
		{
			description: "custom accuracy MUST support explicit direction ('-' for time equal or before value)",
			matcher:     StringAsMatcher("$matchTime(%d-%m-%Y %H:%M:%S, value=now(), accuracy=-10m)"),
			actual:      "25-12-2023 10:10:30",
		},
		{
			description: "expected time MUST support direct specification",
			matcher:     StringAsMatcher("$matchTime(%d-%m-%Y %H:%M:%S, value=25-12-2023 20:30:00)"),
			actual:      "25-12-2023 20:30:40",
		},
		{
			description: "matchTime MUST support timezone specification with direct value",
			matcher:     StringAsMatcher("$matchTime(%d-%m-%Y %H:%M:%S, value=25-12-2023 20:30:00, timezone=utc)"),
			actual:      "25-12-2023 20:30:00",
		},
		{
			description: "matchTime MUST support timezone specification with now() value",
			matcher:     StringAsMatcher("$matchTime(%Y-%m-%d %H:%M:%S, value=now(), timezone=utc)"),
			actual:      time.Date(2023, 12, 25, 10, 20, 30, 0, time.Local).In(time.UTC).Format("2006-01-02 15:04:05"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.description, func(t *testing.T) {
			require.NotNil(t, tt.matcher)
			err := tt.matcher.MatchValues(tt.actual)
			require.NoError(t, err)
		})
	}
}

func Test_TimeMatcher_MatchValues_Errors(t *testing.T) {
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

	tests := []struct {
		description string
		matcher     Matcher
		actual      interface{}
		wantErr     string
	}{
		{
			description: "invalid actual type",
			matcher:     StringAsMatcher("$matchTime(%Y-%m-%d)"),
			actual:      nil,
			wantErr:     makeMatchError("type mismatch", "string", "<nil>"),
		},
		{
			description: "invalid strftime format specified",
			matcher:     StringAsMatcher("$matchTime(%Y-%m-%!)"),
			actual:      "12-25-2023",
			wantErr:     "pattern '%Y-%m-%!': strftime: unsupported directive: %! ",
		},
		{
			description: "time doesn't match to specified strftime format",
			matcher:     StringAsMatcher("$matchTime(%Y-%m-%d)"),
			actual:      "12-25-2023",
			wantErr:     makeMatchError("time does not match the template", "$matchTime(%Y-%m-%d)", "12-25-2023"),
		},
		{
			description: "time doesn't match to specified golang format",
			matcher:     StringAsMatcher("$matchTime(2006-01-02)"),
			actual:      "12-25-2023",
			wantErr:     makeMatchError("time does not match the template", "$matchTime(2006-01-02)", "12-25-2023"),
		},
		{
			description: "invalid duration format in accuracy parameter",
			matcher:     StringAsMatcher("$matchTime(%Y-%m-%d, accuracy=some-wrong-value)"),
			actual:      "12-25-2023",
			wantErr:     "parameter 'accuracy': wrong duration value 'some-wrong-value'",
		},
		{
			description: "invalid duration format in value parameter",
			matcher:     StringAsMatcher("$matchTime(%Y-%m-%d, value=now-1dddd)"),
			actual:      "12-25-2023",
			wantErr:     "parameter 'value': wrong duration value '-1dddd'",
		},
		{
			description: "invalid timezone value parameter",
			matcher:     StringAsMatcher("$matchTime(%Y-%m-%d, value=now, timezone=wrong)"),
			actual:      "2023-12-25",
			wantErr:     makeMatchError("wrong 'timezone' value", "local / utc", "wrong"),
		},
		{
			description: "invalid parameter name",
			matcher:     StringAsMatcher("$matchTime(%Y-%m-%d,fakeparam=aaaa)"),
			actual:      "12-25-2023",
			wantErr:     "parameter 'fakeparam=aaaa': unknown parameter name",
		},
		{
			description: "WHEN actual time before (expected-accuracy) MUST fail with error",
			matcher:     StringAsMatcher("$matchTime(%d-%m-%Y %H:%M:%S, value=now())"),
			actual:      "25-12-2023 10:15:00",
			wantErr:     makeMatchError("values do not match", "25-12-2023 10:15:30 ... 25-12-2023 10:25:30", "25-12-2023 10:15:00"),
		},
		{
			description: "WHEN actual time after (expected+accuracy) MUST fail with error",
			matcher:     StringAsMatcher("$matchTime(%d-%m-%Y %H:%M:%S, value=now())"),
			actual:      "25-12-2023 10:26:00",
			wantErr:     makeMatchError("values do not match", "25-12-2023 10:15:30 ... 25-12-2023 10:25:30", "25-12-2023 10:26:00"),
		},
		{
			description: "WHEN custom accuracy has explicit + actual time before expected MUST fail with error",
			matcher:     StringAsMatcher("$matchTime(%d-%m-%Y %H:%M:%S, value=now(), accuracy=+10m)"),
			actual:      "25-12-2023 10:20:00",
			wantErr:     makeMatchError("values do not match", "25-12-2023 10:20:30 ... 25-12-2023 10:30:30", "25-12-2023 10:20:00"),
		},
		{
			description: "WHEN custom accuracy has explicit - actual time after expected MUST fail with error",
			matcher:     StringAsMatcher("$matchTime(%d-%m-%Y %H:%M:%S, value=now(), accuracy=-10m)"),
			actual:      "25-12-2023 10:21:00",
			wantErr:     makeMatchError("values do not match", "25-12-2023 10:10:30 ... 25-12-2023 10:20:30", "25-12-2023 10:21:00"),
		},
		{
			description: "WHEN value parameter doesn't match pattern check MUST fail",
			matcher:     StringAsMatcher("$matchTime(%d-%m-%Y %H:%M:%S, value=2023-12-25 20:30:00)"),
			actual:      "25-12-2023 20:30:40",
			wantErr:     "parameter 'value': time value '2023-12-25 20:30:00' doesn't match pattern '%d-%m-%Y %H:%M:%S'",
		},
	}

	for _, tt := range tests {
		t.Run(tt.description, func(t *testing.T) {
			require.NotNil(t, tt.matcher)
			err := tt.matcher.MatchValues(tt.actual)
			require.Error(t, err)
			require.Equal(t, tt.wantErr, err.Error())
		})
	}
}
