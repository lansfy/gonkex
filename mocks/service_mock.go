package mocks

import (
	"context"
	"net"
	"net/http"
	"sync"

	"github.com/lansfy/gonkex/colorize"
)

// ServiceMock represents a mock HTTP service for testing purposes
type ServiceMock struct {
	server            *http.Server
	listener          net.Listener
	mock              *Definition
	defaultDefinition *Definition
	mutex             sync.RWMutex
	errors            []error
	checkers          []CheckerInterface

	ServiceName string
}

// NewServiceMock creates a new ServiceMock instance with the given name and mock definition.
// If the mock definition is nil, it creates a default definition with a fail reply.
func NewServiceMock(serviceName string, mock *Definition) *ServiceMock {
	if mock == nil {
		mock = NewDefinition("$", nil, NewFailReply(), CallsNoConstraint, OrderNoValue)
	}
	return &ServiceMock{
		mock:              mock,
		defaultDefinition: mock,
		ServiceName:       serviceName,
	}
}

// StartServer initializes and starts an HTTP server on localhost with a random port.
// After starting the server, you can use ServerAddr() method to get the actual
// address (including the randomly assigned port) where the server is listening.
func (m *ServiceMock) StartServer() error {
	return m.StartServerWithAddr("localhost:0") // loopback, random port
}

// StartServerWithAddr initializes and starts an HTTP server on the specified address.
func (m *ServiceMock) StartServerWithAddr(addr string) error {
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	m.listener = ln
	m.server = &http.Server{
		Addr:    addr,
		Handler: m,
	}
	go func() {
		_ = m.server.Serve(ln)
	}()
	return nil
}

// ShutdownServer gracefully stops the HTTP server using the provided context
func (m *ServiceMock) ShutdownServer(ctx context.Context) error {
	err := m.server.Shutdown(ctx)
	m.listener = nil
	m.server = nil
	return err
}

// ServerAddr returns the actual address (including port) where the mock server is listening.
// Panics if the server hasn't been started.
func (m *ServiceMock) ServerAddr() string {
	if m.listener == nil {
		panic("mock server " + m.ServiceName + " is not started")
	}
	return m.listener.Addr().String()
}

// ServeHTTP handles incoming HTTP requests by executing the mock definition and running checkers.
// Errors from both Definition execution and checkers are accumulated.
// Implements the http.Handler interface.
func (m *ServiceMock) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if m.mock == nil {
		return
	}

	body, err := getRequestBodyCopy(r)
	if err != nil {
		m.errors = append(m.errors, err)
		return
	}

	wrap := createResponseWriterProxy(w)
	m.errors = append(m.errors, m.mock.Execute(wrap, r)...)

	for _, c := range m.checkers {
		setRequestBody(r, body)
		errs := c.CheckRequest(m.ServiceName, r, wrap.CreateHttpResponse()) // nolint:bodyclose // we have single copy of data
		m.errors = append(m.errors, errs...)
	}

	if err := wrap.Flush(); err != nil {
		m.errors = append(m.errors, err)
	}
}

// SetDefinition replaces the current mock definition with a new one
func (m *ServiceMock) SetDefinition(newDefinition *Definition) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.mock = newDefinition
}

// ResetDefinition restores the original default definition
func (m *ServiceMock) ResetDefinition() {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.mock = m.defaultDefinition
}

// ResetRunningContext clears all accumulated errors and resets the mock definition's running context
func (m *ServiceMock) ResetRunningContext() {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.errors = nil
	m.mock.ResetRunningContext()
}

// EndRunningContext finalizes the running context and returns all accumulated errors
func (m *ServiceMock) EndRunningContext() []error {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	errs := m.errors
	errs = append(errs, m.mock.EndRunningContext()...)
	for i := range errs {
		errs[i] = colorize.NewEntityError("mock %s", m.ServiceName).SetSubError(errs[i])
	}
	return errs
}

// SetCheckers configures the request checkers used to validate incoming HTTP requests
func (m *ServiceMock) SetCheckers(checkers []CheckerInterface) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.checkers = checkers
}
