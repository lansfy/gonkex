package yaml_file

import (
	"encoding/json"
)

type VariablesToSet map[int]map[string]string

/*
There can be two types of data in yaml-file:

1. JSON-paths:

	VariablesToSet:
	   <code1>:
	      <varName1>: <JSON_Path1>
	      <varName2>: <JSON_Path2>

2. Plain text:

	VariablesToSet:
	   <code1>: <varName1>
	   <code2>: <varName2>

In this case we unmarshall values to format similar to JSON-paths format with empty paths:

	VariablesToSet:
	   <code1>:
	      <varName1>: ""
	   <code2>:
	      <varName2>: ""
*/
func (v *VariablesToSet) UnmarshalYAML(unmarshal func(interface{}) error) error {
	res := map[int]map[string]string{}

	// try to unmarshall as plain text
	var plain map[int]string
	if err := unmarshal(&plain); err == nil {
		for code, varName := range plain {
			res[code] = map[string]string{
				varName: "",
			}
		}

		*v = res
		return nil
	}

	// json-paths
	if err := unmarshal(&res); err != nil {
		return err
	}

	*v = res
	return nil
}

func (v *VariablesToSet) UnmarshalJSON(data []byte) error {
	unmarshal := func(v interface{}) error {
		return json.Unmarshal(data, v)
	}
	return v.UnmarshalYAML(unmarshal)
}
