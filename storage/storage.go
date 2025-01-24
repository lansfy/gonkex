package storage

import (
	"encoding/json"
)

// StorageInterface defines a storage abstraction that supports querying and loading fixtures.
type StorageInterface interface {
	// GetType returns the type of storage being used (e.g., "postgresql", "redis", etc.).
	GetType() string

	// LoadFixtures loads data fixtures into the storage from a specified location.
	// location: Path to the directory containing the fixtures.
	// names: List of fixture file names to be loaded.
	// Returns an error if the fixtures cannot be loaded.
	LoadFixtures(location string, names []string) error

	// ExecuteQuery executes a query against the storage and retrieves the results.
	// query: The query string to execute.
	// Returns a slice of json.RawMessage representing the query results and an error if the query fails.
	ExecuteQuery(query string) ([]json.RawMessage, error)
}
