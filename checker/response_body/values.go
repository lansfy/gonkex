package response_body

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"strings"

	"github.com/lansfy/gonkex/colorize"
	"github.com/lansfy/gonkex/models"
	"github.com/lansfy/gonkex/xmlparsing"

	"github.com/tidwall/gjson"
	"sigs.k8s.io/yaml"
)

func ExtractValues(varsToSet map[string]string, result *models.Result) (map[string]string, []error) {
	// sort keys
	keys := make([]string, 0, len(varsToSet))
	for k := range varsToSet {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// process variables
	vars := map[string]string{}
	var errors []error
	for _, name := range keys {
		path := varsToSet[name]
		value, err := processPath(path, result)
		if err != nil {
			errors = append(errors, colorize.NewEntityError("variable %s", name).SetSubError(err))
		} else {
			vars[name] = value
		}
	}
	return vars, errors
}

func isJSONResponseBody(result *models.Result) bool {
	return strings.Contains(result.ResponseContentType, "json")
}

func isXMLResponseBody(result *models.Result) bool {
	return strings.Contains(result.ResponseContentType, "xml")
}

func isYAMLResponseBody(result *models.Result) bool {
	return strings.Contains(result.ResponseContentType, "yaml")
}

func processPath(path string, result *models.Result) (string, error) {
	prefix := "body"
	parts := strings.SplitN(path, ":", 2)
	if len(parts) == 2 {
		prefix = parts[0]
		path = strings.Trim(parts[1], " ")
	}

	switch prefix {
	case "body":
		switch {
		case path == "":
			return result.ResponseBody, nil
		case result.ResponseBody == "":
			return "", fmt.Errorf("paths not supported for empty body")
		case isJSONResponseBody(result):
			return getStringFromJSON(result.ResponseBody, path)
		case isXMLResponseBody(result):
			return getStringFromXML(result.ResponseBody, path)
		case isYAMLResponseBody(result):
			return getStringFromYAML(result.ResponseBody, path)
		default:
			return "", fmt.Errorf("paths not supported for plain text body")
		}
	case "header":
		if valArr := result.ResponseHeaders[path]; len(valArr) != 0 {
			return valArr[0], nil
		}
		return "", colorize.NewEntityError("response does not include expected header %s", path)
	case "cookie":
		valArr, ok := result.ResponseHeaders["Set-Cookie"]
		if !ok {
			return "", colorize.NewEntityError("response does not include expected header %s", "Set-Cookie")
		}

		for _, line := range valArr {
			if cookie := parseSetCookies(line); cookie != nil {
				if cookie.Name == path {
					return cookie.Value, nil
				}
			}
		}
		return "", colorize.NewError("%s header does not include expected cookie %s",
			colorize.Cyan("Set-Cookie"), colorize.Green(path))
	default:
		return "", fmt.Errorf("unexpected path prefix '%s' (allowed only [body header cookie])", prefix)
	}
}

func getStringFromJSON(body, path string) (string, error) {
	res := gjson.Get(body, path)
	if !res.Exists() {
		return "", colorize.NewError("path %s does not exist in service response", colorize.Cyan("$."+path))
	}
	return res.String(), nil
}

func getStringFromXML(body, path string) (string, error) {
	parsed, err := xmlparsing.Parse(body)
	if err != nil {
		return "", fmt.Errorf("invalid XML in response: %w", err)
	}
	plainParsed, _ := json.Marshal(parsed)
	return getStringFromJSON(string(plainParsed), path)
}

func getStringFromYAML(body, path string) (string, error) {
	parsed, err := yaml.YAMLToJSON([]byte(body))
	if err != nil {
		return "", fmt.Errorf("invalid YAML in response: %w", err)
	}
	return getStringFromJSON(string(parsed), path)
}

func parseSetCookies(header string) *http.Cookie {
	parts := strings.Split(header, ";")
	if len(parts) == 0 {
		return nil
	}

	// First part is always "key=value"
	keyValue := strings.SplitN(strings.TrimSpace(parts[0]), "=", 2)
	if len(keyValue) != 2 {
		return nil
	}

	return &http.Cookie{
		Name:  strings.TrimSpace(keyValue[0]),
		Value: strings.TrimSpace(keyValue[1]),
	}
}
