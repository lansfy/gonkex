package endpoint

import (
	"net/http"
)

type Format int

const (
	FormatJson Format = iota
	FormatYaml
	FormatText
)

// DefaultPrefix is default prefix for all HelperEndpoints in case if you not override this value in Runner configuration.
const DefaultPrefix = "/gonkex/"

type Helper interface {
	// GetHeaders returns all request headers.
	GetHeaders() map[string][]string
	// GetPath returns request path without Prefix
	GetPath() string

	// GetRequest unmarshals the request body into the provided interface using the specified format (JSON or YAML).
	//
	// Parameters:
	//   - v: Pointer to the struct where the request data should be unmarshaled
	//   - format: The format to use for unmarshaling (FormatJson or FormatYaml)
	GetRequest(v interface{}, format Format) error
	// GetRequestRaw returns the raw request body bytes without any processing.
	GetRequestRaw() []byte

	// SetResponseFormat sets the format that will be used for response serialization.
	// This affects how SetResponse marshals data and what content type is set.
	SetResponseFormat(format Format)
	// SetResponse marshals the provided data into the response body using
	// the currently configured format (set via SetResponseFormat).
	//
	// Parameters:
	//   - v: The data to marshal into the response body
	SetResponse(v interface{}) error
	// SetResponseRaw sets the response body to the provided raw bytes.
	// This bypasses any marshaling and sets the response data directly.
	SetResponseRaw(response []byte)

	// SetStatusCode sets the HTTP response status code.
	SetStatusCode(code int)
	// SetContentType sets custom content type to HTTP response.
	SetContentType(contentType string)

	// GetMocksTransport returns http.RoundTripper, which routes the request to the
	// appropriate mock service based on the hostname in the request URL.
	GetMocksTransport() http.RoundTripper
	// GetMockAddr returns address of mock with specified name.
	GetMockAddr(name string) string
	// GetMeta returns meta field value from current test.
	GetMeta(key string) interface{}
}

type Endpoint func(Helper) error
type EndpointMap map[string]Endpoint
