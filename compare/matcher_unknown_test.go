package compare

import (
	"testing"
)

func Test_unknownMatcher_MatchValues(t *testing.T) {
	tests := []matcherTest{
		{
			description: "WHEN use unknown matcher test MUST fail with error",
			matcher:     "$matchNameWithError(somevalue)",
			actual:      "12345",
			wantErr:     "parse '$matchNameWithError': unknown matcher name",
		},
		{
			description: "WHEN use $matchArray in wrong place test MUST fail with error",
			matcher:     "$matchArray(somevalue)",
			actual:      "12345",
			wantErr:     "parse '$matchArray': must be first element in array",
		},
	}
	processTests(t, tests, Params{})
}
