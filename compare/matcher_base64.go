package compare

import (
	"encoding/base64"

	"github.com/lansfy/gonkex/colorize"
)

type base64Matcher struct {
	data string
}

func (r *base64Matcher) MatchValues(actual interface{}) error {
	actualStr, ok := actual.(string)
	if !ok {
		return makeTypeMismatchError(string(leafString), string(getLeafType(actual)))
	}

	decoded, err := base64.StdEncoding.DecodeString(actualStr)
	if err != nil {
		return colorize.NewNotEqualError("cannot make base64 decode:", nil, err.Error())
	}

	if string(decoded) != r.data {
		return colorize.NewNotEqualError("base64 decoded value does not match:",
			r.data, string(decoded))
	}
	return nil
}
