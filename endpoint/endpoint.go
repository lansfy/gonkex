package endpoint

import (
	"net/http"
)

// DefaultPrefix is default prefix for all HelperEndpoints in case if you not override this value in Runner configuration.
const DefaultPrefix = "/gonkex/"

type Helper interface {
	// GetHeaders returns all request headers.
	GetHeaders() map[string][]string
	// GetPath returns request path without Prefix
	GetPath() string

	// GetRequestAsJson unmarshals the request bytes into the provided object.
	GetRequestAsJson(v interface{}) error
	// GetRequestAsYaml unmarshals the request bytes into the provided object as YAML.
	GetRequestAsYaml(v interface{}) error
	// GetRequestAsBytes returns the raw request bytes.
	GetRequestAsBytes() ([]byte, error)

	// GetMocksTransport returns http.RoundTripper, which routes the request to the
	// appropriate mock service based on the hostname in the request URL.
	GetMocksTransport() http.RoundTripper
	// GetMockAddr returns address of mock with specified name
	GetMockAddr(name string) string
	// GetMeta returns meta field value from current test.
	GetMeta(key string) interface{}

	// SetResponseAsJson marshals the provided object into JSON and stores it as the response.
	SetResponseAsJson(response interface{}) error
	// SetResponseAsYaml marshals the provided object into YAML and stores it as the response.
	SetResponseAsYaml(response interface{}) error
	// SetResponseAsBytes sets the raw response bytes.
	SetResponseAsBytes(response []byte)

	// SetStatusCode sets the HTTP response status code.
	SetStatusCode(code int)
	// SetContentType sets custom content type to HTTP response.
	SetContentType(contentType string)
}

type Endpoint func(Helper) error
type EndpointMap map[string]Endpoint
