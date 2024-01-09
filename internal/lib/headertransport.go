package lib

import (
	"net/http"
)

type HeaderTransport struct {
	Base http.RoundTripper
}

func (t *HeaderTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Header.Set("Accept", "application/json")
	return t.Base.RoundTrip(req)
}