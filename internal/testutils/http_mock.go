package testutils

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
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
		json.NewEncoder(w).Encode(response) // #nosec G104
	}))

	return server
}
