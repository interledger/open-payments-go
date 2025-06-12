package openpayments

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	rs "github.com/interledger/open-payments-go-sdk/pkg/generated/resourceserver"
)

type OutgoingPaymentService struct {
	DoSigned RequestDoer
}

type OutgoingPaymentGetParams struct {
	URL         string // The full URL of the outgoing payment resource.
	AccessToken string
}

type OutgoingPaymentListParams struct {
	BaseURL       string     // The base URL for the outgoing payments collection.
	AccessToken   string
	WalletAddress string
	Pagination    Pagination
}

type OutgoingPaymentListResponse struct {
	Pagination rs.PageInfo                `json:"pagination"`
	Result     []rs.OutgoingPayment       `json:"result"`
}

type OutgoingPaymentCreateParams struct {
	BaseURL     string // The base URL for creating an outgoing payment
	AccessToken string
	Payload     rs.CreateOutgoingPaymentJSONBody
}

func (op *OutgoingPaymentService) Get(ctx context.Context, params OutgoingPaymentGetParams) (rs.OutgoingPayment, error) {
	if params.URL == "" || params.AccessToken == "" {
		return rs.OutgoingPayment{}, fmt.Errorf("missing required url or access token")
	}
	if !strings.Contains(params.URL, "outgoing-payments/") {
		return rs.OutgoingPayment{}, fmt.Errorf("invalid outgoing payment URL")
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, params.URL, nil)
	if err != nil {
		return rs.OutgoingPayment{}, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("GNAP %s", params.AccessToken))

	resp, err := op.DoSigned(req)
	if err != nil {
		return rs.OutgoingPayment{}, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return rs.OutgoingPayment{}, fmt.Errorf("failed to get outgoing payment: %s", resp.Status)
	}

	var outgoingPayment rs.OutgoingPayment
	err = json.NewDecoder(resp.Body).Decode(&outgoingPayment)
	if err != nil {
		return rs.OutgoingPayment{}, fmt.Errorf("failed to decode response body: %w", err)
	}

	return outgoingPayment, nil
}

func (op *OutgoingPaymentService) List(ctx context.Context, params OutgoingPaymentListParams) (*OutgoingPaymentListResponse, error) {
	if params.BaseURL == "" || params.AccessToken == "" || params.WalletAddress == "" {
		return nil, fmt.Errorf("missing required base url, access token, or wallet address")
	}

	query := url.Values{}
	query.Set("wallet-address", params.WalletAddress)
	if params.Pagination.First != "" {
		query.Set("first", params.Pagination.First)
	}
	if params.Pagination.Last != "" {
		query.Set("last", params.Pagination.Last)
	}
	if params.Pagination.Cursor != "" {
		query.Set("cursor", params.Pagination.Cursor)
	}

	// Ensure single trailing slash on base URL
	base := strings.TrimRight(params.BaseURL, "/")
	fullURL := fmt.Sprintf("%s/outgoing-payments?%s", base, query.Encode())

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fullURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("GNAP %s", params.AccessToken))

	resp, err := op.DoSigned(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to list outgoing payments: %s", resp.Status)
	}

	var listResponse OutgoingPaymentListResponse
	err = json.NewDecoder(resp.Body).Decode(&listResponse)
	if err != nil {
		return nil, fmt.Errorf("failed to decode response body: %w", err)
	}

	return &listResponse, nil
}

func (op *OutgoingPaymentService) Create(ctx context.Context, params OutgoingPaymentCreateParams) (rs.OutgoingPayment, error) {
	if params.BaseURL == "" || params.AccessToken == "" {
		return rs.OutgoingPayment{}, fmt.Errorf("missing required base url or access token")
	}

	payloadBytes, err := json.Marshal(params.Payload)
	if err != nil {
		return rs.OutgoingPayment{}, fmt.Errorf("failed to marshal payload: %w", err)
	}

	// ensure trailing slash
	baseURL := strings.TrimRight(params.BaseURL, "/")
	fullURL := fmt.Sprintf("%s/outgoing-payments", baseURL)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, fullURL, bytes.NewBuffer(payloadBytes))
	if err != nil {
		return rs.OutgoingPayment{}, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("GNAP %s", params.AccessToken))

	resp, err := op.DoSigned(req)
	if err != nil {
		return rs.OutgoingPayment{}, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return rs.OutgoingPayment{}, fmt.Errorf("failed to create outgoing payment: %s", resp.Status)
	}

	var outgoingPayment rs.OutgoingPayment
	err = json.NewDecoder(resp.Body).Decode(&outgoingPayment)
	if err != nil {
		return rs.OutgoingPayment{}, fmt.Errorf("failed to decode response body: %w", err)
	}

	return outgoingPayment, nil
} 