package openpayments

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	rs "github.com/interledger/open-payments-go-sdk/pkg/generated/resourceserver"
)

type IncomingPaymentService struct {
	DoUnsigned RequestDoer
	DoSigned RequestDoer
}

type PublicIncomingPaymentService struct {
	DoUnsigned RequestDoer
}

type Pagination struct {
	First string
	Last string
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
	BaseURL       string     // The base URL for the incoming payments collection.
	AccessToken   string
	WalletAddress string
	Pagination    Pagination
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

// TODO: protect against bad pagination args (ie not first AND last). Perhaps do in centralized
// way after many List are implemented
// - NewIncomingPaymentListParams that returns error with bad page args?
func (ip *IncomingPaymentService) List(ctx context.Context, params IncomingPaymentListParams) ([]rs.IncomingPaymentWithMethods, error) {
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

	var incomingPayment []rs.IncomingPaymentWithMethods
	err = json.NewDecoder(resp.Body).Decode(&incomingPayment)
	if err != nil {
		return nil, fmt.Errorf("failed to decode response body: %s", err)
	}

	return incomingPayment, nil
}
