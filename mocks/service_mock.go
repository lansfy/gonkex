package mocks

import (
	"context"
	"net"
	"net/http"
	"sync"

	"github.com/lansfy/gonkex/colorize"
)

var _ http.RoundTripper = (*ServiceMock)(nil)

// ServiceMock represents a mock HTTP service for testing purposes.
// ServiceMock helps with integration testing by simulating external HTTP services with configurable behavior.
// It can verify that requests match expected patterns and return predefined responses,
// allowing for controlled testing of code that interacts with external services.
//
// Each ServiceMock instance maintains its own HTTP server on a dynamically assigned port,
// tracks errors that occur during request processing, and can be configured with
// a set of checkers to validate incoming requests.
type ServiceMock struct {
	server            *http.Server
	listener          net.Listener
	mock              *Definition
	defaultDefinition *Definition
	mutex             sync.RWMutex
	errors            []error
	checkers          []CheckerInterface
	defaultPort       string

	ServiceName string
}

// NewServiceMock creates a new ServiceMock instance with the given name and mock definition.
// If the mock definition is nil, it creates a default definition with a fail reply.
func NewServiceMock(serviceName string, mock *Definition) *ServiceMock {
	name, port, _ := net.SplitHostPort(serviceName)
	if name != "" || port != "" {
		serviceName = name
	} else {
		port = "0" // random port
	}

	if mock == nil {
		mock = NewDefinition("$", nil, NewFailReply(), CallsNoConstraint, OrderNoValue)
	}
	return &ServiceMock{
		mock:              mock,
		defaultDefinition: mock,
		defaultPort:       port,
		ServiceName:       serviceName,
	}
}

// StartServer initializes and starts an HTTP server on localhost with a random port.
// After starting the server, you can use ServerAddr() method to get the actual
// address (including the randomly assigned port) where the server is listening.
func (m *ServiceMock) StartServer() error {
	return m.StartServerWithAddr("localhost:" + m.defaultPort) // loopback, random port
}

// StartServerWithAddr initializes and starts an HTTP server on the specified address.
func (m *ServiceMock) StartServerWithAddr(addr string) error {
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}

	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		server := &http.Server{
			Addr:    addr,
			Handler: m,
		}

		m.listener = ln
		m.server = server
		wg.Done()

		_ = server.Serve(ln)
	}()
	wg.Wait()
	return nil
}

// ShutdownServer gracefully stops the HTTP server using the provided context.
func (m *ServiceMock) ShutdownServer(ctx context.Context) error {
	server := m.server
	m.listener = nil
	m.server = nil
	if server == nil {
		return nil
	}
	return server.Shutdown(ctx)
}

// IsStarted returns true if mock service started.
func (m *ServiceMock) IsStarted() bool {
	return m.listener != nil
}

// ServerAddr returns the actual address (including port) where the mock server is listening.
// Panics if the server hasn't been started.
func (m *ServiceMock) ServerAddr() string {
	if !m.IsStarted() {
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

// RoundTrip implements the http.RoundTripper interface, allowing ServiceMock to be used as a transport for HTTP clients.
// This setup allows you to intercept outgoing HTTP requests in your tests and have them
// processed by your mock service instead of reaching the actual external service.
// The original URL is preserved for pattern matching in the mock definition, while
// the request is physically routed to the mock server's address.
func (m *ServiceMock) RoundTrip(req *http.Request) (*http.Response, error) {
	reqCopy := req.Clone(req.Context())
	reqCopy.URL.Host = m.ServerAddr()
	return http.DefaultTransport.RoundTrip(reqCopy)
}

// SetDefinition replaces the current mock definition with a new one.
func (m *ServiceMock) SetDefinition(newDefinition *Definition) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.mock = newDefinition
}

// ResetDefinition restores the original default definition.
func (m *ServiceMock) ResetDefinition() {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.mock = m.defaultDefinition
}

// ResetRunningContext clears all accumulated errors and resets the mock definition's running context.
func (m *ServiceMock) ResetRunningContext() {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.errors = nil
	m.mock.ResetRunningContext()
}

// EndRunningContext finalizes the running context and returns all accumulated errors.
func (m *ServiceMock) EndRunningContext(intermediate bool) []error {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	errs := m.errors
	errs = append(errs, m.mock.EndRunningContext(intermediate)...)
	for i := range errs {
		errs[i] = colorize.NewEntityError("mock %s", m.ServiceName).SetSubError(errs[i])
	}
	if intermediate {
		m.errors = nil
	}
	return errs
}

// SetCheckers configures the request checkers used to validate incoming HTTP requests.
func (m *ServiceMock) SetCheckers(checkers []CheckerInterface) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.checkers = checkers
}
