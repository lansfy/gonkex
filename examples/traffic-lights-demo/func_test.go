package main

import (
	"net/http/httptest"
	"testing"

	"github.com/lansfy/gonkex/runner"
)

func Test_API(t *testing.T) {
	initServer()

	srv := httptest.NewServer(nil)

	runner.RunWithTesting(t, srv.URL, &runner.RunWithTestingOpts{
		TestsDir: "cases",
	})
}
