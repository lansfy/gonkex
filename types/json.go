package types

import (
	"encoding/json"
	"strings"

	"github.com/lansfy/gonkex/colorize"

	"github.com/tidwall/gjson"
)

type jsonBodyType struct{}

func (b *jsonBodyType) GetName() string {
	return "JSON"
}

func (b *jsonBodyType) IsSupportedContentType(contentType string) bool {
	return strings.Contains(contentType, "json")
}

func (b *jsonBodyType) Decode(body string) (interface{}, error) {
	return decodeJSON(body)
}

func (b *jsonBodyType) ExtractResponseValue(body, path string) (string, error) {
	return getStringFromJSON(body, path)
}

func decodeJSON(body string) (interface{}, error) {
	var expected interface{}
	err := json.Unmarshal([]byte(body), &expected)
	return expected, err
}

func getStringFromJSON(body, path string) (string, error) {
	res := gjson.Get(body, path)
	if !res.Exists() {
		return "", colorize.NewError("path %s does not exist in service response", colorize.Cyan("$."+path))
	}
	return res.String(), nil
}
