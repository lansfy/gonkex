package main

import (
	"net/http/httptest"
	"testing"

	"github.com/lansfy/gonkex/mocks"
	"github.com/lansfy/gonkex/runner"
)

func TestProxy(t *testing.T) {
	m := mocks.NewNop("backend")
	if err := m.Start(); err != nil {
		t.Fatal(err)
	}
	defer m.Shutdown()

	initServer()
	srv := httptest.NewServer(nil)

	runner.RunWithTesting(t, srv.URL, &runner.RunWithTestingOpts{
		TestsDir: "cases",
		Mocks:    m,
	})
}
