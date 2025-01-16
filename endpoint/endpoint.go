package endpoint

var Prefix = "/gonkex/"

type Helper interface {
	GetPath() string

	GetRequestAsJson(v interface{}) error
	GetRequestAsYaml(v interface{}) error
	GetRequestAsBytes() ([]byte, error)

	GetMockAddr(name string) string

	SetResponseAsJson(response interface{}) error
	SetResponseAsYaml(response interface{}) error
	SetResponseAsBytes(response []byte) error

	SetStatusCode(code int)
}

type Endpoint func(Helper) error
type EndpointMap map[string]Endpoint
