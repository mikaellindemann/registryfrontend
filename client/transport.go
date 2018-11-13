package client

import (
	"net/http"
	"strings"
)

type basicAuthRoundTripper struct {
	url      string
	user     string
	password string
	inner    http.RoundTripper
}

func (t *basicAuthRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	if strings.HasPrefix(req.URL.String(), t.url) {
		if t.user != "" || t.password != "" {
			req.SetBasicAuth(t.user, t.password)
		}
	}
	resp, err := t.inner.RoundTrip(req)
	return resp, err
}

type baseUrlRoundTripper struct {
	scheme string
	host   string
	inner  http.RoundTripper
}

func (b *baseUrlRoundTripper) RoundTrip(r *http.Request) (*http.Response, error) {
	r.URL.Scheme = b.scheme
	r.URL.Host = b.host

	return b.inner.RoundTrip(r)
}
