package compare

import (
	"encoding/base64"
	"fmt"

	"github.com/lansfy/gonkex/colorize"
)

type base64Matcher struct {
	data string
}

func (r *base64Matcher) MatchValues(description, entity string, actual interface{}) error {
	decoded, err := base64.StdEncoding.DecodeString(fmt.Sprintf("%v", actual))
	if err != nil {
		return colorize.NewNotEqualError(description+" cannot make base64 decode:", entity, nil, err.Error())
	}

	if string(decoded) != r.data {
		return colorize.NewNotEqualError(description+" base64 decoded value does not match:",
			entity, r.data, string(decoded))
	}
	return nil
}
