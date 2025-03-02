package types

import (
	"fmt"
	"strings"

	"sigs.k8s.io/yaml"
)

type yamlBodyType struct{}

func (b *yamlBodyType) GetName() string {
	return "YAML"
}

func (b *yamlBodyType) IsSupportedContentType(contentType string) bool {
	return strings.Contains(contentType, "yaml")
}

func (b *yamlBodyType) Decode(body string) (interface{}, error) {
	jsonBody, err := yaml.YAMLToJSON([]byte(body))
	if err != nil {
		return nil, err
	}
	return decodeJSON(string(jsonBody))
}

func (b *yamlBodyType) ExtractResponseValue(body, path string) (string, error) {
	parsed, err := yaml.YAMLToJSON([]byte(body))
	if err != nil {
		return "", fmt.Errorf("invalid YAML in response: %w", err)
	}
	return getStringFromJSON(string(parsed), path)
}
