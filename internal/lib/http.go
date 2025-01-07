package lib

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type HeaderTransport struct {
	Base http.RoundTripper
}

func (t *HeaderTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Header.Set("Accept", "application/json")
	return t.Base.RoundTrip(req)
}

func FetchAndDecode[T any](httpClient *http.Client, url string) (T, error) {
	resp, err := httpClient.Get(url)
	if err != nil {
		var zero T
		return zero, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var zero T
		return zero, fmt.Errorf("unexpected HTTP status: %s", resp.Status)
	}

	var result T
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		var zero T
		return zero, fmt.Errorf("failed to decode response body: %w", err)
	}

	return result, nil
}