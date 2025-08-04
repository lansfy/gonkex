package fixtures

import (
	"sort"

	"gopkg.in/yaml.v2"
)

func GenerateYamlResult(loader ContentLoader, names []string, opts *LoadDataOpts) ([]byte, error) {
	coll, err := LoadData(loader, names, opts)
	if err != nil {
		return nil, err
	}

	yamlCollections := yaml.MapSlice{}
	for _, c := range coll {
		yamlCollections = append(yamlCollections, yaml.MapItem{
			Key:   c.Name,
			Value: generateDatabaseItems(c.Items),
		})
	}

	out, err := yaml.Marshal(yamlCollections)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func generateDatabaseItems(items []Item) []yaml.MapSlice {
	var result []yaml.MapSlice
	for _, item := range items {
		fields := make([]string, 0, len(item))
		for name := range item {
			fields = append(fields, name)
		}
		sort.Strings(fields)

		rowValues := yaml.MapSlice{}
		for _, name := range fields {
			rowValues = append(rowValues, yaml.MapItem{
				Key:   name,
				Value: item[name],
			})
		}

		result = append(result, rowValues)
	}
	return result
}
