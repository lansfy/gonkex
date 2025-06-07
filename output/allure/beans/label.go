package beans

type Label struct {
	Name  string `xml:"name,attr"`
	Value string `xml:"value,attr"`
}

func NewLabel(name, value string) *Label {
	return &Label{
		Name:  name,
		Value: value,
	}
}
