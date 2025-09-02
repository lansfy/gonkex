package runner

import (
	"crypto/tls"
	"net/http"
	"net/url"
)

func newClient(proxyURL *url.URL) *http.Client {
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true}, //nolint:gosec // Client is only used for testing.
		Proxy:           http.ProxyURL(proxyURL),
	}

	return &http.Client{
		Transport: transport,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
}
