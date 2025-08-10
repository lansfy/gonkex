package response_body

import (
	"errors"
	"fmt"
	"net/http"
	"sort"
	"strings"

	"github.com/lansfy/gonkex/colorize"
	"github.com/lansfy/gonkex/models"
	"github.com/lansfy/gonkex/types"
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
	var errs []error
	for _, name := range keys {
		path := varsToSet[name]
		value, err := processPath(path, result)
		if err != nil {
			errs = append(errs, colorize.NewEntityError("variable %s", name).SetSubError(err))
		} else {
			vars[name] = value
		}
	}
	return vars, errs
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
		if result.ResponseBody == "" {
			return "", errors.New("paths not supported for empty body")
		}
		for _, b := range types.GetRegisteredBodyTypes() {
			if b.IsSupportedContentType(result.ResponseContentType) {
				return b.ExtractResponseValue(result.ResponseBody, path)
			}
		}
		return "", errors.New("paths not supported for plain text body")
	case "header":
		result.ShowHeaders = true // output should show request headers
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

func parseSetCookies(header string) *http.Cookie {
	parts := strings.Split(header, ";")

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
