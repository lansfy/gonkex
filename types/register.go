package types

var (
	jbType = &jsonBodyType{}
	xbType = &xmlBodyType{}
	ybType = &yamlBodyType{}
)

func GetRegisteredBodyTypes() []BodyType {
	return []BodyType{jbType, xbType, ybType}
}
