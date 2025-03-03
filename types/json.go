package types

import (
	"encoding/json"
	"fmt"
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
	if err != nil {
		return nil, fmt.Errorf("json: %w", err)
	}
	return expected, nil
}

func getStringFromJSON(body, path string) (string, error) {
	_, err := decodeJSON(body)
	if err != nil {
		return "", err
	}
	res := gjson.Get(body, path)
	if !res.Exists() {
		return "", colorize.NewError("path %s does not exist in service response", colorize.Cyan("$."+path))
	}
	return res.String(), nil
}
