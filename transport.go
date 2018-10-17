package registryfrontend

import (
	"net/http"
	"strings"
)

type basicTransport struct {
	Transport http.RoundTripper
	URL       string
	Username  string
	Password  string
}

func (t *basicTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if strings.HasPrefix(req.URL.String(), t.URL) {
		if t.Username != "" || t.Password != "" {
			req.SetBasicAuth(t.Username, t.Password)
		}
	}
	resp, err := t.Transport.RoundTrip(req)
	return resp, err
}

func NewBasicAuthRoundTripper(url, user, password string) http.RoundTripper {
	return &basicTransport{
		Transport: http.DefaultTransport,
		URL:       url,
		Username:  user,
		Password:  password,
	}
}
