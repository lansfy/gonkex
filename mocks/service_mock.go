package mocks

import (
	"context"
	"net"
	"net/http"
	"sync"

	"github.com/lansfy/gonkex/colorize"
)

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

func (m *ServiceMock) StartServer() error {
	return m.StartServerWithAddr("localhost:0") // loopback, random port
}

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

func (m *ServiceMock) ShutdownServer(ctx context.Context) error {
	err := m.server.Shutdown(ctx)
	m.listener = nil
	m.server = nil
	return err
}

func (m *ServiceMock) ServerAddr() string {
	if m.listener == nil {
		panic("mock server " + m.ServiceName + " is not started")
	}
	return m.listener.Addr().String()
}

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

func (m *ServiceMock) SetDefinition(newDefinition *Definition) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.mock = newDefinition
}

func (m *ServiceMock) ResetDefinition() {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.mock = m.defaultDefinition
}

func (m *ServiceMock) ResetRunningContext() {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.errors = nil
	m.mock.ResetRunningContext()
}

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

func (m *ServiceMock) SetCheckers(checkers []CheckerInterface) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.checkers = checkers
}
