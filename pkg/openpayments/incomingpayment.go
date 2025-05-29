package openpayments

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

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

type ListArgs struct {
	WalletAddress string
	Pagination Pagination
}

func (ip *IncomingPaymentService) GetPublic(ctx context.Context, url string) (rs.PublicIncomingPayment, error) {
	return getPublic(ctx, ip.DoUnsigned, url)
}

func (pp *PublicIncomingPaymentService) GetPublic(ctx context.Context, url string) (rs.PublicIncomingPayment, error) {
	return getPublic(ctx, pp.DoUnsigned, url)
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

// TODO: verify working
func (ip *IncomingPaymentService) Get(ctx context.Context, url string, accessToken string) (rs.IncomingPaymentWithMethods, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return rs.IncomingPaymentWithMethods{}, err
	}

	req.Header.Set("Authorization", fmt.Sprintf("GNAP %s", accessToken))

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

// TODO: implement
func (ip *IncomingPaymentService) List(ctx context.Context, baseUrl string, accessToken string, args ListArgs) ([]rs.IncomingPaymentWithMethods, error) {
	return nil, nil

	// Not working yet
	query := url.Values{}
	query.Set("wallet-address", args.WalletAddress)
	if args.Pagination.First != "" {
		query.Set("first", args.Pagination.First)
	}
	if args.Pagination.Last != "" {
		query.Set("last", args.Pagination.Last)
	}
	if args.Pagination.Cursor != "" {
		query.Set("cursor", args.Pagination.Cursor)
	}

	fullURL := fmt.Sprintf("%s?%s", baseUrl, query.Encode())

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fullURL, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", fmt.Sprintf("GNAP %s", accessToken))

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
