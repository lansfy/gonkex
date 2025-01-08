package mocks

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
)

var _ http.ResponseWriter = (*wrapResponseWriter)(nil)

func createResponseWriterProxy(rw http.ResponseWriter) *wrapResponseWriter {
	return &wrapResponseWriter{
		statusCode: http.StatusOK,
		headers:    make(http.Header),
		body:       &bytes.Buffer{},
		writer:     rw,
	}
}

type wrapResponseWriter struct {
	drop       bool
	statusCode int
	headers    http.Header
	body       *bytes.Buffer
	writer     http.ResponseWriter
}

func (w *wrapResponseWriter) Header() http.Header {
	return w.headers
}

func (w *wrapResponseWriter) WriteHeader(statusCode int) {
	w.statusCode = statusCode
}

func (w *wrapResponseWriter) Write(b []byte) (int, error) {
	return w.body.Write(b)
}

func (w *wrapResponseWriter) Flush() error {
	if w.drop {
		return dropConnection(w.writer)
	}
	for key, values := range w.headers {
		for _, value := range values {
			w.writer.Header().Add(key, value)
		}
	}
	w.writer.WriteHeader(w.statusCode)
	_, err := w.writer.Write(w.body.Bytes())
	return err
}

func (w *wrapResponseWriter) CreateHttpResponse() *http.Response {
	if w.drop {
		return nil
	}
	return &http.Response{
		Status:     http.StatusText(w.statusCode),
		StatusCode: w.statusCode,
		Header:     w.headers,
		Body:       io.NopCloser(bytes.NewReader(w.body.Bytes())),
	}
}

func dropConnection(w http.ResponseWriter) error {
	hj, ok := w.(http.Hijacker)
	if !ok {
		return fmt.Errorf("gonkex internal error: drop request: webserver does not support hijacking")
	}
	conn, _, err := hj.Hijack()
	if err != nil {
		return fmt.Errorf("gonkex internal error: connection hijacking: %w", err)
	}
	conn.Close()
	return nil
}
