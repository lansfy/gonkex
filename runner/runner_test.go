package runner

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestDontFollowRedirects(t *testing.T) {
	srv := testServerRedirect()
	defer srv.Close()

	RunWithTesting(t, srv.URL, &RunWithTestingOpts{
		TestsDir: "testdata/dont-follow-redirects",
	})
}

func testServerRedirect() *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/redirect-url", http.StatusFound)
	}))
}
