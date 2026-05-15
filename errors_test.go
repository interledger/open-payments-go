package openpayments

import (
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"
)

func TestNewClientErrorFromResponse_JSONBody(t *testing.T) {
	body := `{"error":{"description":"invalid grant","code":"invalid_client","details":{"field":"client_id"}}}`
	req, _ := http.NewRequest(http.MethodPost, "https://auth.example.com/grant", nil)
	resp := &http.Response{
		StatusCode: 401,
		Status:     "401 Unauthorized",
		Body:       io.NopCloser(strings.NewReader(body)),
	}

	err := newClientErrorFromResponse(req, resp)

	if err.Status != 401 {
		t.Errorf("expected status 401, got %d", err.Status)
	}
	if err.Description != "invalid grant" {
		t.Errorf("expected description 'invalid grant', got %q", err.Description)
	}
	if err.Code != "invalid_client" {
		t.Errorf("expected code 'invalid_client', got %q", err.Code)
	}
	if err.Details["field"] != "client_id" {
		t.Errorf("expected details.field 'client_id', got %v", err.Details["field"])
	}
	if err.Method != http.MethodPost {
		t.Errorf("expected method POST, got %q", err.Method)
	}
	if err.URL != "https://auth.example.com/grant" {
		t.Errorf("expected URL 'https://auth.example.com/grant', got %q", err.URL)
	}
}

func TestNewClientErrorFromResponse_NonJSONBody(t *testing.T) {
	req, _ := http.NewRequest(http.MethodGet, "https://rs.example.com/payments", nil)
	resp := &http.Response{
		StatusCode: 500,
		Status:     "500 Internal Server Error",
		Body:       io.NopCloser(strings.NewReader("something went wrong")),
	}

	err := newClientErrorFromResponse(req, resp)

	if err.Status != 500 {
		t.Errorf("expected status 500, got %d", err.Status)
	}
	if err.Description != "something went wrong" {
		t.Errorf("expected raw body as description, got %q", err.Description)
	}
	if err.Code != "" {
		t.Errorf("expected empty code, got %q", err.Code)
	}
}

func TestNewClientErrorFromResponse_EmptyBody(t *testing.T) {
	req, _ := http.NewRequest(http.MethodDelete, "https://auth.example.com/token/123", nil)
	resp := &http.Response{
		StatusCode: 403,
		Status:     "403 Forbidden",
		Body:       io.NopCloser(strings.NewReader("")),
	}

	err := newClientErrorFromResponse(req, resp)

	if err.Status != 403 {
		t.Errorf("expected status 403, got %d", err.Status)
	}
	if err.Description != "403 Forbidden" {
		t.Errorf("expected status text as description, got %q", err.Description)
	}
}

func TestOpenPaymentsClientError_ErrorString(t *testing.T) {
	err := &OpenPaymentsClientError{
		Description: "not found",
		Status:      404,
		Method:      "GET",
		URL:         "https://example.com/payments/123",
	}

	expected := "Error making Open Payments GET request to https://example.com/payments/123: 404 not found"
	if err.Error() != expected {
		t.Errorf("expected %q, got %q", expected, err.Error())
	}
}

func TestOpenPaymentsClientError_ErrorStringNoStatus(t *testing.T) {
	err := &OpenPaymentsClientError{
		Description: "validation failed",
		Method:      "POST",
		URL:         "https://example.com/payments",
	}

	expected := "Error making Open Payments POST request to https://example.com/payments: validation failed"
	if err.Error() != expected {
		t.Errorf("expected %q, got %q", expected, err.Error())
	}
}

func TestOpenPaymentsClientError_ErrorsAs(t *testing.T) {
	origErr := newClientErrorFromResponse(
		func() *http.Request {
			r, _ := http.NewRequest(http.MethodGet, "https://example.com/test", nil)
			return r
		}(),
		&http.Response{
			StatusCode: 401,
			Status:     "401 Unauthorized",
			Body:       io.NopCloser(strings.NewReader(`{"error":{"description":"unauthorized","code":"invalid_token"}}`)),
		},
	)

	var wrapped error = origErr

	var clientErr *OpenPaymentsClientError
	if !errors.As(wrapped, &clientErr) {
		t.Fatal("expected errors.As to match *OpenPaymentsClientError")
	}
	if clientErr.Status != 401 {
		t.Errorf("expected status 401, got %d", clientErr.Status)
	}
	if clientErr.Code != "invalid_token" {
		t.Errorf("expected code 'invalid_token', got %q", clientErr.Code)
	}
}
