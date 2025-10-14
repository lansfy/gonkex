package compare

import (
	"testing"
)

func Test_regexpMatcher_MatchValues(t *testing.T) {
	tests := []matcherTest{
		{
			description: "WHEN the string matches a regular expression, the check MUST pass",
			matcher:     "$matchRegexp(^t.+t$)",
			actual:      "test",
		},
		{
			description: "WHEN the string not matches a regular expression, the check MUST fail",
			matcher:     "$matchRegexp(^t.+t$)",
			actual:      "not-equal",
			wantErr:     "value does not match regexp:\n     expected: $matchRegexp(^t.+t$)\n       actual: not-equal",
		},

		{
			description: "WHEN the integer number matches a regular expression, the check MUST pass",
			matcher:     "$matchRegexp(^[0-5]+$)",
			actual:      543210,
		},
		{
			description: "WHEN the integer number not matches a regular expression, the check MUST fail",
			matcher:     "$matchRegexp(^[0-5]+$)",
			actual:      12367,
			wantErr:     "value does not match regexp:\n     expected: $matchRegexp(^[0-5]+$)\n       actual: 12367",
		},
		{
			description: "WHEN the float number matches a regular expression, the check MUST pass",
			matcher:     "$matchRegexp(^[0-9]+\\.2.*$)",
			actual:      1.234,
		},
		{
			description: "WHEN the float number not matches a regular expression, the check MUST fail",
			matcher:     "$matchRegexp(^[0-9]+\\.3.*$)",
			actual:      1.23,
			wantErr:     "value does not match regexp:\n     expected: $matchRegexp(^[0-9]+\\.3.*$)\n       actual: 1.23",
		},
		{
			description: "WHEN condition has invalid regular expression, the check MUST fail with error",
			matcher:     "$matchRegexp((unclosed)",
			actual:      "test",
			wantErr:     "parse '$matchRegexp': cannot compile regexp:\n     expected: <nil>\n       actual: missing closing ): `(unclosed`",
		},
	}

	processTests(t, tests, Params{})
}

func Test_regexpMatcher_UnsupportedTypes(t *testing.T) {
	tests := []matcherTest{
		{
			description: "match regexp to array MUST fail with type error",
			matcher:     "$matchRegexp(^test$)",
			actual:      []string{},
			wantErr:     "type mismatch:\n     expected: number / string\n       actual: array",
		},
		{
			description: "match regexp to map MUST fail with type error",
			matcher:     "$matchRegexp(^test$)",
			actual:      map[string]string{},
			wantErr:     "type mismatch:\n     expected: number / string\n       actual: map",
		},
		{
			description: "match regexp to bool MUST fail with type error",
			matcher:     "$matchRegexp(^test$)",
			actual:      true,
			wantErr:     "type mismatch:\n     expected: number / string\n       actual: bool",
		},
		{
			description: "match regexp to nil MUST fail with type error",
			matcher:     "$matchRegexp(^test$)",
			actual:      nil,
			wantErr:     "type mismatch:\n     expected: number / string\n       actual: nil",
		},
		{
			description: "match regexp to invalid type MUST fail with type error",
			matcher:     "$matchRegexp(^test$)",
			actual:      t,
			wantErr:     "type mismatch:\n     expected: number / string\n       actual: *testing.T",
		},
	}

	processTests(t, tests, Params{})
}

func Test_regexpMatcher_IgnoreValues(t *testing.T) {
	tests := []matcherTest{
		{
			description: "WHEN IgnoreValues specified $matchRegexp MUST be ignored with scalar type",
			matcher:     "$matchRegexp(^one$)",
			actual:      "two",
		},
		{
			description: "WHEN IgnoreValues specified and $matchRegexp compares with non-scalar type test MUST fail",
			matcher:     "$matchRegexp(^one$)",
			actual:      []string{"two"},
			wantErr:     "type mismatch:\n     expected: number / string\n       actual: array",
		},
	}

	processTests(t, tests, Params{
		IgnoreValues: true,
	})
}
