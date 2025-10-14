package compare

import (
	"encoding/base64"

	"github.com/lansfy/gonkex/colorize"
)

func createBase64Matcher(args string) Matcher {
	return &base64Matcher{args}
}

type base64Matcher struct {
	data string
}

func (r *base64Matcher) MatchValues(actual interface{}) error {
	actualStr, ok := actual.(string)
	if !ok {
		return makeTypeMismatchError([]leafType{leafString}, getLeafType(actual))
	}

	decoded, err := base64.StdEncoding.DecodeString(actualStr)
	if err != nil {
		return colorize.NewNotEqualError("cannot make base64 decode:", nil, err.Error())
	}

	if matcher := CreateMatcher(r.data); matcher != nil {
		return matcher.MatchValues(string(decoded))
	}

	if string(decoded) == r.data {
		return nil
	}

	return colorize.NewNotEqualError("base64 decoded value does not match:", r.data, string(decoded))
}
