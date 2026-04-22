package openpayments

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// OpError represents an error from the Open Payments API.
// It preserves the HTTP status code and any structured error
// details from the response body, enabling callers to
// programmatically handle different error scenarios.
type OpError struct {
	// StatusCode is the HTTP status code (e.g., 401, 403, 404).
	StatusCode int

	// Status is the HTTP status text (e.g., "401 Unauthorized").
	Status string

	// Message is a human-readable error description.
	Message string

	// Details contains any structured error information from the
	// response body, if the server returned JSON.
	Details map[string]interface{}
}

func (e *OpError) Error() string {
	if e.Message != "" {
		return fmt.Sprintf("%s: %s", e.Status, e.Message)
	}
	return e.Status
}

// IsStatus returns true if the error has the given HTTP status code.
func (e *OpError) IsStatus(code int) bool {
	return e.StatusCode == code
}

// IsUnauthorized returns true if the error is a 401 Unauthorized response.
func (e *OpError) IsUnauthorized() bool {
	return e.StatusCode == http.StatusUnauthorized
}

// IsForbidden returns true if the error is a 403 Forbidden response.
func (e *OpError) IsForbidden() bool {
	return e.StatusCode == http.StatusForbidden
}

// IsNotFound returns true if the error is a 404 Not Found response.
func (e *OpError) IsNotFound() bool {
	return e.StatusCode == http.StatusNotFound
}

// newOpError creates an OpError from an HTTP response.
// It reads and attempts to parse the response body as JSON
// for structured error details.
func newOpError(resp *http.Response, operation string) *OpError {
	opErr := &OpError{
		StatusCode: resp.StatusCode,
		Status:     resp.Status,
		Message:    fmt.Sprintf("failed to %s", operation),
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil || len(body) == 0 {
		return opErr
	}

	// Try to parse the response body as JSON for structured details
	var details map[string]interface{}
	if json.Unmarshal(body, &details) == nil {
		opErr.Details = details
		// If the server sent a description, use it as the message
		if desc, ok := details["description"].(string); ok {
			opErr.Message = desc
		} else if msg, ok := details["message"].(string); ok {
			opErr.Message = msg
		}
	} else {
		// Body wasn't JSON — include it as-is in the message
		opErr.Message = fmt.Sprintf("failed to %s: %s", operation, string(body))
	}

	return opErr
}
