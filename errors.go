package openpayments

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type OpenPaymentsClientError struct {
	Description      string
	Status           int
	Code             string
	ValidationErrors []string
	Details          map[string]any
	Method           string
	URL              string
}

func (e *OpenPaymentsClientError) Error() string {
	if e.Status != 0 {
		return fmt.Sprintf("Error making Open Payments %s request to %s: %d %s",
			e.Method, e.URL, e.Status, e.Description)
	}
	return fmt.Sprintf("Error making Open Payments %s request to %s: %s",
		e.Method, e.URL, e.Description)
}

func newClientErrorFromResponse(req *http.Request, resp *http.Response) *OpenPaymentsClientError {
	bodyBytes, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))

	e := &OpenPaymentsClientError{
		Status: resp.StatusCode,
		Method: req.Method,
		URL:    req.URL.String(),
	}

	var envelope struct {
		Error struct {
			Description string         `json:"description"`
			Code        string         `json:"code"`
			Details     map[string]any `json:"details"`
		} `json:"error"`
	}

	if err := json.Unmarshal(bodyBytes, &envelope); err == nil && envelope.Error.Description != "" {
		e.Description = envelope.Error.Description
		e.Code = envelope.Error.Code
		e.Details = envelope.Error.Details
	} else if len(bodyBytes) > 0 {
		e.Description = string(bodyBytes)
	} else {
		e.Description = resp.Status
	}

	return e
}
