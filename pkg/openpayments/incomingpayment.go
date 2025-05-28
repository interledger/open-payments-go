package openpayments

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

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
func (ip *IncomingPaymentService) Get(ctx context.Context, url string, accessToken string) (rs.IncomingPayment, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return rs.IncomingPayment{}, err
	}

	resp, err := ip.DoSigned(req)
	if err != nil {
		return rs.IncomingPayment{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return rs.IncomingPayment{}, fmt.Errorf("failed to get incoming payment: %s", resp.Status)
	}

	var incomingPayment rs.IncomingPayment
	err = json.NewDecoder(resp.Body).Decode(&incomingPayment)
	if err != nil {
		return rs.IncomingPayment{}, fmt.Errorf("failed to decode response body: %s", err)
	}

	return incomingPayment, nil
}

// TODO: implement
func (ip *IncomingPaymentService) List(ctx context.Context, url string, accessToken string, args ListArgs) ([]rs.IncomingPayment, error) {
	return nil, nil
}
