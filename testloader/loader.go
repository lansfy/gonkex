package testloader

import (
	"github.com/lansfy/gonkex/models"
)

// LoaderInterface defines the interface for loading test definitions from various sources.
type LoaderInterface interface {
	// Load discovers and parses test definitions from the configured source.
	// It returns a slice of TestInterface implementations representing individual test cases.
	//
	// Returns:
	// - []models.TestInterface: A slice of parsed test cases ready for execution
	// - error: An error if the loading process fails (e.g., source not found, parse errors)
	Load() ([]models.TestInterface, error)

	// SetFilter applies a filter to control which test definitions are loaded.
	// The filter behavior is implementation-specific but typically involves pattern matching
	// against test file names, test names, or other test metadata.
	//
	// Parameters:
	// - filterFunc: The function which should return true for valid file.
	SetFilter(filterFunc func(fileName string) bool)
}
