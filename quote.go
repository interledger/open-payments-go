package openpayments

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	rs "github.com/interledger/open-payments-go-sdk/generated/resourceserver"
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

func (ip *QuoteService) Get(ctx context.Context, params QuoteGetParams) (rs.Quote, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, params.URL, nil)
	if err != nil {
		return rs.Quote{}, err
	}

	req.Header.Set("Authorization", fmt.Sprintf("GNAP %s", params.AccessToken))

	resp, err := ip.DoSigned(req)
	if err != nil {
		return rs.Quote{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return rs.Quote{}, fmt.Errorf("failed to get quote: %s", resp.Status)
	}

	var quote rs.Quote
	err = json.NewDecoder(resp.Body).Decode(&quote)
	if err != nil {
		return rs.Quote{}, fmt.Errorf("failed to decode response body: %s", err)
	}

	return quote, nil
}

func (ip *QuoteService) Create(ctx context.Context, params QuoteCreateParams) (rs.Quote, error) {
	payloadBytes, err := json.Marshal(params.Payload)
	if err != nil {
		return rs.Quote{}, fmt.Errorf("failed to marshal payload: %w", err)
	}

	baseURL := strings.TrimRight(params.BaseURL, "/")
	fullURL := fmt.Sprintf("%s/quotes", baseURL)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, fullURL, bytes.NewBuffer(payloadBytes))
	if err != nil {
		return rs.Quote{}, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("GNAP %s", params.AccessToken))

	resp, err := ip.DoSigned(req)
	if err != nil {
		return rs.Quote{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return rs.Quote{}, fmt.Errorf("failed to create quote: %s", resp.Status)
	}

	var quote rs.Quote
	err = json.NewDecoder(resp.Body).Decode(&quote)
	if err != nil {
		return rs.Quote{}, fmt.Errorf("failed to decode response body: %s", err)
	}

	return quote, nil
}
