package fixtures

import (
	"sort"

	"gopkg.in/yaml.v3"
)

func DumpCollection(coll []*Collection, addType bool) ([]byte, error) {
	yamlCollections := MapSlice{}
	for _, c := range coll {
		yamlCollections = append(yamlCollections, MapItem{
			Key:   c.Name,
			Value: dumpCollectionItems(c, addType),
		})
	}

	out, err := yaml.Marshal(yamlCollections)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func dumpCollectionItems(coll *Collection, addType bool) []MapSlice {
	var result []MapSlice
	for _, item := range coll.Items {
		fields := make([]string, 0, len(item))
		for name := range item {
			fields = append(fields, name)
		}
		sort.Strings(fields)

		rowValues := MapSlice{}
		for _, name := range fields {
			rowValues = append(rowValues, MapItem{
				Key:   name,
				Value: item[name],
			})
		}
		if addType {
			rowValues = append(rowValues, MapItem{
				Key:   "_type",
				Value: coll.Type,
			})
		}
		result = append(result, rowValues)
	}
	return result
}
