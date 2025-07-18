package mocks

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/lansfy/gonkex/colorize"
)

var _ http.RoundTripper = (*Mocks)(nil)

// Mocks is a container for managing multiple ServiceMock instances.
// It provides centralized control over a collection of mock HTTP services,
// allowing them to be started, stopped, and configured as a group.
type Mocks struct {
	mocks map[string]*ServiceMock
}

// New creates a new Mocks instance from a list of ServiceMock objects.
func New(mocks ...*ServiceMock) *Mocks {
	m := &Mocks{map[string]*ServiceMock{}}
	for _, v := range mocks {
		m.SetMock(v)
	}
	return m
}

// NewNop creates a new Mocks instance with provided service names.
// Each service is initialized with a empty definition, which will fail reply.
func NewNop(serviceNames ...string) *Mocks {
	mocks := []*ServiceMock{}
	for _, name := range serviceNames {
		mocks = append(mocks, NewServiceMock(name, nil))
	}

	return New(mocks...)
}

// ResetDefinitions restores the original default definition for all mock services.
// This is useful for resetting the mocks between test cases.
func (m *Mocks) ResetDefinitions() {
	for _, v := range m.mocks {
		v.ResetDefinition()
	}
}

// Start initializes and starts HTTP servers for all mock services.
func (m *Mocks) Start() error {
	for _, v := range m.mocks {
		err := v.StartServer()
		if err != nil {
			m.Shutdown()
			return err
		}
	}
	return nil
}

// Stops immediately, with no gracefully closing connections
func (m *Mocks) Shutdown() {
	ctx, cancel := context.WithCancel(context.TODO())
	cancel()
	_ = m.ShutdownContext(ctx)
}

// ShutdownContext gracefully stops all mock servers using the provided context.
func (m *Mocks) ShutdownContext(ctx context.Context) error {
	errs := []string{}
	for _, v := range m.mocks {
		if !v.IsStarted() {
			continue
		}
		if err := v.ShutdownServer(ctx); err != nil {
			errs = append(errs, fmt.Sprintf("%s: %s", v.mock.path, err.Error()))
		}
	}
	if len(errs) != 0 {
		return fmt.Errorf("%s", strings.Join(errs, "; "))
	}
	return nil
}

// SetMock adds or replaces a ServiceMock in the internal map, indexed by its ServiceName.
func (m *Mocks) SetMock(mock *ServiceMock) {
	m.mocks[mock.ServiceName] = mock
}

// Service retrieves a ServiceMock by its name. Returns nil if the service does not exist.
func (m *Mocks) Service(serviceName string) *ServiceMock {
	mock := m.mocks[serviceName]
	return mock
}

// ResetRunningContext clears all accumulated errors and resets the running context for all mock services.
// This is typically called before starting a new test case.
func (m *Mocks) ResetRunningContext() {
	for _, v := range m.mocks {
		v.ResetRunningContext()
	}
}

// EndRunningContext finalizes the running context for all mock services and returns all accumulated errors.
// This is typically called after completing a test case to verify that all expectations were met.
func (m *Mocks) EndRunningContext() []error {
	var errors []error
	for _, v := range m.mocks {
		errors = append(errors, v.EndRunningContext()...)
	}
	return errors
}

// GetNames returns the names of all registered mock services.
func (m *Mocks) GetNames() []string {
	names := []string{}
	for n := range m.mocks {
		names = append(names, n)
	}
	return names
}

// RoundTrip implements the http.RoundTripper interface, allowing Mocks to be used
// as a transport for HTTP clients. It routes the request to the appropriate mock service
// based on the hostname in the request URL. If no matching service is found, it returns an error.
func (m *Mocks) RoundTrip(req *http.Request) (*http.Response, error) {
	host := req.URL.Hostname()
	service := m.Service(host)
	if service == nil {
		return nil, fmt.Errorf("unknown mock name '%s'", host)
	}
	return service.RoundTrip(req)
}

// LoadDefinitions loads mock definitions for multiple services at once using the provided loader.
// It updates each service with its corresponding definition from the map.
// Returns an error if any service is not found or if any definition fails to load.
func (m *Mocks) LoadDefinitions(loader Loader, definitions map[string]interface{}) error {
	for serviceName, definition := range definitions {
		service := m.Service(serviceName)
		if service == nil {
			return fmt.Errorf("unknown mock name '%s'", serviceName)
		}

		def, err := loader.LoadDefinition(definition)
		if err != nil {
			return colorize.NewEntityError("load definition for %s", serviceName).SetSubError(err)
		}
		service.SetDefinition(def)
	}
	return nil
}

// SetCheckers configures the request checkers for all mock services.
// These checkers are used to validate incoming HTTP requests against expectations.
func (m *Mocks) SetCheckers(checkers []CheckerInterface) {
	for _, v := range m.mocks {
		v.SetCheckers(checkers)
	}
}

// RegisterEnvironmentVariables sets environment variables for all mock services.
// It generates environment variable names using the given prefix and the service name,
// then assigns the corresponding server address for each mock service.
func (m *Mocks) RegisterEnvironmentVariables(prefix string) error {
	for _, name := range m.GetNames() {
		varName := strings.ToUpper(prefix + name)
		err := os.Setenv(varName, m.Service(name).ServerAddr())
		if err != nil {
			return fmt.Errorf("register environment variable %q: %w", varName, err)
		}
	}
	return nil
}
