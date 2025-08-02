package mocks

import (
	"bufio"
	"errors"
	"net"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

// mockNoHijackWriter does NOT implement http.Hijacker
type mockNoHijackWriter struct {
	http.ResponseWriter
}

// mockHijackError returns an error on Hijack
type mockHijackError struct {
	http.ResponseWriter
}

func (m *mockHijackError) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	return nil, nil, errors.New("hijack error")
}

// mockHijackSuccess returns a dummy Conn and closes it
type mockConn struct {
	net.Conn
	closed bool
}

func (m *mockConn) Close() error {
	m.closed = true
	return nil
}

type mockHijackSuccess struct {
	http.ResponseWriter
	conn *mockConn
}

func (m *mockHijackSuccess) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	return m.conn, nil, nil
}

func Test_dropConnection(t *testing.T) {
	t.Run("not hijacker", func(t *testing.T) {
		err := dropConnection(&mockNoHijackWriter{})
		require.EqualError(t, err, "gonkex internal error: drop request: webserver does not support hijacking")
	})

	t.Run("hijack error", func(t *testing.T) {
		err := dropConnection(&mockHijackError{})
		require.EqualError(t, err, "gonkex internal error: connection hijacking: hijack error")
	})

	t.Run("hijack success", func(t *testing.T) {
		conn := &mockConn{}
		err := dropConnection(&mockHijackSuccess{conn: conn})
		require.NoError(t, err)
		require.True(t, conn.closed)
	})
}
