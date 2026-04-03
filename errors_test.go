package openpayments

import (
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"
)

func newTestResponse(statusCode int, body string) *http.Response {
	return &http.Response{
		StatusCode: statusCode,
		Status:     http.StatusText(statusCode),
		Body:       io.NopCloser(strings.NewReader(body)),
	}
}

func TestOpErrorBasic(t *testing.T) {
	resp := newTestResponse(404, "")
	err := newOpError(resp, "get incoming payment")

	if err.StatusCode != 404 {
		t.Errorf("expected status code 404, got %d", err.StatusCode)
	}
	if !err.IsNotFound() {
		t.Error("expected IsNotFound to be true")
	}
	if err.IsUnauthorized() {
		t.Error("expected IsUnauthorized to be false")
	}
}

func TestOpErrorWithJSONBody(t *testing.T) {
	body := `{"code":"invalid_token","description":"The access token is expired"}`
	resp := newTestResponse(401, body)
	err := newOpError(resp, "get outgoing payment")

	if err.StatusCode != 401 {
		t.Errorf("expected status code 401, got %d", err.StatusCode)
	}
	if !err.IsUnauthorized() {
		t.Error("expected IsUnauthorized to be true")
	}
	// Should use the "description" field from the JSON body
	if err.Message != "The access token is expired" {
		t.Errorf("expected message from JSON body, got: %s", err.Message)
	}
	if err.Details == nil {
		t.Fatal("expected details to be populated")
	}
	if err.Details["code"] != "invalid_token" {
		t.Errorf("expected code=invalid_token, got: %v", err.Details["code"])
	}
}

func TestOpErrorWithPlainTextBody(t *testing.T) {
	resp := newTestResponse(500, "Internal Server Error occurred")
	err := newOpError(resp, "create quote")

	if err.StatusCode != 500 {
		t.Errorf("expected status code 500, got %d", err.StatusCode)
	}
	if err.Details != nil {
		t.Error("expected no details for plain text body")
	}
	if !strings.Contains(err.Message, "Internal Server Error occurred") {
		t.Errorf("expected message to contain body text, got: %s", err.Message)
	}
}

func TestOpErrorImplementsError(t *testing.T) {
	resp := newTestResponse(403, "")
	err := newOpError(resp, "list payments")

	// Should be usable as a regular error
	var opErr *OpError
	if !errors.As(err, &opErr) {
		t.Error("expected errors.As to match *OpError")
	}
	if !opErr.IsForbidden() {
		t.Error("expected IsForbidden to be true")
	}
}

func TestOpErrorIsStatus(t *testing.T) {
	resp := newTestResponse(429, "")
	err := newOpError(resp, "create payment")

	if !err.IsStatus(429) {
		t.Error("expected IsStatus(429) to be true")
	}
	if err.IsStatus(200) {
		t.Error("expected IsStatus(200) to be false")
	}
}
