package testutils

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"

	openpayments "github.com/interledger/open-payments-go"
)

func Mock(method string, path string, status int, response any) *httptest.Server {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != method {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		if r.URL.Path != path {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		w.WriteHeader(status)
		json.NewEncoder(w).Encode(response)
	}))

	return server
}

// Spy can be improved in the future if we want to spy and mock at the same time.
// This would require by accepting a captureResponse argument as well.
func Spy(status int, capture **http.Request) openpayments.RequestDoer {
	spy := func(req *http.Request) (*http.Response, error) {
		*capture = req
		return &http.Response{
			StatusCode: status,
			Body:       io.NopCloser(strings.NewReader("")),
		}, nil
	}
	return spy
}
