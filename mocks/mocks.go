package mocks

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/lansfy/gonkex/colorize"
)

type Mocks struct {
	mocks map[string]*ServiceMock
}

func New(mocks ...*ServiceMock) *Mocks {
	mocksMap := make(map[string]*ServiceMock, len(mocks))
	for _, v := range mocks {
		mocksMap[v.ServiceName] = v
	}
	return &Mocks{
		mocks: mocksMap,
	}
}

func NewNop(serviceNames ...string) *Mocks {
	mocksMap := make(map[string]*ServiceMock, len(serviceNames))
	for _, name := range serviceNames {
		mocksMap[name] = NewServiceMock(name, nil)
	}
	return &Mocks{
		mocks: mocksMap,
	}
}

func (m *Mocks) ResetDefinitions() {
	for _, v := range m.mocks {
		v.ResetDefinition()
	}
}

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

func (m *Mocks) ShutdownContext(ctx context.Context) error {
	errs := make([]string, 0, len(m.mocks))
	for _, v := range m.mocks {
		if err := v.ShutdownServer(ctx); err != nil {
			errs = append(errs, fmt.Sprintf("%s: %s", v.mock.path, err.Error()))
		}
	}
	if len(errs) != 0 {
		return fmt.Errorf("%s", strings.Join(errs, "; "))
	}
	return nil
}

func (m *Mocks) SetMock(mock *ServiceMock) {
	m.mocks[mock.ServiceName] = mock
}

func (m *Mocks) Service(serviceName string) *ServiceMock {
	mock := m.mocks[serviceName]
	return mock
}

func (m *Mocks) ResetRunningContext() {
	for _, v := range m.mocks {
		v.ResetRunningContext()
	}
}

func (m *Mocks) EndRunningContext() []error {
	var errors []error
	for _, v := range m.mocks {
		errors = append(errors, v.EndRunningContext()...)
	}
	return errors
}

func (m *Mocks) GetNames() []string {
	names := []string{}
	for n := range m.mocks {
		names = append(names, n)
	}
	return names
}

func (m *Mocks) LoadDefinitions(loader Loader, definitions map[string]interface{}) error {
	if m == nil && len(definitions) != 0 {
		return fmt.Errorf("object Mocks has nil value, but mock required for test")
	}
	for serviceName, definition := range definitions {
		service := m.Service(serviceName)
		if service == nil {
			return fmt.Errorf("unknown mock name: %s", serviceName)
		}

		def, err := loader.LoadDefinition(definition)
		if err != nil {
			return colorize.NewEntityError("load definition for %s", serviceName).SetSubError(err)
		}
		service.SetDefinition(def)
	}
	return nil
}

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
