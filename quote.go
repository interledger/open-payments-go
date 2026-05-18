package openpayments

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	rs "github.com/interledger/open-payments-go/generated/resourceserver"
)

type QuoteService struct {
	DoUnsigned RequestDoer
	DoSigned   RequestDoer
}

type QuoteGetParams struct {
	URL         string // The full URL of the quote resource.
	AccessToken string
}

type QuoteCreateParams struct {
	BaseURL     string // The base URL for creating a quote (e.g., wallet address URL).
	AccessToken string
	// TODO: cant use rs.CreateQuoteJSONBody (unexported `union`). is there a better workaround
	// for this payload than any?
	Payload any
}

func (qs *QuoteService) Get(ctx context.Context, params QuoteGetParams) (rs.Quote, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, params.URL, nil)
	if err != nil {
		return rs.Quote{}, err
	}

	req.Header.Set("Authorization", fmt.Sprintf("GNAP %s", params.AccessToken))

	resp, err := qs.DoSigned(req)
	if err != nil {
		return rs.Quote{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return rs.Quote{}, newClientErrorFromResponse(req, resp)
	}

	var quote rs.Quote
	err = json.NewDecoder(resp.Body).Decode(&quote)
	if err != nil {
		return rs.Quote{}, fmt.Errorf("failed to decode response body: %w", err)
	}

	return quote, nil
}

func (qs *QuoteService) Create(ctx context.Context, params QuoteCreateParams) (rs.Quote, error) {
	payloadBytes, err := json.Marshal(params.Payload)
	if err != nil {
		return rs.Quote{}, fmt.Errorf("failed to marshal payload: %w", err)
	}

	fullURL, err := url.JoinPath(params.BaseURL, "quotes")
	if err != nil {
		return rs.Quote{}, fmt.Errorf("failed to construct URL: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, fullURL, bytes.NewBuffer(payloadBytes))
	if err != nil {
		return rs.Quote{}, err
	}

	req.Header.Set("Authorization", fmt.Sprintf("GNAP %s", params.AccessToken))

	resp, err := qs.DoSigned(req)
	if err != nil {
		return rs.Quote{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return rs.Quote{}, newClientErrorFromResponse(req, resp)
	}

	var quote rs.Quote
	err = json.NewDecoder(resp.Body).Decode(&quote)
	if err != nil {
		return rs.Quote{}, fmt.Errorf("failed to decode response body: %s", err)
	}

	return quote, nil
}
