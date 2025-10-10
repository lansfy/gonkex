package compare

import (
	"testing"
)

func Test_base64Matcher_MatchValues(t *testing.T) {
	tests := []matcherTest{
		{
			description: "valid base64 match",
			matcher:     "$matchBase64(somevalue)",
			actual:      "c29tZXZhbHVl", // encoded "somevalue"
		},
		{
			description: "valid base64, but no match",
			matcher:     "$matchBase64(somevalue)",
			actual:      "b3RoZXJ2YWx1ZQ==", // encoded "othervalue"
			wantErr:     "base64 decoded value does not match:\n     expected: somevalue\n       actual: othervalue",
		},
		{
			description: "invalid base64",
			matcher:     "$matchBase64(somevalue)",
			actual:      "inva$$id",
			wantErr:     "cannot make base64 decode:\n     expected: <nil>\n       actual: illegal base64 data at input byte 4",
		},
	}
	processTests(t, tests)
}

func Test_base64Matcher_UnsupportedTypes(t *testing.T) {
	tests := []matcherTest{
		{
			description: "match base64 to number MUST fail with type error",
			matcher:     "$matchBase64(somevalue)",
			actual:      12345,
			wantErr:     "type mismatch:\n     expected: string\n       actual: number",
		},
		{
			description: "match base64 to array MUST fail with type error",
			matcher:     "$matchBase64(somevalue)",
			actual:      []string{},
			wantErr:     "type mismatch:\n     expected: string\n       actual: array",
		},
		{
			description: "match base64 to map MUST fail with type error",
			matcher:     "$matchBase64(somevalue)",
			actual:      map[string]string{},
			wantErr:     "type mismatch:\n     expected: string\n       actual: map",
		},
		{
			description: "match base64 to bool MUST fail with type error",
			matcher:     "$matchBase64(somevalue)",
			actual:      true,
			wantErr:     "type mismatch:\n     expected: string\n       actual: bool",
		},
		{
			description: "match base64 to nil MUST fail with type error",
			matcher:     "$matchBase64(somevalue)",
			actual:      nil,
			wantErr:     "type mismatch:\n     expected: string\n       actual: nil",
		},
		{
			description: "match base64 to invalid type MUST fail with type error",
			matcher:     "$matchBase64(somevalue)",
			actual:      t,
			wantErr:     "type mismatch:\n     expected: string\n       actual: *testing.T",
		},
	}
	processTests(t, tests)
}
