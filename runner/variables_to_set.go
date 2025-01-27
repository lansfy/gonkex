package runner

import (
	"fmt"

	"github.com/tidwall/gjson"
)

func ExtractVariablesFromResponse(varsToSet map[string]string, body string, isJSON bool) (map[string]string, error) {
	vars := map[string]string{}
	names, paths := split(varsToSet)
	var err error
	if isJSON {
		err = fromJSON(vars, names, paths, body)
	} else {
		err = fromPlainText(vars, names, body)
	}
	if err != nil {
		return nil, err
	}

	return vars, nil
}

func fromJSON(vars map[string]string, names, paths []string, body string) error {
	for n, res := range gjson.GetMany(body, paths...) {
		if !res.Exists() {
			return fmt.Errorf("path '%s' does not exist in given json", paths[n])
		}
		vars[names[n]] = res.String()
	}

	return nil
}

func fromPlainText(vars map[string]string, names []string, body string) error {
	if len(names) == 1 {
		vars[names[0]] = body
		return nil
	}

	return fmt.Errorf(
		"count of variables for plain-text response should be 1, %d given",
		len(names),
	)
}

// split returns keys and values of given map as separate slices
func split(m map[string]string) (keys, values []string) {
	values = make([]string, 0, len(m))
	keys = make([]string, 0, len(m))

	for k, v := range m {
		keys = append(keys, k)
		values = append(values, v)
	}

	return keys, values
}
