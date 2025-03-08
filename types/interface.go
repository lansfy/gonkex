package types

// BodyType defines an interface for handling service response bodies.
type BodyType interface {
	// GetName returns the name of the body type (e.g., JSON, XML, YAML).
	GetName() string

	// IsSupportedContentType checks if the given content type is supported by this body type.
	// Returns true if supported, otherwise false.
	IsSupportedContentType(contentType string) bool

	// Decode parses the given body string and returns the decoded data as an interface{}.
	// Returns an error if the decoding process fails.
	Decode(body string) (interface{}, error)

	// ExtractResponseValue extracts a specific value from the body based on the provided path.
	// Returns the extracted value as a string or an error if extraction fails.
	ExtractResponseValue(body, path string) (string, error)
}
