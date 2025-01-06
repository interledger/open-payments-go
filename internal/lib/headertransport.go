package lib

import (
	"net/http"
)

type HeaderTransport struct {
	Base        http.RoundTripper
	SigningFunc func(*http.Request) error
}

func (t *HeaderTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Header.Set("Accept", "application/json")
	if t.SigningFunc != nil {
		if err := t.SigningFunc(req); err != nil {
			return nil, err
		}
	}
	return t.Base.RoundTrip(req)
}