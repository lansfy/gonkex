package xmlparsing

import (
	"encoding/xml"
)

func Parse(rawXML string) (map[string]interface{}, error) {
	var n node
	if err := xml.Unmarshal([]byte(rawXML), &n); err != nil {
		return nil, err
	}
	return buildMap([]node{n}), nil
}

type node struct {
	XMLName  xml.Name
	Attrs    []xml.Attr `xml:",any,attr"`
	Content  string     `xml:",chardata"`
	Children []node     `xml:",any"`
}

func buildMap(nodes []node) map[string]interface{} {
	result := make(map[string]interface{})

	for name, group := range regroupNodesByName(nodes) {
		switch len(group) {
		case 0:
			continue
		case 1:
			result[name] = buildNode(&group[0])
		default:
			result[name] = buildArray(group)
		}
	}

	return result
}

func buildArray(nodes []node) []interface{} {
	arr := make([]interface{}, len(nodes))
	for i := range nodes {
		arr[i] = buildNode(&nodes[i])
	}

	return arr
}

func buildNode(n *node) interface{} {
	hasAttrs := len(n.Attrs) > 0
	hasChildren := len(n.Children) > 0

	if hasAttrs && hasChildren {
		result := buildMap(n.Children)
		result["-attrs"] = buildAttributes(n.Attrs)
		return result
	}

	if hasAttrs {
		return map[string]interface{}{
			"-attrs":  buildAttributes(n.Attrs),
			"content": n.Content,
		}
	}

	if hasChildren {
		return buildMap(n.Children)
	}

	return n.Content
}

func buildAttributes(attrs []xml.Attr) map[string]string {
	m := make(map[string]string, len(attrs))
	for _, attr := range attrs {
		m[joinXMLName(attr.Name)] = attr.Value
	}

	return m
}

func regroupNodesByName(nodes []node) map[string][]node {
	grouped := make(map[string][]node)
	for _, n := range nodes {
		name := joinXMLName(n.XMLName)

		if _, ok := grouped[name]; !ok {
			grouped[name] = make([]node, 0)
		}

		grouped[name] = append(grouped[name], n)
	}

	return grouped
}

func joinXMLName(xmlName xml.Name) string {
	name := xmlName.Local
	if xmlName.Space != "" {
		name = xmlName.Space + ":" + name
	}

	return name
}
