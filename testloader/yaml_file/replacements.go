package yaml_file

import (
	"strings"

	"github.com/lansfy/gonkex/models"
)

func performQuery(val string, perform func(string) string) string {
	val = perform(val)
	var query strings.Builder
	query.Grow(len(val) + 1)
	if val != "" && val[0] != '?' {
		query.WriteString("?")
	}
	query.WriteString(val)
	return query.String()
}

func performInterface(value interface{}, perform func(string) string) {
	if mapValue, ok := value.(map[interface{}]interface{}); ok {
		for key := range mapValue {
			if strValue, ok := mapValue[key].(string); ok {
				mapValue[key] = perform(strValue)
			} else {
				performInterface(mapValue[key], perform)
			}
		}
	}
	if arrValue, ok := value.([]interface{}); ok {
		for idx := range arrValue {
			if strValue, ok := arrValue[idx].(string); ok {
				arrValue[idx] = perform(strValue)
			} else {
				performInterface(arrValue[idx], perform)
			}
		}
	}
}

func performForm(form *models.Form, perform func(string) string) *models.Form {
	files := make(map[string]string, len(form.Files))

	for k, v := range form.Files {
		files[k] = perform(v)
	}

	return &models.Form{Files: files}
}

func performHeaders(headers map[string]string, perform func(string) string) map[string]string {
	res := make(map[string]string)

	for k, v := range headers {
		res[k] = perform(v)
	}

	return res
}

func performResponses(responses map[int]string, perform func(string) string) map[int]string {
	res := make(map[int]string)

	for k, v := range responses {
		res[k] = perform(v)
	}

	return res
}

func performDbResponses(responses []string, perform func(string) string) []string {
	if responses == nil {
		return nil
	}

	res := make([]string, len(responses))

	for idx, v := range responses {
		res[idx] = perform(v)
	}

	return res
}
