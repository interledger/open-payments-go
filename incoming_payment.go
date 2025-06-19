package openpayments

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	rs "github.com/interledger/open-payments-go-sdk/generated/resourceserver"
)

type IncomingPaymentService struct {
	DoUnsigned RequestDoer
	DoSigned   RequestDoer
}

type PublicIncomingPaymentService struct {
	DoUnsigned RequestDoer
}

type Pagination struct {
	First  string
	Last   string
	Cursor string
}

type IncomingPaymentGetPublicParams struct {
	URL string // The full URL of the public incoming payment resource.
}

type IncomingPaymentGetParams struct {
	URL         string // The full URL of the incoming payment resource.
	AccessToken string
}

type IncomingPaymentListParams struct {
	BaseURL       string // The base URL for the incoming payments collection.
	AccessToken   string
	WalletAddress string
	Pagination    Pagination
}

type IncomingPaymentListResponse struct {
	Pagination rs.PageInfo                     `json:"pagination"`
	Result     []rs.IncomingPaymentWithMethods `json:"result"`
}

type IncomingPaymentCreateParams struct {
	BaseURL     string // The base URL for creating an incoming payment
	AccessToken string
	Payload     rs.CreateIncomingPaymentJSONBody
}

type IncomingPaymentCompleteParams struct {
	URL         string // The incoming payment url
	AccessToken string
}

func (ip *IncomingPaymentService) GetPublic(ctx context.Context, params IncomingPaymentGetPublicParams) (rs.PublicIncomingPayment, error) {
	return getPublic(ctx, ip.DoUnsigned, params.URL)
}

func (pp *PublicIncomingPaymentService) GetPublic(ctx context.Context, params IncomingPaymentGetPublicParams) (rs.PublicIncomingPayment, error) {
	return getPublic(ctx, pp.DoUnsigned, params.URL)
}

func getPublic(ctx context.Context, doUnsigned RequestDoer, url string) (rs.PublicIncomingPayment, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return rs.PublicIncomingPayment{}, err
	}

	resp, err := doUnsigned(req)
	if err != nil {
		return rs.PublicIncomingPayment{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return rs.PublicIncomingPayment{}, fmt.Errorf("failed to get incoming payment: %s", resp.Status)
	}

	var incomingPayment rs.PublicIncomingPayment
	err = json.NewDecoder(resp.Body).Decode(&incomingPayment)
	if err != nil {
		return rs.PublicIncomingPayment{}, fmt.Errorf("failed to decode response body: %s", err)
	}

	return incomingPayment, nil
}

func (ip *IncomingPaymentService) Get(ctx context.Context, params IncomingPaymentGetParams) (rs.IncomingPaymentWithMethods, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, params.URL, nil)
	if err != nil {
		return rs.IncomingPaymentWithMethods{}, err
	}

	req.Header.Set("Authorization", fmt.Sprintf("GNAP %s", params.AccessToken))

	resp, err := ip.DoSigned(req)
	if err != nil {
		return rs.IncomingPaymentWithMethods{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return rs.IncomingPaymentWithMethods{}, fmt.Errorf("failed to get incoming payment: %s", resp.Status)
	}

	var incomingPayment rs.IncomingPaymentWithMethods
	err = json.NewDecoder(resp.Body).Decode(&incomingPayment)
	if err != nil {
		return rs.IncomingPaymentWithMethods{}, fmt.Errorf("failed to decode response body: %s", err)
	}

	return incomingPayment, nil
}

func (ip *IncomingPaymentService) List(ctx context.Context, params IncomingPaymentListParams) (*IncomingPaymentListResponse, error) {
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

	base := strings.TrimRight(params.BaseURL, "/")
	fullURL := fmt.Sprintf("%s/incoming-payments?%s", base, query.Encode())

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fullURL, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", fmt.Sprintf("GNAP %s", params.AccessToken))

	resp, err := ip.DoSigned(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get incoming payment: %s", resp.Status)
	}

	var listResponse IncomingPaymentListResponse
	err = json.NewDecoder(resp.Body).Decode(&listResponse)
	if err != nil {
		return nil, fmt.Errorf("failed to decode response body: %s", err)
	}

	return &listResponse, nil
}

// TODO: should Create handle adding /incoming-paymnets or nah? php and rust do, ts doesnt
func (ip *IncomingPaymentService) Create(ctx context.Context, params IncomingPaymentCreateParams) (rs.IncomingPaymentWithMethods, error) {
	payloadBytes, err := json.Marshal(params.Payload)
	if err != nil {
		return rs.IncomingPaymentWithMethods{}, fmt.Errorf("failed to marshal payload: %w", err)
	}

	baseURL := strings.TrimRight(params.BaseURL, "/")
	fullURL := fmt.Sprintf("%s/incoming-payments", baseURL)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, fullURL, bytes.NewBuffer(payloadBytes))
	if err != nil {
		return rs.IncomingPaymentWithMethods{}, err
	}

	// TODO: do this more centrally? in DoSigned when content-legnth > 0?
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("GNAP %s", params.AccessToken))

	resp, err := ip.DoSigned(req)
	if err != nil {
		return rs.IncomingPaymentWithMethods{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return rs.IncomingPaymentWithMethods{}, fmt.Errorf("failed to create incoming payment: %s", resp.Status)
	}

	var incomingPayment rs.IncomingPaymentWithMethods
	err = json.NewDecoder(resp.Body).Decode(&incomingPayment)
	if err != nil {
		return rs.IncomingPaymentWithMethods{}, fmt.Errorf("failed to decode response body: %s", err)
	}

	return incomingPayment, nil
}

func (ip *IncomingPaymentService) Complete(ctx context.Context, params IncomingPaymentCompleteParams) (rs.IncomingPaymentWithMethods, error) {
	fullURL := fmt.Sprintf("%s/complete", strings.TrimRight(params.URL, "/"))

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, fullURL, nil)
	if err != nil {
		return rs.IncomingPaymentWithMethods{}, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("GNAP %s", params.AccessToken))

	resp, err := ip.DoSigned(req)
	if err != nil {
		return rs.IncomingPaymentWithMethods{}, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return rs.IncomingPaymentWithMethods{}, fmt.Errorf("failed to complete incoming payment: %s", resp.Status)
	}

	// TODO: add the validation like TS? php/rust. Not sure if we should on principle.
	// https://github.com/interledger/open-payments/blob/main/packages/open-payments/src/client/incoming-payment.ts#L217
	// TS client errors if returned payment is neither complete or if amounts are wrong.

	var incomingPayment rs.IncomingPaymentWithMethods
	err = json.NewDecoder(resp.Body).Decode(&incomingPayment)
	if err != nil {
		return rs.IncomingPaymentWithMethods{}, fmt.Errorf("failed to decode response body: %w", err)
	}

	return incomingPayment, nil
}
