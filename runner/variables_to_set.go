package runner

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/lansfy/gonkex/colorize"
	"github.com/lansfy/gonkex/models"

	"github.com/tidwall/gjson"
)

func extractVariablesFromResponse(varsToSet map[string]string, result *models.Result) (map[string]string, error) {
	vars := map[string]string{}
	for name, path := range varsToSet {
		value, err := processPath(path, result)
		if err != nil {
			return nil, colorize.NewEntityError("variable %s", name).SetSubError(err)
		}
		vars[name] = value
	}
	return vars, nil
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
		if path == "" {
			return result.ResponseBody, nil
		}
		isJSON := strings.Contains(result.ResponseContentType, "json") && result.ResponseBody != ""
		if isJSON {
			return getStringFromJSON(result.ResponseBody, path)
		}
		return "", fmt.Errorf("paths not supported for plain text body")
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
		return "", colorize.NewError("path %s does not exist in service response", colorize.Green(path))
	}
	return res.String(), nil
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
