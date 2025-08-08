package fixtures

import (
	"gopkg.in/yaml.v3"
)

type MapItem struct {
	Key   string
	Value interface{}
}

type MapSlice []MapItem

// MarshalYAML ensures keys are marshaled in insertion order.
func (ms MapSlice) MarshalYAML() (interface{}, error) {
	node := &yaml.Node{
		Kind:    yaml.MappingNode,
		Content: make([]*yaml.Node, 0, len(ms)*2),
	}

	for _, item := range ms {
		keyNode := &yaml.Node{
			Kind:  yaml.ScalarNode,
			Tag:   "!!str",
			Value: item.Key,
		}

		valNode := &yaml.Node{}
		if err := valNode.Encode(item.Value); err != nil {
			return nil, err
		}

		node.Content = append(node.Content, keyNode, valNode)
	}

	return node, nil
}

// UnmarshalYAML reads keys in order into MapSlice.
func (ms *MapSlice) UnmarshalYAML(value *yaml.Node) error {
	if value.Kind != yaml.MappingNode {
		return nil
	}

	out := make([]MapItem, 0, len(value.Content)/2)
	for i := 0; i < len(value.Content); i += 2 {
		key := value.Content[i]
		val := value.Content[i+1]

		var v interface{}
		if err := val.Decode(&v); err != nil {
			return err
		}

		out = append(out, MapItem{
			Key:   key.Value,
			Value: v,
		})
	}

	*ms = out
	return nil
}
