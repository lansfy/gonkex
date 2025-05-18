package compare

import (
	"fmt"
	"strings"
)

func extractArgs(input string, defaultParams map[string]string) (string, map[string]string, error) {
	parts := strings.Split(input, ",")
	baseValue := ""
	if len(parts) != 0 {
		baseValue = parts[0]
	}

	params := map[string]string{}
	for i := 1; i < len(parts); i++ {
		keyValue := strings.SplitN(parts[i], "=", 2)
		if len(keyValue) != 2 {
			return "", nil, makeParamError(parts[i], "invalid parameter format")
		}

		key := strings.TrimSpace(keyValue[0])
		if key == "" {
			return "", nil, makeParamError(parts[i], "empty parameter name")
		}

		if _, ok := defaultParams[key]; !ok {
			return "", nil, makeParamError(parts[i], "unknown parameter name")
		}

		if _, ok := params[key]; ok {
			return "", nil, makeParamError(parts[i], "duplicate parameter name")
		}

		params[key] = keyValue[1]
	}

	for key, value := range defaultParams {
		if _, ok := params[key]; !ok {
			params[key] = value
		}
	}

	return baseValue, params, nil
}

func makeParamError(part, message string) error {
	return fmt.Errorf("parameter '%s': %s", part, message)
}
