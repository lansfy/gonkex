package main

import (
	"net/http/httptest"
	"os"
	"testing"
	"text/template"

	"github.com/lansfy/gonkex/mocks"
	"github.com/lansfy/gonkex/runner"
)

// defaultFunc returns the default value if the provided value is empty (nil or empty string)
func defaultFunc(defaultValue interface{}, value interface{}) interface{} {
	if value == nil {
		return defaultValue
	}
	if str, ok := value.(string); ok && str == "" {
		return defaultValue
	}
	return value
}

func TestProxy(t *testing.T) {
	m := mocks.NewNop("backend")
	if err := m.Start(); err != nil {
		t.Fatal(err)
	}
	defer m.Shutdown()

	os.Setenv("BACKEND_ADDR", m.Service("backend").ServerAddr())
	initServer()
	srv := httptest.NewServer(nil)

	funcMap := template.FuncMap{
		"default": defaultFunc,
	}

	runner.RunWithTesting(t, srv.URL, &runner.RunWithTestingOpts{
		TestsDir:           "cases",
		Mocks:              m,
		TemplateReplyFuncs: funcMap,
	})
}
