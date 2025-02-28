package yaml_file

import (
	"strings"
)

func performQuery(val string, perform func(string) string) string {
	val = perform(val)
	var query strings.Builder
	query.Grow(len(val) + 1)
	if val != "" && val[0] != '?' {
		_, _ = query.WriteString("?")
	}
	_, _ = query.WriteString(val)
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

func performStringMap(src map[string]string, perform func(string) string) map[string]string {
	dst := make(map[string]string, len(src))
	for k, v := range src {
		dst[k] = perform(v)
	}
	return dst
}

func performForm(form *Form, perform func(string) string) *Form {
	return &Form{
		Files:  performStringMap(form.Files, perform),
		Fields: performStringMap(form.Fields, perform),
	}
}

func performHeaders(headers map[string]string, perform func(string) string) map[string]string {
	return performStringMap(headers, perform)
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
