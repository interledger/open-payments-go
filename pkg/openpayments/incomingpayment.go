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

func (ip *IncomingPaymentService) GetPublic(ctx context.Context, url string) (rs.PublicIncomingPayment, error) {
	return getPublic(ctx, ip.DoUnsigned, url)
}

func (pp *PublicIncomingPaymentService) GetPublic(ctx context.Context, url string) (rs.PublicIncomingPayment, error) {
	return getPublic(ctx, pp.DoUnsigned, url)
}
func getPublic(ctx context.Context, do RequestDoer, url string) (rs.PublicIncomingPayment, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return rs.PublicIncomingPayment{}, err
	}

	resp, err := do(req)
	if err != nil {
		return rs.PublicIncomingPayment{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return rs.PublicIncomingPayment{}, fmt.Errorf("failed to get incoming payment: %s", resp.Status)
	}

	var incomingPaymentResponse rs.PublicIncomingPayment
	err = json.NewDecoder(resp.Body).Decode(&incomingPaymentResponse)
	if err != nil {
		return rs.PublicIncomingPayment{}, fmt.Errorf("failed to decode response body: %s", err)
	}

	return incomingPaymentResponse, nil
}